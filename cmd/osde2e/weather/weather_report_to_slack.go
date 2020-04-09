package weather

import (
	"context"
	"flag"
	"log"

	"github.com/google/subcommands"

	"github.com/openshift/osde2e/cmd/osde2e/common"
	"github.com/openshift/osde2e/pkg/weather"
)

// ReportToSlackCommand is the command for generating an OSD weather report based on osde2e test runs and sending it to a Slack webhook.
type ReportToSlackCommand struct {
	configString string
	customConfig string

	subcommands.Command
}

// Name is the name of the weather-report command
func (*ReportToSlackCommand) Name() string {
	return "weather-report-to-slack"
}

// Synopsis is a short summary of the weather-report command
func (*ReportToSlackCommand) Synopsis() string {
	return "Produces a report based on osde2e test runs and sends it to a Slack webhook."
}

// Usage describes how the weather-report-to-slack command is used
func (*ReportToSlackCommand) Usage() string {
	return "weather-report-to-slack"
}

// SetFlags describes the arguments used by the weather-report command
func (t *ReportToSlackCommand) SetFlags(f *flag.FlagSet) {
	f.StringVar(&t.configString, "configs", "", "A comma separated list of built in configs to use")
	f.StringVar(&t.customConfig, "custom-config", "", "Custom config file for osde2e")
}

// Execute actually generates the weather report
func (t *ReportToSlackCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if err := common.LoadConfigs(t.configString, t.customConfig); err != nil {
		log.Printf("error loading initial state: %v", err)
		return subcommands.ExitFailure
	}

	if f.NArg() != 0 {
		log.Printf("Unexpected number of arguments.")
		log.Printf(t.Usage())
		return subcommands.ExitFailure
	}

	err := weather.SendReportToSlack()

	if err != nil {
		log.Printf("error while sending report to slack: %v", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
