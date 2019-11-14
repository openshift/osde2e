package osde2e_scale

import (
	"testing"

	"github.com/openshift/osde2e/common"
	"github.com/openshift/osde2e/pkg/config"

	// import suites to be tested
	_ "github.com/openshift/osde2e/test/scale"
)

func TestScale(t *testing.T) {
	cfg := config.Cfg
	common.RunE2ETests(t, cfg)
}
