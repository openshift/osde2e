package osde2e_test

import (
	"flag"
	"testing"

	"github.com/openshift/osde2e/common"
	"github.com/openshift/osde2e/pkg/config"

	// import suites to be tested
	_ "github.com/openshift/osde2e/test/openshift"
	_ "github.com/openshift/osde2e/test/operators"
	_ "github.com/openshift/osde2e/test/state"
	_ "github.com/openshift/osde2e/test/verify"
	_ "github.com/openshift/osde2e/test/workloads/guestbook"
)

func init() {
	var filename string
	testing.Init()

	cfg := config.Cfg

	flag.StringVar(&filename, "e2e-config", ".osde2e.yaml", "Config file for osde2e")
	flag.Parse()

	cfg.LoadFromYAML(filename)

}
func TestE2E(t *testing.T) {
	cfg := config.Cfg
	common.RunE2ETests(t, cfg)
}
