package installselectors

import (
	"testing"

	"github.com/Masterminds/semver"
	"github.com/openshift/osde2e/pkg/common/spi"
)

func TestDefaultVersionSelectVersion(t *testing.T) {
	tests := []struct {
		name            string
		versions        *spi.VersionList
		expectedVersion *semver.Version
	}{
		{
			name: "get default version",
			versions: spi.NewVersionListBuilder().
				AvailableVersions([]*spi.Version{
					spi.NewVersionBuilder().Version(semver.MustParse("4.1.0")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.2.0")).Build(),
					spi.NewVersionBuilder().Default(true).Version(semver.MustParse("4.3.0")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.4.0")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.5.0")).Build(),
				}).
				Build(),
			expectedVersion: semver.MustParse("4.3.0"),
		},
		{
			name: "get default version with override",
			versions: spi.NewVersionListBuilder().
				DefaultVersionOverride(semver.MustParse("4.6.0")).
				AvailableVersions([]*spi.Version{
					spi.NewVersionBuilder().Version(semver.MustParse("4.1.0")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.2.0")).Build(),
					spi.NewVersionBuilder().Default(true).Version(semver.MustParse("4.3.0")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.4.0")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.5.0")).Build(),
				}).
				Build(),
			expectedVersion: semver.MustParse("4.6.0"),
		},
		{
			name: "empty version list",
			versions: spi.NewVersionListBuilder().
				AvailableVersions([]*spi.Version{}).
				Build(),
			expectedVersion: nil,
		},
	}

	for _, test := range tests {
		selector := defaultVersion{}

		selectedVersion, descriptor, err := selector.SelectVersion(test.versions)
		if err != nil {
			t.Errorf("test %s: error while selecting version: %v", test.name, err)
		}

		if descriptor != "current default" {
			t.Errorf("test %s: descriptor (%s) does not match expected 'current default'", test.name, descriptor)
		}

		failIfVersionsNotEqual(t, test.name, selectedVersion, test.expectedVersion)
	}
}
