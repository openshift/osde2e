package gate

import (
	"context"
	"flag"
	"log"

	"github.com/google/subcommands"

	"github.com/openshift/osde2e/cmd/osde2e/common"
	"github.com/openshift/osde2e/pkg/gate"
)

// ReportAnalysisCommand is the command for interpreting gate reports for use by the OCP release gate canary job.
type ReportAnalysisCommand struct {
	configString string
	customConfig string

	subcommands.Command
}

// Name is the name of the gate-report command
func (*ReportAnalysisCommand) Name() string {
	return "gate-report-analysis"
}

// Synopsis is a short summary of the gate-report-anlysis command
func (*ReportAnalysisCommand) Synopsis() string {
	return "Interprets a gate report and returns a pass/fail value based on a report."
}

// Usage describes how the gate-report-analysis command is used
func (*ReportAnalysisCommand) Usage() string {
	return "gate-report-analysis <report-file>"
}

// SetFlags describes the arguments used by the gate-report-analysis command
func (t *ReportAnalysisCommand) SetFlags(f *flag.FlagSet) {
	f.StringVar(&t.configString, "configs", "", "A comma separated list of built in configs to use")
	f.StringVar(&t.customConfig, "custom-config", "", "Custom config file for osde2e")
}

// Execute actually executes the gate report analysis
func (t *ReportAnalysisCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if err := common.LoadConfigs(t.configString, t.customConfig); err != nil {
		log.Printf("error loading initial state: %v", err)
		return subcommands.ExitFailure
	}

	if f.NArg() != 1 {
		log.Printf("Unexpected number of arguments.")
		log.Printf(t.Usage())
		return subcommands.ExitFailure
	}

	releaseViable, err := gate.AnalyzeReport(f.Arg(0))

	if err != nil {
		log.Printf("error while analyzing report: %v", err)
		return subcommands.ExitFailure
	}

	if releaseViable {
		return subcommands.ExitSuccess
	}

	return subcommands.ExitFailure
}
