package runner

import (
	"testing"

	"github.com/openshift/osde2e/pkg/e2e"

	// import suites to be tested
	_ "github.com/openshift/osde2e/pkg/e2e/addons"
	_ "github.com/openshift/osde2e/pkg/e2e/openshift"
	_ "github.com/openshift/osde2e/pkg/e2e/operators"
	_ "github.com/openshift/osde2e/pkg/e2e/scale"
	_ "github.com/openshift/osde2e/pkg/e2e/state"
	_ "github.com/openshift/osde2e/pkg/e2e/verify"
	_ "github.com/openshift/osde2e/pkg/e2e/workloads/guestbook"
)

func TestRunner(t *testing.T) {
	e2e.RunTests(t)
}
