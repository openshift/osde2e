package aws

import (
	"testing"
)

func TestCreateS3URLs(t *testing.T) {
	tests := []struct {
		name          string
		bucket        string
		keys          []string
		expectedS3URL string
	}{
		{
			name:          "happy path",
			bucket:        "osde2e-metrics",
			keys:          []string{"incoming", "blah.prom"},
			expectedS3URL: "s3://osde2e-metrics/incoming/blah.prom",
		},
		{
			name:          "slashes in bucket",
			bucket:        "/osde2e-metrics/",
			keys:          []string{"incoming", "blah.prom"},
			expectedS3URL: "s3://osde2e-metrics/incoming/blah.prom",
		},
		{
			name:          "slashes in keys",
			bucket:        "osde2e-metrics",
			keys:          []string{"/incoming/", "/blah.prom"},
			expectedS3URL: "s3://osde2e-metrics/incoming/blah.prom",
		},
		{
			name:          "slashes in bucket and keys",
			bucket:        "/osde2e-metrics/",
			keys:          []string{"/incoming/", "/blah.prom"},
			expectedS3URL: "s3://osde2e-metrics/incoming/blah.prom",
		},
	}

	for _, test := range tests {
		createdS3URL := CreateS3URL(test.bucket, test.keys...)
		if createdS3URL != test.expectedS3URL {
			t.Errorf("error during test %s: created string %s does not much %s", test.name, createdS3URL, test.expectedS3URL)
		}
	}
}
