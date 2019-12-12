package osde2e_addons

import (
	"testing"

	"github.com/openshift/osde2e/common"
	"github.com/openshift/osde2e/pkg/config"

	// import suites to be tested
	_ "github.com/openshift/osde2e/test/addons"
)

func TestAddons(t *testing.T) {
	cfg := config.Cfg
	common.RunE2ETests(t, cfg)
}
