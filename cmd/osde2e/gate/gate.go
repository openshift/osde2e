package gate

import (
	"context"
	"flag"
	"log"

	"github.com/google/subcommands"

	"github.com/openshift/osde2e/cmd/osde2e/common"
	"github.com/openshift/osde2e/pkg/gate"
)

// GateCommand is the command for running end to end tests on OSD clusters
type GateCommand struct {
	configString string
	customConfig string
	output       string

	subcommands.Command
}

// Name is the name of the gate command
func (*GateCommand) Name() string {
	return "gate"
}

// Synopsis is a short summary of the gate command
func (*GateCommand) Synopsis() string {
	return "Analyzes previous and determines whether a version of OpenShift is ready to ship."
}

// Usage describes how the gate command is used
func (*GateCommand) Usage() string {
	return "gate <environment> <openshift-version>"
}

// SetFlags describes the arguments used by the gate command
func (t *GateCommand) SetFlags(f *flag.FlagSet) {
	f.StringVar(&t.configString, "configs", "", "A comma separated list of built in configs to use")
	f.StringVar(&t.customConfig, "custom-config", "", "Custom config file for osde2e")
	f.StringVar(&t.output, "output", "-", "Where to output the report. Use '-' for standard out")
}

// Execute actually executes the gate analysis
func (t *GateCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if err := common.LoadConfigs(t.configString, t.customConfig); err != nil {
		log.Printf("error loading initial state: %v", err)
		return subcommands.ExitFailure
	}

	if f.NArg() != 2 {
		log.Printf("Unexpected number of arguments.")
		log.Printf(t.Usage())
		return subcommands.ExitFailure
	}

	releaseViable, err := gate.GenerateReleaseReportForOSD(f.Arg(0), f.Arg(1), t.output)

	if err != nil {
		log.Printf("error while checking for release viability: %v", err)
		return subcommands.ExitFailure
	}

	if releaseViable {
		return subcommands.ExitSuccess
	}

	return subcommands.ExitFailure
}
