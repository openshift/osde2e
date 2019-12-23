package osde2e_addons

import (
	"flag"
	"testing"

	"github.com/openshift/osde2e/common"
	"github.com/openshift/osde2e/pkg/config"

	// import suites to be tested
	_ "github.com/openshift/osde2e/test/addons"
)

func init() {
	var filename string
	testing.Init()

	cfg := config.Cfg

	flag.StringVar(&filename, "e2e-config", ".osde2e.yaml", "Config file for osde2e")
	flag.Parse()

	cfg.LoadFromYAML(filename)

}

func TestAddons(t *testing.T) {
	cfg := config.Cfg
	common.RunE2ETests(t, cfg)
}
