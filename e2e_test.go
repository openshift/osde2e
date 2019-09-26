package osde2e

import (
	"testing"

	"github.com/openshift/osde2e/pkg/config"

	// import suites to be tested
	_ "github.com/openshift/osde2e/test/openshift"
	_ "github.com/openshift/osde2e/test/operators"
	_ "github.com/openshift/osde2e/test/state"
	_ "github.com/openshift/osde2e/test/verify"
)

func TestE2E(t *testing.T) {
	cfg := config.Cfg
	RunE2ETests(t, cfg)
}
