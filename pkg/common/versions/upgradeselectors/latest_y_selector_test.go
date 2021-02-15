package upgradeselectors

import (
	"testing"

	"github.com/Masterminds/semver"
	"github.com/openshift/osde2e/pkg/common/spi"
)

func TestLatestYVersionSelectVersion(t *testing.T) {
	tests := []struct {
		name            string
		installVersion  *spi.Version
		versions        *spi.VersionList
		expectedVersion *spi.Version
		expectedErr     bool
	}{
		{
			name:           "get latest y version",
			installVersion: spi.NewVersionBuilder().Version(semver.MustParse("4.2.0")).Build(),
			versions: spi.NewVersionListBuilder().
				AvailableVersions([]*spi.Version{
					spi.NewVersionBuilder().Version(semver.MustParse("4.1.0")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.2.0")).AvailableUpgrades(map[*semver.Version]bool{
						semver.MustParse("4.2.2"): true,
						semver.MustParse("4.2.4"): true,
						semver.MustParse("4.3.0"): true,
					}).Build(),
					spi.NewVersionBuilder().Default(true).Version(semver.MustParse("4.3.0")).AvailableUpgrades(map[*semver.Version]bool{
						semver.MustParse("4.3.2"): true,
						semver.MustParse("4.3.4"): true,
						semver.MustParse("4.4.0"): true,
					}).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.4.0")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.5.0")).Build(),
				}).
				Build(),
			expectedVersion: spi.NewVersionBuilder().Version(semver.MustParse("4.3.0")).Build(),
		},
		{
			name:           "get latest y nightly version",
			installVersion: spi.NewVersionBuilder().Version(semver.MustParse("4.2.0")).Build(),
			versions: spi.NewVersionListBuilder().
				AvailableVersions([]*spi.Version{
					spi.NewVersionBuilder().Version(semver.MustParse("4.1.0")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.2.0")).AvailableUpgrades(map[*semver.Version]bool{
						semver.MustParse("4.3.0-0.nightly-2020-10-31-200727"): true,
						semver.MustParse("4.3.0-0.nightly-2020-11-01-200727"): true,
						semver.MustParse("4.3.0-0.nightly-2020-10-31-200727"): true,
						semver.MustParse("4.3.0-0.nightly-2020-11-02-200727"): true,
						semver.MustParse("4.3.0-0.nightly-2020-11-02-423456"): true,
					}).Build(),
					spi.NewVersionBuilder().Default(true).Version(semver.MustParse("4.3.0")).AvailableUpgrades(map[*semver.Version]bool{
						semver.MustParse("4.3.2"): true,
						semver.MustParse("4.3.4"): true,
						semver.MustParse("4.4.0"): true,
					}).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.4.0")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.5.0")).Build(),
				}).
				Build(),
			expectedVersion: spi.NewVersionBuilder().Version(semver.MustParse("4.3.0-0.nightly-2020-11-02-423456")).Build(),
		},
		{
			name:           "get latest y version out of order",
			installVersion: spi.NewVersionBuilder().Version(semver.MustParse("4.2.0")).Build(),
			versions: spi.NewVersionListBuilder().
				AvailableVersions([]*spi.Version{
					spi.NewVersionBuilder().Version(semver.MustParse("4.4.0")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.2.0")).AvailableUpgrades(map[*semver.Version]bool{
						semver.MustParse("4.2.2"): true,
						semver.MustParse("4.2.4"): true,
						semver.MustParse("4.3.0"): true,
					}).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.5.0")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.1.0")).Build(),
					spi.NewVersionBuilder().Default(true).Version(semver.MustParse("4.3.0")).AvailableUpgrades(map[*semver.Version]bool{
						semver.MustParse("4.3.2"): true,
						semver.MustParse("4.3.4"): true,
						semver.MustParse("4.4.0"): true,
					}).Build(),
				}).
				Build(),
			expectedVersion: spi.NewVersionBuilder().Version(semver.MustParse("4.3.0")).Build(),
		},
		{
			name:           "get latest version from candidate channel",
			installVersion: spi.NewVersionBuilder().Version(semver.MustParse("4.2.0-candidate")).Build(),
			versions: spi.NewVersionListBuilder().
				AvailableVersions([]*spi.Version{
					spi.NewVersionBuilder().Version(semver.MustParse("4.4.0-candidate")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.2.0-candidate")).AvailableUpgrades(map[*semver.Version]bool{
						semver.MustParse("4.2.2-candidate"): true,
						semver.MustParse("4.2.4-candidate"): true,
						semver.MustParse("4.3.3-candidate"): true,
					}).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.5.0-candidate")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.1.0-candidate")).Build(),
					spi.NewVersionBuilder().Default(true).Version(semver.MustParse("4.3.0-candidate")).AvailableUpgrades(map[*semver.Version]bool{
						semver.MustParse("4.3.2-candidate"): true,
						semver.MustParse("4.3.4-candidate"): true,
						semver.MustParse("4.4.0-candidate"): true,
					}).Build(),
				}).
				Build(),
			expectedVersion: spi.NewVersionBuilder().Version(semver.MustParse("4.3.3-candidate")).Build(),
		},
		{
			name:           "get latest feature candidate from candidate channel",
			installVersion: spi.NewVersionBuilder().Version(semver.MustParse("4.2.0-candidate")).Build(),
			versions: spi.NewVersionListBuilder().
				AvailableVersions([]*spi.Version{
					spi.NewVersionBuilder().Version(semver.MustParse("4.4.0-candidate")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.2.0-candidate")).AvailableUpgrades(map[*semver.Version]bool{
						semver.MustParse("4.2.2-candidate"): true,
						semver.MustParse("4.2.4-candidate"): true,
						semver.MustParse("4.3.3-fc.0"):      true,
					}).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.5.0-candidate")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.1.0-candidate")).Build(),
					spi.NewVersionBuilder().Default(true).Version(semver.MustParse("4.3.0-candidate")).AvailableUpgrades(map[*semver.Version]bool{
						semver.MustParse("4.3.2-candidate"): true,
						semver.MustParse("4.3.4-candidate"): true,
						semver.MustParse("4.4.0-candidate"): true,
					}).Build(),
				}).
				Build(),
			expectedVersion: spi.NewVersionBuilder().Version(semver.MustParse("4.3.3-fc.0")).Build(),
		},
		{
			name:           "get latest release candidate from candidate channel",
			installVersion: spi.NewVersionBuilder().Version(semver.MustParse("4.2.0-candidate")).Build(),
			versions: spi.NewVersionListBuilder().
				AvailableVersions([]*spi.Version{
					spi.NewVersionBuilder().Version(semver.MustParse("4.4.0-candidate")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.2.0-candidate")).AvailableUpgrades(map[*semver.Version]bool{
						semver.MustParse("4.2.2-candidate"): true,
						semver.MustParse("4.2.4-candidate"): true,
						semver.MustParse("4.3.3-rc.0"):      true,
					}).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.5.0-candidate")).Build(),
					spi.NewVersionBuilder().Version(semver.MustParse("4.1.0-candidate")).Build(),
					spi.NewVersionBuilder().Default(true).Version(semver.MustParse("4.3.0-candidate")).AvailableUpgrades(map[*semver.Version]bool{
						semver.MustParse("4.3.2-candidate"): true,
						semver.MustParse("4.3.4-candidate"): true,
						semver.MustParse("4.4.0-candidate"): true,
					}).Build(),
				}).
				Build(),
			expectedVersion: spi.NewVersionBuilder().Version(semver.MustParse("4.3.3-rc.0")).Build(),
		},
		{
			name:           "no versions",
			installVersion: spi.NewVersionBuilder().Version(semver.MustParse("4.2.0")).Build(),
			versions: spi.NewVersionListBuilder().
				AvailableVersions([]*spi.Version{}).
				Build(),
			expectedVersion: nil,
			expectedErr:     true,
		},
	}

	for _, test := range tests {
		selector := latestYVersion{}
		selectedVersion, descriptor, err := selector.SelectVersion(test.installVersion, test.versions)

		if err != nil && !test.expectedErr {
			t.Errorf("test %s: error while selecting version: %v", test.name, err)
		}

		if err == nil {
			expectedDescriptor := "latest y version"
			if descriptor != expectedDescriptor {
				t.Errorf("test %s: descriptor (%s) does not match expected '%s'", test.name, descriptor, expectedDescriptor)
			}

			if (selectedVersion == nil || test.expectedVersion == nil) && selectedVersion != test.expectedVersion {
				t.Errorf("test %s: expected selected version (%v) to match expected version (%v) and one is nil", test.name, selectedVersion.Version().Original(), test.expectedVersion.Version().Original())
			} else if selectedVersion != nil && !selectedVersion.Version().Equal(test.expectedVersion.Version()) {
				t.Errorf("test %s: selected version (%v) does not match expected version (%v)", test.name, selectedVersion.Version().Original(), test.expectedVersion.Version().Original())
			}
		}
	}
}
