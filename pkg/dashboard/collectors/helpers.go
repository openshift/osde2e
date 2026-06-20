package collectors

import (
	"net/url"
	"time"
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

// parseTimestamp parses a JUnit timestamp string, returning zero time on failure.
// Callers should apply their own fallback (e.g. S3 LastModified) for zero results.
func parseTimestamp(ts string) time.Time {
	if t, err := time.Parse("2006-01-02T15:04:05", ts); err == nil {
		return t
	}
	if t, err := time.Parse(time.RFC3339, ts); err == nil {
		return t
	}
	return time.Time{}
}

// s3URL returns a dashboard proxy URL that streams the S3 object through the server.
func s3URL(bucket, key string) string {
	return "/dashboard/s3?key=" + url.QueryEscape(key)
}

// junitURL returns a dashboard URL that fetches the JUnit XML from S3 and renders it as HTML.
func junitURL(bucket, key string) string {
	return "/dashboard/junit?key=" + url.QueryEscape(key)
}
