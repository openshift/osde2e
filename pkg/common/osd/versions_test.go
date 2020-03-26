package osd

import (
	"testing"

	"github.com/Masterminds/semver"
)

func TestVersion440Constraint(t *testing.T) {
	tests := []struct {
		Name           string
		Version        string
		ExpectedResult bool
	}{
		{
			Name:           "4.3.0 test",
			Version:        "4.3.0",
			ExpectedResult: false,
		},
		{
			Name:           "4.4.0 test",
			Version:        "4.4.0",
			ExpectedResult: true,
		},
		{
			Name:           "4.4.0-rc.0 test",
			Version:        "4.4.0-rc.0",
			ExpectedResult: true,
		},
	}

	for _, test := range tests {
		if Version440.Check(semver.MustParse(test.Version)) != test.ExpectedResult {
			t.Errorf("test %s did not produce the expected result: %t", test.Name, test.ExpectedResult)
		}
	}
}

func TestNextReleaseAfterGivenVersionFromVersionList(t *testing.T) {
	tests := []struct {
		Name                     string
		GivenVersion             *semver.Version
		VersionList              []string
		ReleasesFromGivenVersion int
		ExpectedVersion          string
	}{
		{
			Name:                     "no nightly, distance 1 (4.3.0)",
			GivenVersion:             semver.MustParse("4.3.0"),
			VersionList:              []string{"4.3.0", "4.3.1", "4.4.0", "4.4.2", "4.5.0", "4.5.5", "4.6.1"},
			ReleasesFromGivenVersion: 1,
			ExpectedVersion:          "4.4.2",
		},
		{
			Name:                     "no nightly, distance 1, given version doesn't exist in version list (4.4.1)",
			GivenVersion:             semver.MustParse("4.4.1"),
			VersionList:              []string{"4.3.0", "4.3.1", "4.4.0", "4.4.2", "4.5.0", "4.5.5", "4.6.1"},
			ReleasesFromGivenVersion: 1,
			ExpectedVersion:          "4.5.5",
		},
		{
			Name:                     "no nightly, distance 2 (4.3.0)",
			GivenVersion:             semver.MustParse("4.3.0"),
			VersionList:              []string{"4.3.0", "4.3.1", "4.4.0", "4.4.2", "4.5.0", "4.5.5", "4.6.1"},
			ReleasesFromGivenVersion: 2,
			ExpectedVersion:          "4.5.5",
		},
		{
			Name:                     "rc should be selected, distance 1 (4.3.0)",
			GivenVersion:             semver.MustParse("4.3.0"),
			VersionList:              []string{"4.3.0", "4.3.1", "4.4.0", "4.4.2", "4.4.3-rc.0", "4.5.0"},
			ReleasesFromGivenVersion: 1,
			ExpectedVersion:          "4.4.3-rc.0",
		},
		{
			Name:                     "rc should be skipped, distance 1 (4.3.0)",
			GivenVersion:             semver.MustParse("4.3.0"),
			VersionList:              []string{"4.3.0", "4.3.1", "4.4.0", "4.4.2", "4.4.3-rc.0", "4.4.3", "4.5.0"},
			ReleasesFromGivenVersion: 1,
			ExpectedVersion:          "4.4.3",
		},
		{
			Name:                     "nightly should be selected, distance 1 (4.3.0)",
			GivenVersion:             semver.MustParse("4.3.0"),
			VersionList:              []string{"4.3.0", "4.3.1", "4.4.0-0.nightly-1", "4.4.0", "4.4.2", "4.4.3-rc.0", "4.4.3", "4.5.0"},
			ReleasesFromGivenVersion: 1,
			ExpectedVersion:          "4.4.0-0.nightly-1",
		},
		{
			Name:                     "second nightly should be selected, distance 1 (4.3.0)",
			GivenVersion:             semver.MustParse("4.3.0"),
			VersionList:              []string{"4.3.0", "4.3.1", "4.4.0-0.nightly-1", "4.4.0-0.nightly-2", "4.4.0", "4.4.2", "4.4.3-rc.0", "4.4.3", "4.5.0"},
			ReleasesFromGivenVersion: 1,
			ExpectedVersion:          "4.4.0-0.nightly-2",
		},
	}

	for _, test := range tests {
		selectedVersion, err := nextReleaseAfterGivenVersionFromVersionList(test.GivenVersion, test.VersionList, test.ReleasesFromGivenVersion)

		if err != nil {
			t.Errorf("error selecting version from list: %v", err)
		}

		if selectedVersion != test.ExpectedVersion {
			t.Errorf("test %s did not produce the expected result: %s, got %s instead", test.Name, test.ExpectedVersion, selectedVersion)
		}
	}
}
