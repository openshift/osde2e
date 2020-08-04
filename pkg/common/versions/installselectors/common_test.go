package installselectors

import (
	"testing"

	"github.com/Masterminds/semver"
)

// Testing utility function for asserting two versions should be equal.
func failIfVersionsNotEqual(t *testing.T, testName string, selectedVersion *semver.Version, expectedVersion *semver.Version) {
	if (selectedVersion == nil || expectedVersion == nil) && selectedVersion != expectedVersion {
		t.Errorf("test %s: expected selected version (%v) to match expected version (%v) and one is nil", testName, selectedVersion, expectedVersion)
	} else if selectedVersion != nil && !selectedVersion.Equal(expectedVersion) {
		t.Errorf("test %s: selected version (%v) does not match expected version (%v)", testName, selectedVersion, expectedVersion)
	}
}
