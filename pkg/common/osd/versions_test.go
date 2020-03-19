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
