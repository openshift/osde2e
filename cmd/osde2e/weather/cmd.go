package weather

import (
	"fmt"

	"github.com/openshift/osde2e/cmd/osde2e/common"
	"github.com/openshift/osde2e/pkg/weather"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "weather-report",
	Short: "Weather report.",
	Long:  "Produces a report based on osde2e test runs.",
	Args:  cobra.OnlyValidArgs,
	RunE:  run,
}

var args struct {
	configString    string
	customConfig    string
	secretLocations string
	output          string
	outputType      string
}

func init() {
	flags := Cmd.Flags()

	flags.StringVar(
		&args.configString,
		"configs",
		"",
		"A comma separated list of built in configs to use",
	)
	flags.StringVar(
		&args.customConfig,
		"custom-config",
		"",
		"Custom config file for osde2e",
	)
	flags.StringVar(
		&args.secretLocations,
		"secret-locations",
		"",
		"A comma separated list of possible secret directory locations for loading secret configs.",
	)
	flags.StringVar(
		&args.output,
		"output",
		"-",
		"Where to output the report. Use '-' for standard out",
	)
	flags.StringVar(
		&args.outputType,
		"outputType",
		"json",
		"What format to output the report in. Defaults to json.",
	)

	Cmd.RegisterFlagCompletionFunc("outputType", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "markdown"}, cobra.ShellCompDirectiveDefault
	})

}

func run(cmd *cobra.Command, argv []string) error {
	if err := common.LoadConfigs(args.configString, args.customConfig, args.secretLocations); err != nil {
		return fmt.Errorf("error loading initial state: %v", err)
	}

	err := weather.GenerateWeatherReportForOSD(args.output, args.outputType)

	if err != nil {
		return fmt.Errorf("error while generating report: %v", err)
	}

	return nil
}
