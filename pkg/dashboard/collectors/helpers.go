package collectors

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

// suiteStatus returns "passed", "failed", or "error" based on a parsed JUnit suite.
func suiteStatus(suite *JUnitTestSuite) string {
	if suite.Failures > 0 {
		return "failed"
	}
	if suite.Errors > 0 {
		return "error"
	}
	return "passed"
}

// parseTimestamp parses a JUnit timestamp string, falling back to time.Now().
func parseTimestamp(ts string) time.Time {
	if t, err := time.Parse("2006-01-02T15:04:05", ts); err == nil {
		return t
	}
	if t, err := time.Parse(time.RFC3339, ts); err == nil {
		return t
	}
	return time.Now()
}

// presignURL creates a 7-day presigned URL for an S3 key using the given client and bucket.
func presignURL(client *s3.S3, bucket, key string) string {
	req, _ := client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	url, err := req.Presign(7 * 24 * time.Hour)
	if err != nil {
		return ""
	}
	return url
}
