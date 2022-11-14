package installselectors

import (
	"fmt"
	"testing"

	"github.com/Masterminds/semver"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/spi"
)

func TestDeltaReleaseFromDefaultVersionSelectVersion(t *testing.T) {
	tests := []struct {
		name            string
		versions        *spi.VersionList
		delta           int
		expectedVersion *semver.Version
		expectedErr     bool
	}{
		{
			name: "get +1 from default",
			versions: spi.NewVersionListBuilder().
				AvailableVersions([]*spi.Version{
					spi.NewVersionBuilder().Version(semver.MustParse("4.1.0")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.2.0")).Build(),
					spi.NewVersionBuilder().Default(true).Version(semver.MustParse("4.3.0")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.4.0")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.5.0")).Build(),
				}).
				Build(),
			delta:           1,
			expectedVersion: semver.MustParse("4.4.0"),
		},
		{
			name: "get -1 from default",
			versions: spi.NewVersionListBuilder().
				AvailableVersions([]*spi.Version{
					spi.NewVersionBuilder().Version(semver.MustParse("4.1.0")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.2.0")).Build(),
					spi.NewVersionBuilder().Default(true).Version(semver.MustParse("4.3.0")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.4.0")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.5.0")).Build(),
				}).
				Build(),
			delta:           -1,
			expectedVersion: semver.MustParse("4.2.0"),
		},
		{
			name: "get +5 from default, expect error",
			versions: spi.NewVersionListBuilder().
				AvailableVersions([]*spi.Version{
					spi.NewVersionBuilder().Version(semver.MustParse("4.1.0")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.2.0")).Build(),
					spi.NewVersionBuilder().Default(true).Version(semver.MustParse("4.3.0")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.4.0")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.5.0")).Build(),
				}).
				Build(),
			delta:           5,
			expectedVersion: nil,
			expectedErr:     true,
		},
		{
			name: "get -5 from default, expect error",
			versions: spi.NewVersionListBuilder().
				AvailableVersions([]*spi.Version{
					spi.NewVersionBuilder().Version(semver.MustParse("4.1.0")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.2.0")).Build(),
					spi.NewVersionBuilder().Default(true).Version(semver.MustParse("4.3.0")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.4.0")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.5.0")).Build(),
				}).
				Build(),
			delta:           -5,
			expectedVersion: nil,
			expectedErr:     true,
		},
	}

	for _, test := range tests {
		viper.Reset()

		viper.Set(config.Cluster.DeltaReleaseFromDefault, test.delta)

		selector := deltaReleaseFromDefault{}
		selectedVersion, descriptor, err := selector.SelectVersion(test.versions)

		if err != nil && !test.expectedErr {
			t.Errorf("test %s: error while selecting version: %v", test.name, err)
		}

		if err == nil {
			expectedDescriptor := fmt.Sprintf("version %d releases from the default", test.delta)
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
