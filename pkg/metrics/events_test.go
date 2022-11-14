package metrics

import (
	"testing"

	"github.com/Masterminds/semver"
	"github.com/prometheus/common/model"
)

func TestSampleToEvent(t *testing.T) {
	tests := []struct {
		name           string
		sample         *model.SampleStream
		expectedOutput Event
	}{
		{
			name: "regular parse",
			sample: &model.SampleStream{
				Metric: map[model.LabelName]model.LabelValue{
					"install_version": "openshift-v4.1.0",
					"upgrade_version": "",
					"cloud_provider":  "test",
					"environment":     "prod",
					"event":           "test-event",
					"cluster_id":      "1234567",
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
			expectedOutput: Event{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: nil,
				CloudProvider:  "test",
				Environment:    "prod",
				Event:          "test-event",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Timestamp:      1,
			},
		},
	}

	for _, test := range tests {
		event, err := sampleToEvent(test.sample)
		if err != nil {
			t.Errorf("test %s failed while converting the sample to event: %v", test.name, err)
		}

		if !event.Equal(test.expectedOutput) {
			t.Errorf("test %s failed because the produced event %v does not match the expected output %v", test.name, event, test.expectedOutput)
		}
	}
}
