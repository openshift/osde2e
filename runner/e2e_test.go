package runner

import (
	"testing"

	"github.com/openshift/osde2e/common"

	// import suites to be tested
	_ "github.com/openshift/osde2e/test/addons"
	_ "github.com/openshift/osde2e/test/openshift"
	_ "github.com/openshift/osde2e/test/operators"
	_ "github.com/openshift/osde2e/test/scale"
	_ "github.com/openshift/osde2e/test/state"
	_ "github.com/openshift/osde2e/test/verify"
	_ "github.com/openshift/osde2e/test/workloads/guestbook"
)

func TestRunner(t *testing.T) {
	common.RunE2ETests(t)
}
