package collectors

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sqs"
	awscommon "github.com/openshift/osde2e/pkg/common/aws"
	"github.com/openshift/osde2e/pkg/dashboard/models"
	"github.com/openshift/osde2e/pkg/dashboard/store"
	"gopkg.in/yaml.v3"
)

// s3Event is the top-level SQS message body for S3 event notifications.
type s3Event struct {
	Records []struct {
		S3 struct {
			Bucket struct{ Name string } `json:"bucket"`
			Object struct{ Key string }  `json:"object"`
		} `json:"s3"`
	} `json:"Records"`
}

// SQSConsumer polls an SQS queue for S3 ObjectCreated events and writes
// parsed JUnit results into the Store.
type SQSConsumer struct {
	sqsClient *sqs.SQS
	opCollect *DeliverableCollector
	store     *store.Store
	queueURL  string
	bucket    string
}

// NewSQSConsumer creates a new consumer.
func NewSQSConsumer(queueURL, bucket, region string, st *store.Store) (*SQSConsumer, error) {
	opCollect, err := NewDeliverableCollector(bucket, region, 0)
	if err != nil {
		return nil, fmt.Errorf("create deliverable collector: %w", err)
	}

	sess, err := awscommon.CcsAwsSession.GetSession()
	if err != nil {
		return nil, fmt.Errorf("get AWS session: %w", err)
	}

	return &SQSConsumer{
		sqsClient: sqs.New(sess),
		opCollect: opCollect,
		store:     st,
		queueURL:  queueURL,
		bucket:    bucket,
	}, nil
}

// Run starts a long-poll loop that processes messages until ctx is cancelled.
// Call in a goroutine: go consumer.Run(ctx)
func (c *SQSConsumer) Run(ctx context.Context) {
	log.Printf("SQS consumer: started, queue=%s", c.queueURL)
	for {
		select {
		case <-ctx.Done():
			log.Printf("SQS consumer: stopped")
			return
		default:
		}

		msgs, err := c.sqsClient.ReceiveMessageWithContext(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(c.queueURL),
			MaxNumberOfMessages: aws.Int64(10),
			WaitTimeSeconds:     aws.Int64(20), // long poll — blocks up to 20s if queue empty
			VisibilityTimeout:   aws.Int64(60),
		})
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("SQS consumer: receive error: %v — retrying in 10s", err)
			select {
			case <-time.After(10 * time.Second):
			case <-ctx.Done():
				return
			}
			continue
		}

		for _, msg := range msgs.Messages {
			if err := c.processMessage(aws.StringValue(msg.Body)); err != nil {
				log.Printf("SQS consumer: process error: %v", err)
				// Leave on queue — will become visible again after VisibilityTimeout.
				continue
			}
			_, _ = c.sqsClient.DeleteMessage(&sqs.DeleteMessageInput{
				QueueUrl:      aws.String(c.queueURL),
				ReceiptHandle: msg.ReceiptHandle,
			})
		}
	}
}

// processMessage parses one SQS message body (direct S3 event or SNS-wrapped).
func (c *SQSConsumer) processMessage(body string) error {
	// SNS wraps the S3 JSON event inside a "Message" string field.
	var wrapper struct{ Message string }
	raw := body
	if err := json.Unmarshal([]byte(body), &wrapper); err == nil && wrapper.Message != "" {
		raw = wrapper.Message
	}

	var event s3Event
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		return fmt.Errorf("unmarshal S3 event: %w", err)
	}

	for _, rec := range event.Records {
		if err := c.processKey(rec.S3.Bucket.Name, rec.S3.Object.Key); err != nil {
			log.Printf("SQS consumer: skip %s: %v", rec.S3.Object.Key, err)
		}
	}
	return nil
}

// processKey downloads, parses, and stores the result for a single S3 JUnit key.
// Expected key format: test-results/<component>/<date>/<job-id>/<file>.xml
func (c *SQSConsumer) processKey(bucket, key string) error {
	if !strings.HasSuffix(key, ".xml") || !strings.Contains(key, "junit") {
		return nil
	}

	parts := strings.SplitN(key, "/", 5)
	if len(parts) < 5 {
		return fmt.Errorf("unexpected key format: %s", key)
	}

	component := parts[1]
	dateStr := parts[2]
	jobID := parts[3]

	name, version, env := parseComponentPath(component)

	// Always read the log to get env + image tag — these paths are unversioned
	// so parseComponentPath returns "unknown" for both env and version.
	logEnv, logVersion := c.opCollect.fetchMetaFromLog(name, dateStr, jobID)
	if env == "unknown" && logEnv != "" {
		env = logEnv
	}
	if version == "unknown" && logVersion != "" {
		version = logVersion
	}

	suite, ts, err := c.opCollect.downloadAndParseJUnit(key)
	if err != nil {
		return fmt.Errorf("parse junit %s: %w", key, err)
	}

	status := suiteStatus(suite)

	s3Dir := strings.Join(parts[:4], "/")

	// Only fetch LLM analysis for failed runs — no point for passing ones.
	var llm *models.LLMAnalysis
	if status != "passed" {
		llm = c.fetchLLMAnalysis(bucket, s3Dir)
	}

	rec := store.RunRecord{
		Name: name,
		Env:          env,
		Version:      version,
		Status:       status,
		Passed:       suite.Tests - suite.Failures - suite.Errors - suite.Skipped,
		Failed:       suite.Failures + suite.Errors,
		Total:        suite.Tests,
		JobID:        jobID,
		Date:         dateStr,
		LastRun:      ts,
		LogURL:       c.opCollect.generatePresignedURL(s3Dir + "/test_output.log"),
		JUnitURL:     c.opCollect.generatePresignedURL(key),
		FailedTests:  extractFailedTests(suite),
		LLMAnalysis:  llm,
	}

	if err := c.store.UpsertRun(rec); err != nil {
		return fmt.Errorf("upsert: %w", err)
	}

	log.Printf("SQS consumer: stored %s %s %s → %s", name, version, env, status)
	return nil
}

// summaryYAML mirrors the relevant fields of summary.yaml produced by the LLM analysis job.
type summaryYAML struct {
	Response string `yaml:"response"`
	Status   string `yaml:"status"`
}

// llmResponse is the JSON embedded in the response field (may be wrapped in ```json ... ```).
type llmResponse struct {
	RootCause       string   `json:"root_cause"`
	Recommendations []string `json:"recommendations"`
}

// reJSONBlock strips ```json ... ``` markdown fences if present.
var reJSONBlock = regexp.MustCompile("(?s)```(?:json)?\\s*(\\{.*?\\})\\s*```")

// fetchLLMAnalysis looks for a summary.yaml under the job's S3 prefix and parses it.
// It tries both known path patterns:
//  1. test-results/<op>/<date>/<jobID>/llm-analysis/summary.yaml
//  2. test-results/<op>/<date>/<jobID>/install/*/llm-analysis/summary.yaml
func (c *SQSConsumer) fetchLLMAnalysis(bucket, s3Dir string) *models.LLMAnalysis {
	// Pattern 1: shallow path
	candidates := []string{
		s3Dir + "/llm-analysis/summary.yaml",
	}

	// Pattern 2: deep path — list install/* to find the e2e image subdirectory
	listOut, err := c.opCollect.s3Client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(s3Dir + "/install/"),
	})
	if err == nil {
		for _, obj := range listOut.Contents {
			key := aws.StringValue(obj.Key)
			if strings.HasSuffix(key, "/llm-analysis/summary.yaml") {
				candidates = append(candidates, key)
			}
		}
	}

	for _, key := range candidates {
		out, err := c.opCollect.s3Client.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			continue
		}
		data, err := io.ReadAll(out.Body)
		out.Body.Close()
		if err != nil {
			continue
		}

		var sy summaryYAML
		if err := yaml.Unmarshal(data, &sy); err != nil || sy.Response == "" {
			continue
		}

		// Extract the JSON — may be bare or wrapped in ```json ... ```
		raw := sy.Response
		if m := reJSONBlock.FindStringSubmatch(raw); len(m) == 2 {
			raw = m[1]
		}

		var resp llmResponse
		if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &resp); err != nil {
			log.Printf("fetchLLMAnalysis: parse response JSON from %s: %v", key, err)
			continue
		}

		if resp.RootCause == "" {
			continue
		}

		return &models.LLMAnalysis{
			RootCause:       resp.RootCause,
			Recommendations: resp.Recommendations,
		}
	}

	return nil
}

// Backfill scans all historical S3 objects and populates the store from scratch.
// Run once at first startup or when the DB is missing/corrupt.
func (c *SQSConsumer) Backfill() error {
	log.Printf("Backfill: scanning s3://%s/test-results/ ...", c.bucket)

	// Collect unique (component, date, jobID) → best junit key
	type runKey struct{ component, date, jobID string }
	seen := make(map[runKey]string)

	err := c.opCollect.s3Client.ListObjectsV2Pages(&s3.ListObjectsV2Input{
		Bucket: aws.String(c.bucket),
		Prefix: aws.String("test-results/"),
	}, func(page *s3.ListObjectsV2Output, _ bool) bool {
		for _, obj := range page.Contents {
			key := aws.StringValue(obj.Key)
			if !strings.HasSuffix(key, ".xml") || !strings.Contains(key, "junit") {
				continue
			}
			parts := strings.SplitN(key, "/", 5)
			if len(parts) < 5 {
				continue
			}
			rk := runKey{parts[1], parts[2], parts[3]}
			if _, exists := seen[rk]; !exists {
				seen[rk] = key
			}
		}
		return true
	})
	if err != nil {
		return fmt.Errorf("list S3: %w", err)
	}

	log.Printf("Backfill: %d unique runs found — downloading in parallel...", len(seen))

	// Fan out with the same worker pool size as the operator collector.
	type work struct {
		bucket, key string
	}
	jobs := make(chan work, len(seen))
	for _, key := range seen {
		jobs <- work{c.bucket, key}
	}
	close(jobs)

	type result struct{ err error }
	results := make(chan result, len(seen))

	workers := downloadWorkers
	if workers > len(seen) {
		workers = len(seen)
	}
	for i := 0; i < workers; i++ {
		go func() {
			for j := range jobs {
				err := c.processKey(j.bucket, j.key)
				results <- result{err}
			}
		}()
	}

	ok, failed := 0, 0
	for range seen {
		r := <-results
		if r.err != nil {
			failed++
		} else {
			ok++
		}
	}

	log.Printf("Backfill: complete. ok=%d failed=%d", ok, failed)
	return nil
}