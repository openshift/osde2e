package osde2e_scale

import (
	"flag"
	"testing"

	"github.com/openshift/osde2e/common"
	"github.com/openshift/osde2e/pkg/config"

	// import suites to be tested
	_ "github.com/openshift/osde2e/test/scale"
)

func init() {
	testing.Init()

	cfg := config.Cfg

	flag.StringVar(&cfg.File, "e2e-config", ".osde2e.yaml", "Config file for osde2e")
	flag.Parse()

	cfg.LoadFromYAML(cfg.File)

}

func TestScale(t *testing.T) {
	cfg := config.Cfg
	common.RunE2ETests(t, cfg)
}
