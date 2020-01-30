package osde2e_addons

import (
	"testing"

	"github.com/openshift/osde2e/common"

	// import suites to be tested
	_ "github.com/openshift/osde2e/test/addons"
)

func TestAddons(t *testing.T) {
	common.RunE2ETests(t)
}
