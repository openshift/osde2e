package test

import (
	"context"
	"flag"
	"log"
	"strings"

	"github.com/google/subcommands"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/load"
	"github.com/openshift/osde2e/pkg/common/state"
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

// TestCommand is the command for running end to end tests on OSD clusters
type TestCommand struct {
	configString string
	customConfig string

	subcommands.Command
}

// Name is the name of the test command
func (*TestCommand) Name() string {
	return "test"
}

// Synopsis is a short summary of the test command
func (*TestCommand) Synopsis() string {
	return "Runs end to end tests on a cluster using the provided arguments."
}

// Usage describes how the test command is used
func (*TestCommand) Usage() string {
	return "test [-configs config1,config2] [-customConfig osde2e-custom-config.yaml]"
}

// SetFlags describes the arguments used by the test command
func (t *TestCommand) SetFlags(f *flag.FlagSet) {
	f.StringVar(&t.configString, "configs", "", "A comma separated list of built in configs to use")
	f.StringVar(&t.customConfig, "custom-config", "", "Custom config file for osde2e")
}

// Execute actually executes the tests
func (t *TestCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	var configs []string
	if t.configString != "" {
		configs = strings.Split(t.configString, ",")
	}

	for _, config := range configs {
		log.Printf("Will load config %s", config)
	}

	// Load config and initial state
	if err := load.IntoObject(config.Instance, configs, t.customConfig); err != nil {
		log.Printf("error loading config: %v", err)
		return subcommands.ExitFailure
	}

	if err := load.IntoObject(state.Instance, configs, t.customConfig); err != nil {
		log.Printf("error loading initial state: %v", err)
		return subcommands.ExitFailure
	}

	if e2e.RunTests() {
		return subcommands.ExitSuccess
	}

	return subcommands.ExitFailure
}
