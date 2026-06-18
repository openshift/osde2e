package collectors

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	awscommon "github.com/openshift/osde2e/pkg/common/aws"
	"github.com/openshift/osde2e/pkg/dashboard/models"
)

// JUnitTestSuite represents a single <testsuite> element
type JUnitTestSuite struct {
	XMLName   xml.Name        `xml:"testsuite"`
	Name      string          `xml:"name,attr"`
	Tests     int             `xml:"tests,attr"`
	Failures  int             `xml:"failures,attr"`
	Errors    int             `xml:"errors,attr"`
	Skipped   int             `xml:"skipped,attr"`
	Time      float64         `xml:"time,attr"`
	Timestamp string          `xml:"timestamp,attr"`
	TestCases []JUnitTestCase `xml:"testcase"`
}

// jUnitTestSuites represents a <testsuites> wrapper (may contain multiple <testsuite> children)
type jUnitTestSuites struct {
	XMLName    xml.Name         `xml:"testsuites"`
	Tests      int              `xml:"tests,attr"`
	Failures   int              `xml:"failures,attr"`
	Errors     int              `xml:"errors,attr"`
	Time       float64          `xml:"time,attr"`
	TestSuites []JUnitTestSuite `xml:"testsuite"`
}

// JUnitTestCase represents a single test case
type JUnitTestCase struct {
	Name      string  `xml:"name,attr"`
	Classname string  `xml:"classname,attr"`
	Time      float64 `xml:"time,attr"`
	Failure   *string `xml:"failure,omitempty"`
	Error     *string `xml:"error,omitempty"`
	Skipped   *string `xml:"skipped,omitempty"`
}

// parseJUnitData parses raw JUnit XML bytes handling both <testsuite> and <testsuites> root elements.
// When the root is <testsuites>, suites are merged into a single JUnitTestSuite by summing counters
// and taking the timestamp from the first child suite.
func parseJUnitData(data []byte) (*JUnitTestSuite, error) {
	// Peek at the root element name
	type rootPeek struct {
		XMLName xml.Name
	}
	var peek rootPeek
	if err := xml.Unmarshal(data, &peek); err != nil {
		return nil, fmt.Errorf("failed to peek XML root: %w", err)
	}

	switch peek.XMLName.Local {
	case "testsuite":
		var suite JUnitTestSuite
		if err := xml.Unmarshal(data, &suite); err != nil {
			return nil, fmt.Errorf("failed to unmarshal <testsuite>: %w", err)
		}
		return &suite, nil

	case "testsuites":
		var suites jUnitTestSuites
		if err := xml.Unmarshal(data, &suites); err != nil {
			return nil, fmt.Errorf("failed to unmarshal <testsuites>: %w", err)
		}
		// Merge all child suites into one
		merged := &JUnitTestSuite{Name: "merged"}
		for _, s := range suites.TestSuites {
			merged.Tests += s.Tests
			merged.Failures += s.Failures
			merged.Errors += s.Errors
			merged.Skipped += s.Skipped
			merged.Time += s.Time
			merged.TestCases = append(merged.TestCases, s.TestCases...)
			if merged.Timestamp == "" && s.Timestamp != "" {
				merged.Timestamp = s.Timestamp
				merged.Name = s.Name
			}
		}
		return merged, nil

	default:
		return nil, fmt.Errorf("unexpected XML root element: <%s>", peek.XMLName.Local)
	}
}

// TestResultsCollector collects test results from S3
type TestResultsCollector struct {
	s3Client *s3.S3
	bucket   string
	region   string
}

// NewTestResultsCollector creates a new test results collector using existing AWS session
func NewTestResultsCollector(bucket, region string) (*TestResultsCollector, error) {
	sess, err := awscommon.CcsAwsSession.GetSession()
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS session: %w", err)
	}

	s3Client := s3.New(sess, aws.NewConfig().WithRegion(region))

	return &TestResultsCollector{
		s3Client: s3Client,
		bucket:   bucket,
		region:   region,
	}, nil
}

// CollectRecentTests retrieves recent test results from S3
func (c *TestResultsCollector) CollectRecentTests(maxResults int) ([]models.TestResult, error) {
	// List objects in the test-results/ prefix
	prefix := "test-results/"

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(c.bucket),
		Prefix: aws.String(prefix),
	}

	var allResults []models.TestResult
	resultsByJob := make(map[string]*models.TestResult)

	err := c.s3Client.ListObjectsV2Pages(input, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, obj := range page.Contents {
			key := aws.StringValue(obj.Key)

			// Skip if not a JUnit XML file
			if !strings.HasSuffix(key, ".xml") || !strings.Contains(key, "junit") {
				continue
			}

			// Parse the S3 key to extract metadata
			// Format: test-results/<component>/<date>/<job-id>/junit*.xml
			parts := strings.Split(key, "/")
			if len(parts) < 4 {
				continue
			}

			component := parts[1]
			date := parts[2]
			jobID := parts[3]

			jobKey := fmt.Sprintf("%s-%s-%s", component, date, jobID)

			// Only process if we haven't seen this job yet
			if _, exists := resultsByJob[jobKey]; !exists {
				result, err := c.parseJUnitXML(key, component, date, jobID)
				if err != nil {
					log.Printf("Warning: failed to parse %s: %v", key, err)
					continue
				}

				resultsByJob[jobKey] = result
			}
		}

		// Stop if we have enough results
		return len(resultsByJob) < maxResults
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list S3 objects: %w", err)
	}

	// Convert map to slice
	for _, result := range resultsByJob {
		allResults = append(allResults, *result)
	}

	// Sort by timestamp (most recent first)
	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].Timestamp.After(allResults[j].Timestamp)
	})

	// Limit results
	if len(allResults) > maxResults {
		allResults = allResults[:maxResults]
	}

	log.Printf("Collected %d test results from S3", len(allResults))
	return allResults, nil
}

// parseJUnitXML downloads and parses a JUnit XML file from S3
func (c *TestResultsCollector) parseJUnitXML(key, component, date, jobID string) (*models.TestResult, error) {
	// Download the file
	output, err := c.s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download %s: %w", key, err)
	}
	defer output.Body.Close()

	// Parse XML
	data, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", key, err)
	}

	suite, err := parseJUnitData(data)
	if err != nil {
		return nil, err
	}

	// Parse timestamp
	timestamp, err := time.Parse("2006-01-02T15:04:05", suite.Timestamp)
	if err != nil {
		timestamp, err = time.Parse(time.RFC3339, suite.Timestamp)
		if err != nil {
			timestamp = time.Now()
		}
	}

	// Determine status
	status := "passed"
	if suite.Failures > 0 {
		status = "failed"
	} else if suite.Errors > 0 {
		status = "error"
	}

	// Build per-test-case list
	testCases := make([]models.TestCase, 0, len(suite.TestCases))
	for _, tc := range suite.TestCases {
		tcStatus := "passed"
		var msg string
		if tc.Failure != nil {
			tcStatus = "failed"
			msg = *tc.Failure
		} else if tc.Error != nil {
			tcStatus = "error"
			msg = *tc.Error
		} else if tc.Skipped != nil {
			tcStatus = "skipped"
			msg = *tc.Skipped
		}
		// Trim long messages to 500 chars for the UI
		if len(msg) > 500 {
			msg = msg[:500] + "…"
		}
		testCases = append(testCases, models.TestCase{
			Name:     tc.Name,
			Duration: tc.Time,
			Status:   tcStatus,
			Message:  msg,
		})
	}

	s3Path := path.Dir(key)
	logURL := c.generatePresignedURL(path.Join(s3Path, "test_output.log"))
	junitURL := c.generatePresignedURL(key)

	return &models.TestResult{
		JobID:        jobID,
		JobName:      component,
		Component:    component,
		Date:         date,
		Status:       status,
		TotalTests:   suite.Tests,
		PassedTests:  suite.Tests - suite.Failures - suite.Errors - suite.Skipped,
		FailedTests:  suite.Failures,
		ErrorTests:   suite.Errors,
		SkippedTests: suite.Skipped,
		Duration:     suite.Time,
		S3Path:       s3Path,
		LogURL:       logURL,
		JUnitXMLURL:  junitURL,
		Timestamp:    timestamp,
		TestCases:    testCases,
	}, nil
}

// generatePresignedURL creates a presigned URL for an S3 object
func (c *TestResultsCollector) generatePresignedURL(key string) string {
	req, _ := c.s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})

	url, err := req.Presign(7 * 24 * time.Hour) // 7 days
	if err != nil {
		log.Printf("Warning: failed to generate presigned URL for %s: %v", key, err)
		return ""
	}

	return url
}

// GetTestResultByJobID retrieves detailed test results for a specific job
func (c *TestResultsCollector) GetTestResultByJobID(jobID string) (*models.TestResult, error) {
	// Search for the job in S3
	prefix := "test-results/"

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(c.bucket),
		Prefix: aws.String(prefix),
	}

	var result *models.TestResult

	err := c.s3Client.ListObjectsV2Pages(input, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, obj := range page.Contents {
			key := aws.StringValue(obj.Key)

			if !strings.Contains(key, jobID) || !strings.HasSuffix(key, ".xml") {
				continue
			}

			parts := strings.Split(key, "/")
			if len(parts) < 4 {
				continue
			}

			component := parts[1]
			date := parts[2]

			testResult, err := c.parseJUnitXML(key, component, date, jobID)
			if err != nil {
				log.Printf("Warning: failed to parse %s: %v", key, err)
				continue
			}

			result = testResult
			return false // Stop pagination
		}

		return true
	})

	if err != nil {
		return nil, fmt.Errorf("failed to search for job %s: %w", jobID, err)
	}

	if result == nil {
		return nil, fmt.Errorf("job %s not found", jobID)
	}

	return result, nil
}