package config

import "testing"

func TestHasMatches(t *testing.T) {
	tests := []struct {
		Name                    string
		Regex                   string
		Text                    string
		IgnoreStrings           []string
		ExpectedNumberOfMatches int
	}{
		{
			Name:  "match found",
			Regex: "no such host",
			Text: `time="2020-05-09T16:12:33Z" level=debug msg="Still waiting for the Kubernetes API: Get https://api.osde2e-4-3.b9t7.p2.openshiftapps.com:6443/version?timeout=32s: dial tcp: lookup api.osde2e-4-3.b9t7.p2.openshiftapps.com on 10.121.6.76:53: no such host"
time="2020-05-09T16:12:33Z" level=debug msg="Still waiting for the Kubernetes API: Get https://api.osde2e-4-3.b9t7.p2.openshiftapps.com:6443/version?timeout=32s: dial tcp: lookup api.osde2e-4-3.b9t7.p2.openshiftapps.com on 10.121.6.76:53: no such host"
time="2020-05-09T16:12:33Z" level=debug msg="Still waiting for the Kubernetes API: Get https://api.osde2e-4-3.b9t7.p2.openshiftapps.com:6443/version?timeout=32s: dial tcp: lookup api.osde2e-4-3.b9t7.p2.openshiftapps.com on 10.121.6.76:53: no such host"`,
			IgnoreStrings:           []string{},
			ExpectedNumberOfMatches: 3,
		},
		{
			Name:  "no match found",
			Regex: "blah blah blah",
			Text: `time="2020-05-09T16:12:33Z" level=debug msg="Still waiting for the Kubernetes API: Get https://api.osde2e-4-3.b9t7.p2.openshiftapps.com:6443/version?timeout=32s: dial tcp: lookup api.osde2e-4-3.b9t7.p2.openshiftapps.com on 10.121.6.76:53: no such host"
time="2020-05-09T16:12:33Z" level=debug msg="Still waiting for the Kubernetes API: Get https://api.osde2e-4-3.b9t7.p2.openshiftapps.com:6443/version?timeout=32s: dial tcp: lookup api.osde2e-4-3.b9t7.p2.openshiftapps.com on 10.121.6.76:53: no such host"
time="2020-05-09T16:12:33Z" level=debug msg="Still waiting for the Kubernetes API: Get https://api.osde2e-4-3.b9t7.p2.openshiftapps.com:6443/version?timeout=32s: dial tcp: lookup api.osde2e-4-3.b9t7.p2.openshiftapps.com on 10.121.6.76:53: no such host"`,
			IgnoreStrings:           []string{},
			ExpectedNumberOfMatches: 0,
		},
		{
			Name:  "match found but ignored",
			Regex: "no such host",
			Text: `time="2020-05-09T16:12:33Z" level=debug msg="Still waiting for the Kubernetes API: Get https://api.osde2e-4-3.b9t7.p2.openshiftapps.com:6443/version?timeout=32s: dial tcp: lookup api.osde2e-4-3.b9t7.p2.openshiftapps.com on 10.121.6.76:53: no such host"
time="2020-05-09T16:12:33Z" level=debug msg="Still waiting for the Kubernetes API: Get https://api.osde2e-4-3.b9t7.p2.openshiftapps.com:6443/version?timeout=32s: dial tcp: lookup api.osde2e-4-3.b9t7.p2.openshiftapps.com on 10.121.6.76:53: no such host"
time="2020-05-09T16:12:33Z" level=debug msg="Still waiting for the Kubernetes API: Get https://api.osde2e-4-3.b9t7.p2.openshiftapps.com:6443/version?timeout=32s: dial tcp: lookup api.osde2e-4-3.b9t7.p2.openshiftapps.com on 10.121.6.76:53: no such host"`,
			IgnoreStrings:           []string{"Still waiting for the Kubernetes API"},
			ExpectedNumberOfMatches: 0,
		},
	}

	for _, test := range tests {
		logMetric := LogMetric{
			Name:                  "test metric",
			RegEx:                 test.Regex,
			IgnoreIfMatchContains: test.IgnoreStrings,
			HighThreshold:         1, // We're not testing thresholds as part of this test
			LowThreshold:          1,
		}
		numMatches := logMetric.HasMatches([]byte(test.Text))
		if numMatches != test.ExpectedNumberOfMatches {
			t.Errorf("test %s: number of matches (%d) did not match expected number of matches (%d)", test.Name, numMatches, test.ExpectedNumberOfMatches)
		}
	}
}
