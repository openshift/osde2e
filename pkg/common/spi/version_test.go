package spi

import (
	"reflect"
	"testing"

	"github.com/Masterminds/semver"
)

func TestVersionBuilder(t *testing.T) {
	builtVersion := NewVersionBuilder().
		Version(semver.MustParse("1.2.3")).
		Default(true).
		Build()

	definedVersion := Version{
		version:   semver.MustParse("1.2.3"),
		isDefault: true,
	}

	if !reflect.DeepEqual(definedVersion, *builtVersion) {
		t.Errorf("version made through builder and version defined normally are not equal")
	}
}

func TestVersionListBuilder(t *testing.T) {
	version := NewVersionBuilder().
		Version(semver.MustParse("1.2.3")).
		Default(true).
		Build()
	overrideVersion := semver.MustParse("4.5.6")

	builtVersionList := NewVersionListBuilder().
		AvailableVersions([]*Version{version}).
		DefaultVersionOverride(overrideVersion).
		Build()

	definedVersionList := VersionList{
		availableVersions:      []*Version{version},
		defaultVersionOverride: overrideVersion,
	}

	if !reflect.DeepEqual(definedVersionList, *builtVersionList) {
		t.Errorf("version list made through builder and version list defined normally are not equal")
	}
}

func TestVersionListDefault(t *testing.T) {
	tests := []struct {
		Name                   string
		AvailableVersions      []*Version
		DefaultVersionOverride *semver.Version
		ExpectedDefault        *semver.Version
	}{
		{
			Name: "No default override",
			AvailableVersions: []*Version{
				NewVersionBuilder().
					Version(semver.MustParse("1.2.3")).
					Default(false).
					Build(),
				NewVersionBuilder().
					Version(semver.MustParse("2.3.4")).
					Default(false).
					Build(),
				NewVersionBuilder().
					Version(semver.MustParse("4.5.6")).
					Default(true).
					Build(),
			},
			DefaultVersionOverride: nil,
			ExpectedDefault:        semver.MustParse("4.5.6"),
		},
		{
			Name: "Default override",
			AvailableVersions: []*Version{
				NewVersionBuilder().
					Version(semver.MustParse("1.2.3")).
					Default(false).
					Build(),
				NewVersionBuilder().
					Version(semver.MustParse("2.3.4")).
					Default(false).
					Build(),
				NewVersionBuilder().
					Version(semver.MustParse("4.5.6")).
					Default(true).
					Build(),
			},
			DefaultVersionOverride: semver.MustParse("5.6.7"),
			ExpectedDefault:        semver.MustParse("5.6.7"),
		},
		{
			Name: "No default",
			AvailableVersions: []*Version{
				NewVersionBuilder().
					Version(semver.MustParse("1.2.3")).
					Default(false).
					Build(),
				NewVersionBuilder().
					Version(semver.MustParse("2.3.4")).
					Default(false).
					Build(),
				NewVersionBuilder().
					Version(semver.MustParse("4.5.6")).
					Default(false).
					Build(),
			},
			DefaultVersionOverride: nil,
			ExpectedDefault:        nil,
		},
	}

	for _, test := range tests {
		versionList := NewVersionListBuilder().
			AvailableVersions(test.AvailableVersions).
			DefaultVersionOverride(test.DefaultVersionOverride).
			Build()

		versionListDefault := versionList.Default()
		if !testSemverEquals(test.ExpectedDefault, versionListDefault) {
			t.Errorf("test name: %s, expected default %v does not match version list default %v", test.Name, test.ExpectedDefault, versionListDefault)
		}
	}
}

func testSemverEquals(version1 *semver.Version, version2 *semver.Version) bool {
	if version1 == nil && version2 == nil {
		return true
	}

	if version1 == nil && version2 != nil || version2 == nil && version1 != nil {
		return false
	}

	return version1.Equal(version2)
}
