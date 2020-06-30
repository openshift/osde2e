package metrics

import (
	"testing"
	"time"

	"github.com/Masterminds/semver"
	"github.com/prometheus/common/model"
)

func TestSampleToJUnitResult(t *testing.T) {
	tests := []struct {
		name           string
		sample         *model.SampleStream
		expectedOutput JUnitResult
	}{
		{
			name: "regular parse",
			sample: &model.SampleStream{
				Metric: map[model.LabelName]model.LabelValue{
					"install_version": "openshift-v4.1.0",
					"upgrade_version": "",
					"cloud_provider":  "test",
					"environment":     "prod",
					"suite":           "test-suite",
					"testname":        "test-name",
					"result":          "passed",
					"cluster_id":      "1234567",
					"phase":           "install",
					"job":             "test-job1",
					"job_id":          "9999",
				},
				Values: []model.SamplePair{
					{
						Timestamp: 1,
						Value:     10,
					},
				},
			},
			expectedOutput: JUnitResult{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: nil,
				CloudProvider:  "test",
				Environment:    "prod",
				Suite:          "test-suite",
				TestName:       "test-name",
				Result:         Passed,
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Phase:          Install,
				Duration:       10 * time.Second,
			},
		},
	}

	for _, test := range tests {
		jUnitResult, err := sampleToJUnitResult(test.sample)

		if err != nil {
			t.Errorf("test %s failed while converting the sample to JUnit result: %v", test.name, err)
		}

		if !jUnitResult.Equal(test.expectedOutput) {
			t.Errorf("test %s failed because the produced JUnit result %v does not match the expected output %v", test.name, jUnitResult, test.expectedOutput)
		}
	}
}
