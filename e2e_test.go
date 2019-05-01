package osde2e

import (
	"testing"

	// import suites to be tested
	_ "github.com/openshift/osde2e/pkg/verify"
)

func TestE2E(t *testing.T) {
	RunE2ETests(t, Cfg)
}
