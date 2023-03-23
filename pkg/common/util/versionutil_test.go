package util

import (
	"testing"

	"github.com/Masterminds/semver"
)

func TestVersionConstraint(t *testing.T) {
	Version420, err := semver.NewConstraint(">= 4.2.0-0")
	if err != nil {
		t.Errorf("failed to build test version")
	}

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
