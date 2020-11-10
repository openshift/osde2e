package installselectors

import (
	"testing"

	"github.com/Masterminds/semver"
	"github.com/openshift/osde2e/pkg/common/spi"
)

func TestLatestZVersionSelectVersion(t *testing.T) {
	tests := []struct {
		name            string
		versions        *spi.VersionList
		expectedVersion *semver.Version
		expectedErr     bool
	}{
		{
			name: "get latest version",
			versions: spi.NewVersionListBuilder().
				AvailableVersions([]*spi.Version{
					spi.NewVersionBuilder().Version(semver.MustParse("4.3.1")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.3.2")).Build(),
					spi.NewVersionBuilder().Default(true).Version(semver.MustParse("4.3.3")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.3.4")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.3.5")).Build(),
				}).
				Build(),
			expectedVersion: semver.MustParse("4.3.5"),
		},
		{
			name: "get latest version out of order",
			versions: spi.NewVersionListBuilder().
				AvailableVersions([]*spi.Version{
					spi.NewVersionBuilder().Version(semver.MustParse("4.3.1")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.3.2")).Build(),
					spi.NewVersionBuilder().Default(true).Version(semver.MustParse("4.3.3")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.3.5")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.3.4")).Build(),
				}).
				Build(),
			expectedVersion: semver.MustParse("4.3.5"),
		},
		{
			name: "no valid target",
			versions: spi.NewVersionListBuilder().
				AvailableVersions([]*spi.Version{
					spi.NewVersionBuilder().Version(semver.MustParse("4.3.0")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.3.3")).Build(),
					spi.NewVersionBuilder().Default(true).Version(semver.MustParse("4.3.5")).Build(),
				}).
				Build(),
			expectedVersion: nil,
			expectedErr:     true,
		},
		{
			name: "no versions",
			versions: spi.NewVersionListBuilder().
				AvailableVersions([]*spi.Version{}).
				Build(),
			expectedVersion: nil,
			expectedErr:     true,
		},
	}

	for _, test := range tests {
		selector := latestZVersion{}
		selectedVersion, descriptor, err := selector.SelectVersion(test.versions)

		if err != nil && !test.expectedErr {
			t.Errorf("test %s: error while selecting version: %v", test.name, err)
		}

		if err == nil {
			expectedDescriptor := "latest z version from default"
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
}
