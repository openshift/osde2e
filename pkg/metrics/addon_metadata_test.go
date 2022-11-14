package metrics

import (
	"testing"

	"github.com/Masterminds/semver"
	"github.com/prometheus/common/model"
)

func TestSampleToAddonMetadata(t *testing.T) {
	tests := []struct {
		name           string
		sample         *model.SampleStream
		expectedOutput AddonMetadata
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
			expectedOutput: AddonMetadata{
				Metadata: Metadata{
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
				Phase: Install,
			},
		},
	}

	for _, test := range tests {
		addonMetadata, err := sampleToAddonMetadata(test.sample)
		if err != nil {
			t.Errorf("test %s failed while converting the sample to addon metadata: %v", test.name, err)
		}

		if !addonMetadata.Equal(test.expectedOutput) {
			t.Errorf("test %s failed because the produced addon metadata %v does not match the expected output %v", test.name, addonMetadata, test.expectedOutput)
		}
	}
}
