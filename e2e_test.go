package osde2e

import (
	"testing"

	"github.com/openshift/osde2e/pkg/config"

	// import suites to be tested
	_ "github.com/openshift/osde2e/pkg/verify"
)

func TestE2E(t *testing.T) {
	cfg := config.Cfg
	RunE2ETests(t, cfg)
}
