package collectors

import (
	"fmt"
	"io"
	"log"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	awssession "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/openshift/osde2e/pkg/dashboard/models"
)

const downloadWorkers = 20

// versionRegex matches semver tags (v1.2.3) or short git SHAs (7-10 hex chars)
var versionRegex = regexp.MustCompile(`^(v\d+(\.\d+)*|[0-9a-f]{7,10})$`)

var knownEnvSuffixes = []string{"integration", "stage", "prod", "int"}

// DeliverableCollector scans S3 for operator test results grouped by name, version, and environment.
type DeliverableCollector struct {
	s3Client     *s3.S3
	bucket       string
	region       string
	lookbackDays int
}

// NewDeliverableCollector creates a new collector using the standard AWS credential chain
// (env vars → ~/.aws/credentials → IAM role), independent of the osde2e viper config.
func NewDeliverableCollector(bucket, region string, lookbackDays int) (*DeliverableCollector, error) {
	sess, err := awssession.NewSession(aws.NewConfig().WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	s3Client := s3.New(sess)

	if lookbackDays <= 0 {
		lookbackDays = 30
	}

	return &DeliverableCollector{
		s3Client:     s3Client,
		bucket:       bucket,
		region:       region,
		lookbackDays: lookbackDays,
	}, nil
}

// S3Client returns the underlying S3 client and bucket name, used by the server's S3 proxy handler.
func (c *DeliverableCollector) S3Client() (*s3.S3, string) { return c.s3Client, c.bucket }

// parseComponentPath splits an S3 component string into operator name, version, and environment.
func parseComponentPath(component string) (name, version, env string) {
	tokens := strings.Split(component, "-")

	env = "unknown"
	if len(tokens) > 0 {
		last := tokens[len(tokens)-1]
		for _, suffix := range knownEnvSuffixes {
			if strings.EqualFold(last, suffix) {
				env = strings.ToLower(last)
				tokens = tokens[:len(tokens)-1]
				break
			}
		}
	}

	version = "unknown"
	versionIdx := -1
	for i := len(tokens) - 1; i >= 0; i-- {
		if versionRegex.MatchString(tokens[i]) {
			version = tokens[i]
			versionIdx = i
			break
		}
	}

	if versionIdx > 0 {
		name = strings.Join(tokens[:versionIdx], "-")
	} else if versionIdx == 0 {
		name = "unknown"
	} else {
		name = strings.Join(tokens, "-")
	}

	if name == "" {
		name = "unknown"
	}

	return name, version, env
}

// candidate holds a JUnit key identified during listing, before downloading.
type candidate struct {
	key       string
	component string
	dateStr   string
	jobID     string
	modified  time.Time // S3 LastModified, used to pick the newest per group
}

// downloadResult is the outcome of fetching and parsing one candidate.
type downloadResult struct {
	name    string
	version string
	env     string
	jobID   string
	s3Dir   string
	key     string
	suite   *JUnitTestSuite
	ts      time.Time
}

// CollectDeliverables scans S3 for junit XML files within the lookback window,
// groups them by operator name + version, and returns the latest result per environment.
func (c *DeliverableCollector) CollectDeliverables() ([]models.DeliverableStatus, error) {
	cutoff := time.Now().UTC().AddDate(0, 0, -c.lookbackDays)

	// Phase 1: list all matching keys, deduplicate to newest per (name, version, env).
	// S3 listing is cheap; downloading is not. We only download one file per group.
	type groupKey struct{ name, version, env string }
	newestByGroup := make(map[groupKey]*candidate)

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(c.bucket),
		Prefix: aws.String("test-results/"),
	}

	err := c.s3Client.ListObjectsV2Pages(input, func(page *s3.ListObjectsV2Output, _ bool) bool {
		for _, obj := range page.Contents {
			key := aws.StringValue(obj.Key)
			if !strings.HasSuffix(key, ".xml") || !strings.Contains(key, "junit") {
				continue
			}

			// Format: test-results/<component>/<date>/<job-id>/<filename>
			parts := strings.SplitN(key, "/", 5)
			if len(parts) < 5 {
				continue
			}

			dateStr := parts[2]
			keyDate, err := time.Parse("2006-01-02", dateStr)
			if err != nil || keyDate.Before(cutoff) {
				continue
			}

			component := parts[1]
			name, version, env := parseComponentPath(component)
			gk := groupKey{name, version, env}

			modified := aws.TimeValue(obj.LastModified)
			existing, seen := newestByGroup[gk]
			if !seen || modified.After(existing.modified) {
				newestByGroup[gk] = &candidate{
					key:       key,
					component: component,
					dateStr:   dateStr,
					jobID:     parts[3],
					modified:  modified,
				}
			}
		}
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list S3 objects: %w", err)
	}

	log.Printf("Deliverable collector: %d unique (name, version, env) groups to download", len(newestByGroup))

	// Phase 2: fan out downloads with a worker pool.
	candidates := make([]*candidate, 0, len(newestByGroup))
	groupKeys := make([]groupKey, 0, len(newestByGroup))
	for gk, cand := range newestByGroup {
		candidates = append(candidates, cand)
		groupKeys = append(groupKeys, gk)
	}

	results := make([]*downloadResult, len(candidates))
	var wg sync.WaitGroup
	sem := make(chan struct{}, downloadWorkers)

	for i, cand := range candidates {
		wg.Add(1)
		go func(i int, cand *candidate, gk groupKey) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			suite, ts, err := c.downloadAndParseJUnit(cand.key)
			if err != nil {
				log.Printf("Warning: skipping %s: %v", cand.key, err)
				return
			}

			parts := strings.SplitN(cand.key, "/", 5)
			s3Dir := strings.Join(parts[:4], "/")

			_, version, _ := parseComponentPath(cand.component)

			env := gk.env
			if env == "unknown" {
				if detected := c.fetchEnvFromLog(gk.name, parts[2], parts[3]); detected != "" {
					env = detected
				}
			}

			results[i] = &downloadResult{
				name:    gk.name,
				version: version,
				env:     env,
				jobID:   cand.jobID,
				s3Dir:   s3Dir,
				key:     cand.key,
				suite:   suite,
				ts:      ts,
			}
		}(i, cand, groupKeys[i])
	}
	wg.Wait()

	// Phase 3: build the index.
	index := make(map[string]*models.DeliverableStatus)
	for _, r := range results {
		if r == nil {
			continue
		}

		status := suiteStatus(r.suite)
		logURL := s3URL(c.bucket, r.s3Dir+"/test_output.log")
		junitURL := junitURL(c.bucket, r.key)

		indexKey := r.name
		op, exists := index[indexKey]
		if !exists {
			op = &models.DeliverableStatus{
				Name:    r.name,
				Version: r.version,
				Results: make(map[string]*models.EnvironmentResult),
			}
			index[indexKey] = op
		}

		failedTests := extractFailedTests(r.suite)

		op.Results[r.env] = &models.EnvironmentResult{
			Status:      status,
			Version:     r.version,
			Total:       r.suite.Tests,
			Passed:      r.suite.Tests - r.suite.Failures - r.suite.Errors - r.suite.Skipped,
			Failed:      r.suite.Failures,
			Skipped:     r.suite.Skipped,
			Errors:      r.suite.Errors,
			LastRun:     r.ts,
			JobID:       r.jobID,
			LogURL:      logURL,
			JUnitURL:    junitURL,
			FailedTests: failedTests,
		}

		if r.ts.After(op.LastUpdated) {
			op.LastUpdated = r.ts
		}
	}

	result := make([]models.DeliverableStatus, 0, len(index))
	for _, op := range index {
		result = append(result, *op)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Name != result[j].Name {
			return result[i].Name < result[j].Name
		}
		return result[i].Version < result[j].Version
	})

	log.Printf("Collected deliverable status for %d operator+version combinations", len(result))
	return result, nil
}

// adHocImageRegex extracts the image tag from AdHocTestImages in two formats:
//  1. "Successfully added property[AdHocTestImages] - quay.io/.../operator-e2e:c7fabd7"
//  2. "--properties AdHocTestImages:quay.io/.../operator-e2e:ec3ce7b" (rosa CLI args)
var adHocImageRegex = regexp.MustCompile(`AdHocTestImages[:\]] ?-? ?\S+:(\S+?)[ "]`)

// fetchMetaFromLog reads test_output.log and extracts both the environment
// ("Will load config <env>") and the image tag from the AdHocTestImages property line.
func (c *DeliverableCollector) fetchMetaFromLog(name, date, jobID string) (env, version string) {
	logKey := fmt.Sprintf("test-results/%s/%s/%s/test_output.log", name, date, jobID)
	output, err := c.s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(logKey),
	})
	if err != nil {
		return "", ""
	}
	defer output.Body.Close()

	buf := make([]byte, 16384) // 16KB — enough for the header lines
	n, _ := output.Body.Read(buf)
	content := string(buf[:n])

	for _, e := range []string{"stage", "prod", "int"} {
		if strings.Contains(content, "Will load config "+e) {
			env = e
			break
		}
	}

	if m := adHocImageRegex.FindStringSubmatch(content); len(m) == 2 {
		version = strings.TrimSpace(m[1])
	}

	return env, version
}

// fetchEnvFromLog is kept for callers that only need the environment.
func (c *DeliverableCollector) fetchEnvFromLog(name, date, jobID string) string {
	env, _ := c.fetchMetaFromLog(name, date, jobID)
	return env
}

// downloadAndParseJUnit fetches and parses a JUnit XML from S3.
func (c *DeliverableCollector) downloadAndParseJUnit(key string) (*JUnitTestSuite, time.Time, error) {
	output, err := c.s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("GetObject failed: %w", err)
	}
	defer output.Body.Close()

	data, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("read failed: %w", err)
	}

	suite, err := parseJUnitData(data)
	if err != nil {
		return nil, time.Time{}, err
	}

	ts := parseTimestamp(suite.Timestamp)

	return suite, ts, nil
}

// extractFailedTests pulls failed/errored test case names and messages from a suite.
func extractFailedTests(suite *JUnitTestSuite) []models.FailedTestCase {
	var out []models.FailedTestCase
	for _, tc := range suite.TestCases {
		var msg string
		if tc.Failure != nil {
			msg = *tc.Failure
		} else if tc.Error != nil {
			msg = *tc.Error
		} else {
			continue
		}
		if len(msg) > 600 {
			msg = msg[:600] + "…"
		}
		out = append(out, models.FailedTestCase{Name: tc.Name, Message: msg})
	}
	return out
}

// CollectPipelineHistory scans all S3 runs for a named operator and returns every
// (version, env, date, jobID) tuple found, sorted newest first.
func (c *DeliverableCollector) CollectPipelineHistory(operatorName string) (*models.PipelineHistory, error) {
	prefix := "test-results/"

	type runKey struct {
		component string
		dateStr   string
		jobID     string
	}
	seen := make(map[runKey]bool)
	var candidates []runKey

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(c.bucket),
		Prefix: aws.String(prefix),
	}

	err := c.s3Client.ListObjectsV2Pages(input, func(page *s3.ListObjectsV2Output, _ bool) bool {
		for _, obj := range page.Contents {
			key := aws.StringValue(obj.Key)
			if !strings.HasSuffix(key, ".xml") || !strings.Contains(key, "junit") {
				continue
			}
			parts := strings.SplitN(key, "/", 5)
			if len(parts) < 5 {
				continue
			}
			component := parts[1]
			name, _, _ := parseComponentPath(component)
			if name != operatorName {
				continue
			}
			rk := runKey{component: component, dateStr: parts[2], jobID: parts[3]}
			if !seen[rk] {
				seen[rk] = true
				candidates = append(candidates, rk)
			}
		}
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list S3 objects: %w", err)
	}

	// Fan-out: download each unique run in parallel
	type rawRun struct {
		version string
		env     string
		dateStr string
		jobID   string
		s3Dir   string
		key     string
		suite   *JUnitTestSuite
		ts      time.Time
	}

	rawRuns := make([]*rawRun, len(candidates))
	var wg sync.WaitGroup
	sem := make(chan struct{}, downloadWorkers)

	for i, rk := range candidates {
		wg.Add(1)
		go func(i int, rk runKey) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// Find the JUnit XML key for this run
			listOut, err := c.s3Client.ListObjectsV2(&s3.ListObjectsV2Input{
				Bucket: aws.String(c.bucket),
				Prefix: aws.String(fmt.Sprintf("test-results/%s/%s/%s/", rk.component, rk.dateStr, rk.jobID)),
			})
			if err != nil {
				return
			}
			var junitKey string
			for _, obj := range listOut.Contents {
				k := aws.StringValue(obj.Key)
				if strings.HasSuffix(k, ".xml") && strings.Contains(k, "junit") {
					junitKey = k
					break
				}
			}
			if junitKey == "" {
				return
			}

			suite, ts, err := c.downloadAndParseJUnit(junitKey)
			if err != nil {
				log.Printf("Warning: history skip %s: %v", junitKey, err)
				return
			}

			_, version, env := parseComponentPath(rk.component)
			if env == "unknown" {
				if detected := c.fetchEnvFromLog(operatorName, rk.dateStr, rk.jobID); detected != "" {
					env = detected
				}
			}

			s3Dir := fmt.Sprintf("test-results/%s/%s/%s", rk.component, rk.dateStr, rk.jobID)
			rawRuns[i] = &rawRun{
				version: version,
				env:     env,
				dateStr: rk.dateStr,
				jobID:   rk.jobID,
				s3Dir:   s3Dir,
				key:     junitKey,
				suite:   suite,
				ts:      ts,
			}
		}(i, rk)
	}
	wg.Wait()

	var runs []models.PipelineRun
	for _, r := range rawRuns {
		if r == nil {
			continue
		}
		runs = append(runs, models.PipelineRun{
			Version:  r.version,
			Env:      r.env,
			Status:   suiteStatus(r.suite),
			Date:     r.dateStr,
			JobID:    r.jobID,
			LastRun:  r.ts,
			LogURL:   s3URL(c.bucket, r.s3Dir+"/test_output.log"),
			JUnitURL: junitURL(c.bucket, r.key),
			Failed:   extractFailedTests(r.suite),
			Total:    r.suite.Tests,
			Passed:   r.suite.Tests - r.suite.Failures - r.suite.Errors - r.suite.Skipped,
		})
	}

	// Sort newest first
	sort.Slice(runs, func(i, j int) bool {
		return runs[i].LastRun.After(runs[j].LastRun)
	})

	return &models.PipelineHistory{
		Name: operatorName,
		Runs: runs,
	}, nil
}
