package metrics

import (
	"testing"

	"github.com/Masterminds/semver"
	"github.com/prometheus/common/model"
)

func TestSampleToMetadata(t *testing.T) {
	tests := []struct {
		name           string
		sample         *model.SampleStream
		expectedOutput Metadata
	}{
		{
			name: "regular parse",
			sample: &model.SampleStream{
				Metric: map[model.LabelName]model.LabelValue{
					"install_version": "openshift-v4.1.0",
					"upgrade_version": "",
					"cloud_provider":  "test",
					"environment":     "prod",
					"metadata_name":   "test-metadata",
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
			expectedOutput: Metadata{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: nil,
				CloudProvider:  "test",
				Environment:    "prod",
				MetadataName:   "test-metadata",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Value:          10,
				Timestamp:      1,
			},
		},
	}

	for _, test := range tests {
		metadata, err := sampleToMetadata(test.sample)
		if err != nil {
			t.Errorf("test %s failed while converting the sample to metadata: %v", test.name, err)
		}

		if !metadata.Equal(test.expectedOutput) {
			t.Errorf("test %s failed because the produced metadata %v does not match the expected output %v", test.name, metadata, test.expectedOutput)
		}
	}
}
