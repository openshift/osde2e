package test

import (
	"context"
	"flag"
	"log"

	"github.com/google/subcommands"

	"github.com/openshift/osde2e/cmd/osde2e/common"
	"github.com/openshift/osde2e/pkg/e2e"

	// import suites to be tested
	_ "github.com/openshift/osde2e/pkg/e2e/addons"
	_ "github.com/openshift/osde2e/pkg/e2e/openshift"
	_ "github.com/openshift/osde2e/pkg/e2e/operators"
	_ "github.com/openshift/osde2e/pkg/e2e/osd"
	_ "github.com/openshift/osde2e/pkg/e2e/scale"
	_ "github.com/openshift/osde2e/pkg/e2e/state"
	_ "github.com/openshift/osde2e/pkg/e2e/verify"
	_ "github.com/openshift/osde2e/pkg/e2e/workloads/guestbook"
)

// Command is the command for running end to end tests on OSD clusters
type Command struct {
	configString string
	customConfig string

	subcommands.Command
}

// Name is the name of the test command
func (*Command) Name() string {
	return "test"
}

// Synopsis is a short summary of the test command
func (*Command) Synopsis() string {
	return "Runs end to end tests on a cluster using the provided arguments."
}

// Usage describes how the test command is used
func (*Command) Usage() string {
	return "test [-configs config1,config2] [-customConfig osde2e-custom-config.yaml]"
}

// SetFlags describes the arguments used by the test command
func (t *Command) SetFlags(f *flag.FlagSet) {
	f.StringVar(&t.configString, "configs", "", "A comma separated list of built in configs to use")
	f.StringVar(&t.customConfig, "custom-config", "", "Custom config file for osde2e")
}

// Execute actually executes the tests
func (t *Command) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if err := common.LoadConfigs(t.configString, t.customConfig); err != nil {
		log.Printf("error loading initial state: %v", err)
		return subcommands.ExitFailure
	}

	if e2e.RunTests() {
		return subcommands.ExitSuccess
	}

	return subcommands.ExitFailure
}
