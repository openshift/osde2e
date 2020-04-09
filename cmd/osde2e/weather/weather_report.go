package weather

import (
	"context"
	"flag"
	"log"

	"github.com/google/subcommands"

	"github.com/openshift/osde2e/cmd/osde2e/common"
	"github.com/openshift/osde2e/pkg/weather"
)

// ReportCommand is the command for generating an OSD weather report based on osde2e test runs.
type ReportCommand struct {
	configString string
	customConfig string
	output       string
	outputType   string

	subcommands.Command
}

// Name is the name of the weather-report command
func (*ReportCommand) Name() string {
	return "weather-report"
}

// Synopsis is a short summary of the weather-report command
func (*ReportCommand) Synopsis() string {
	return "Produces a report based on osde2e test runs."
}

// Usage describes how the weather-report command is used
func (*ReportCommand) Usage() string {
	return "weather-report"
}

// SetFlags describes the arguments used by the weather-report command
func (t *ReportCommand) SetFlags(f *flag.FlagSet) {
	f.StringVar(&t.configString, "configs", "", "A comma separated list of built in configs to use")
	f.StringVar(&t.customConfig, "custom-config", "", "Custom config file for osde2e")
	f.StringVar(&t.output, "output", "-", "Where to output the report. Use '-' for standard out")
	f.StringVar(&t.outputType, "outputType", "json", "What format to output the report in. Defaults to json.")
}

// Execute actually generates the weather report
func (t *ReportCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if err := common.LoadConfigs(t.configString, t.customConfig); err != nil {
		log.Printf("error loading initial state: %v", err)
		return subcommands.ExitFailure
	}

	if f.NArg() != 0 {
		log.Printf("Unexpected number of arguments.")
		log.Printf(t.Usage())
		return subcommands.ExitFailure
	}

	err := weather.GenerateWeatherReportForOSD(t.output, t.outputType)

	if err != nil {
		log.Printf("error while generating report: %v", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
