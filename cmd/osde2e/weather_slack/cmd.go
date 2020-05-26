package weather_slack

import (
	"fmt"
	"github.com/openshift/osde2e/cmd/osde2e/common"
	"github.com/openshift/osde2e/pkg/weather"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "weather-report-to-slack",
	Short: "Weather report to slack.",
	Long: "Produces a report based on osde2e test runs and sends it to a Slack webhook.",
	Args: cobra.OnlyValidArgs,
	RunE: run,
}

var args struct {
	configString string
	customConfig string
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
}

func run(cmd *cobra.Command, argv []string) error {
	if err := common.LoadConfigs(args.configString, args.customConfig); err != nil {
		return fmt.Errorf("error loading initial state: %v", err)
	}

	err := weather.SendReportToSlack()

	if err != nil {
		return fmt.Errorf("error while sending report to slack: %v", err)
	}

	return nil
}
