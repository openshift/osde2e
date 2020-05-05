package util

import (
	"testing"

	"github.com/Masterminds/semver"
)

func TestVersion420(t *testing.T) {
	tests := []struct {
		Name     string
		Version  *semver.Version
		Expected bool
	}{
		{
			Name:     "passes constraint",
			Version:  semver.MustParse("4.2.1"),
			Expected: true,
		},
		{
			Name:     "rc passes constraint",
			Version:  semver.MustParse("4.2.1-rc.0"),
			Expected: true,
		},
		{
			Name:     "fails constraint",
			Version:  semver.MustParse("4.1.9"),
			Expected: false,
		},
		{
			Name:     "rc fails constraint",
			Version:  semver.MustParse("4.1.9-rc.0"),
			Expected: false,
		},
	}

	for _, test := range tests {
		if Version420.Check(test.Version) != test.Expected {
			t.Errorf("test %s did not produce the expected result (%t) when using version %v", test.Name, test.Expected, test.Version)
		}
	}
}

func TestVersion440(t *testing.T) {
	tests := []struct {
		Name     string
		Version  *semver.Version
		Expected bool
	}{
		{
			Name:     "passes constraint",
			Version:  semver.MustParse("4.4.1"),
			Expected: true,
		},
		{
			Name:     "rc passes constraint",
			Version:  semver.MustParse("4.4.1-rc.0"),
			Expected: true,
		},
		{
			Name:     "fails constraint",
			Version:  semver.MustParse("4.3.9"),
			Expected: false,
		},
		{
			Name:     "rc fails constraint",
			Version:  semver.MustParse("4.3.9-rc.0"),
			Expected: false,
		},
	}

	for _, test := range tests {
		if Version440.Check(test.Version) != test.Expected {
			t.Errorf("test %s did not produce the expected result (%t) when using version %v", test.Name, test.Expected, test.Version)
		}
	}
}
