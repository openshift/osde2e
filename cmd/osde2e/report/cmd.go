package report

import (
	"fmt"

	"github.com/openshift/osde2e/cmd/osde2e/common"
	"github.com/openshift/osde2e/pkg/reporting"
	"github.com/spf13/cobra"
)

var args struct {
	configString    string
	customConfig    string
	secretLocations string
	listReporters   bool
	listReportTypes bool
	output          string
	toSlack         bool
}

var Cmd = &cobra.Command{
	Use:   "report <reportName> <reportType>",
	Short: "OSDe2e reporting.",
	Long:  "Produces a report based on osde2e test runs.",
	Args:  cobra.OnlyValidArgs,
	RunE:  run,
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
	flags.BoolVar(
		&args.listReporters,
		"list-reporters",
		false,
		"List the reporters supported by the reporting tool",
	)
	flags.BoolVar(
		&args.listReportTypes,
		"list-report-types",
		false,
		"List the report types supported by the reporting tool for the given reporter",
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
	flags.BoolVar(
		&args.toSlack,
		"to-slack",
		false,
		"Send the report to Slack.",
	)
}

func run(cmd *cobra.Command, argv []string) error {
	if err := common.LoadConfigs(args.configString, args.customConfig, args.secretLocations); err != nil {
		return fmt.Errorf("error loading initial state: %v", err)
	}

	reporters := reporting.ListReporters()

	if args.listReporters {
		printListOfValidReporters(reporters)
		return nil
	}

	if len(argv) == 0 {
		return nil
	}

	reportName := argv[0]

	foundReporter := false
	for _, reporter := range reporters {
		if reportName == reporter {
			foundReporter = true
		}
	}

	if !foundReporter {
		fmt.Printf("Unable to find reporter %s.\n", reportName)
		printListOfValidReporters(reporters)
		return fmt.Errorf("unable to find reporter %s", reportName)
	}

	reportTypes, err := reporting.ListReportTypes(reportName)
	if err != nil {
		return fmt.Errorf("error getting report types: %v", err)
	}

	if args.listReportTypes {
		printListOfValidReportTypes(reportName, reportTypes)
	}

	if len(argv) == 1 {
		return nil
	}

	reportType := argv[1]

	foundReportType := false
	for _, supportedReportType := range reportTypes {
		if reportType == supportedReportType {
			foundReportType = true
		}
	}

	if !foundReportType {
		fmt.Printf("Unable to find report type %s for report %s.", reportType, reportName)
		printListOfValidReportTypes(reportName, reportTypes)
		return fmt.Errorf("unable to find report type %s for reporter %s", reportType, reportName)
	}

	report, err := reporting.GenerateReport(reportName, reportType)
	if err != nil {
		return fmt.Errorf("error while generating report: %v", err)
	}

	if args.toSlack {
		err = reporting.SendReportToSlack(reportName, report)

		if err != nil {
			return fmt.Errorf("error sending report to slack: %v", err)
		}
	} else {
		err = reporting.WriteReport(report, args.output)

		if err != nil {
			return fmt.Errorf("error writing report: %v", err)
		}
	}

	return nil
}

func printListOfValidReporters(reporters []string) {
	fmt.Println("Valid list of reporters:")
	for _, reporter := range reporters {
		fmt.Printf("- %s\n", reporter)
	}
}

func printListOfValidReportTypes(reportName string, reportTypes []string) {
	fmt.Printf("Valid list of report types for report %s:\n", reportName)
	for _, reportType := range reportTypes {
		fmt.Printf("- %s\n", reportType)
	}
}
