package installselectors

import (
	"testing"

	"github.com/Masterminds/semver"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/spi"
)

func TestSpecificImages(t *testing.T) {
	tests := []struct {
		name            string
		versions        *spi.VersionList
		expectedVersion *semver.Version
		image           string
		expectedErr     bool
	}{
		{
			name: "get specific nightly 1",
			versions: spi.NewVersionListBuilder().
				AvailableVersions([]*spi.Version{
					spi.NewVersionBuilder().Version(semver.MustParse("4.2.0")).Build(),
					spi.NewVersionBuilder().Default(true).Version(semver.MustParse("4.3.0")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.4.0")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.4.6")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.4.0-0.nightly-2020-11-06-072238-nightly")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.4.0-0.nightly-2020-11-06-130917-nightly")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.4.0-0.nightly-2020-11-07-020245-nightly")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.5.0")).Build(),
				}).
				Build(),
			image:           "registry.ci.openshift.org/ocp/release:4.4.0-0.nightly-2020-11-07-020245",
			expectedVersion: semver.MustParse("4.4.0-0.nightly-2020-11-07-020245-nightly"),
		},
		{
			name: "get specific nightly 2",
			versions: spi.NewVersionListBuilder().
				AvailableVersions([]*spi.Version{
					spi.NewVersionBuilder().Version(semver.MustParse("4.4.0")).Build(),
					spi.NewVersionBuilder().Default(true).Version(semver.MustParse("4.5.0")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.6.0-0.nightly-2020-11-06-232229-nightly")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.6.0-0.nightly-2020-11-07-024648-nightly")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.6.0-0.nightly-2020-11-07-035509-nightly")).Build(),
				}).
				Build(),
			image:           "registry.ci.openshift.org/ocp/release:4.6.0-0.nightly-2020-11-07-035509",
			expectedVersion: semver.MustParse("4.6.0-0.nightly-2020-11-07-035509-nightly"),
		},
		{
			name: "get specific nightly out of order",
			versions: spi.NewVersionListBuilder().
				AvailableVersions([]*spi.Version{
					spi.NewVersionBuilder().Version(semver.MustParse("4.4.0-0.nightly-2020-11-06-130917-nightly")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.7.0-0.nightly-2021-01-15-100321-nightly")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.4.0-0.nightly-2020-11-07-020245-nightly")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.4.0-0.nightly-2020-11-05-113252-nightly")).Build(),
					spi.NewVersionBuilder().Default(true).Version(semver.MustParse("4.4.3")).Build(),
				}).
				Build(),
			image:           "registry.ci.openshift.org/ocp/release:4.7.0-0.nightly-2021-01-15-100321",
			expectedVersion: semver.MustParse("4.7.0-0.nightly-2021-01-15-100321-nightly"),
		},
		{
			name: "no versions",
			versions: spi.NewVersionListBuilder().
				AvailableVersions([]*spi.Version{}).
				Build(),
			image:           "",
			expectedVersion: nil,
			expectedErr:     true,
		},
	}

	for _, test := range tests {
		viper.Set(config.Cluster.ReleaseImageLatest, test.image)
		selector := specificImage{}
		selectedVersion, descriptor, err := selector.SelectVersion(test.versions)

		if err != nil && !test.expectedErr {
			t.Errorf("test %s: error while selecting version: %v", test.name, err)
		}

		if err == nil {
			expectedDescriptor := "specific image"
			if descriptor != expectedDescriptor {
				t.Errorf("test %s: descriptor (%s) does not match expected '%s'", test.name, descriptor, expectedDescriptor)
			}

			if (selectedVersion == nil || test.expectedVersion == nil) && selectedVersion != test.expectedVersion {
				t.Errorf("test %s: expected selected version (%v) to match expected version (%v) and one is nil", test.name, selectedVersion, test.expectedVersion)
			} else if selectedVersion != nil && !selectedVersion.Equal(test.expectedVersion) {
				t.Errorf("test %s: selected version (%v) does not match expected version (%v)", test.name, selectedVersion, test.expectedVersion)
			}
		}
	}
	viper.Set(config.Cluster.ReleaseImageLatest, "")
}
