package weather

import (
	"fmt"

	"github.com/openshift/osde2e/pkg/common/report"
)

// GenerateWeatherReportForOSD will generate a JSON report for all jobs run by osde2e.
func GenerateWeatherReportForOSD(output string, outputType string) error {
	report, err := report.GenerateReport()

	if err != nil {
		return fmt.Errorf("error while generating report: %v", err)
	}

	if len(report.Jobs) == 0 {
		return fmt.Errorf("no jobs found while generating the weather report")
	}

	if outputType == "json" {
		err = report.WriteJSON(output)
	} else if outputType == "markdown" {
		err = report.WriteMarkdown(output)
	} else if outputType == "sd-report" {
		err = report.WriteSDReport(output)
	} else {
		err = fmt.Errorf("unrecognized output type: %s", outputType)
	}

	if err != nil {
		return fmt.Errorf("error while writing out report: %v", err)
	}

	return nil
}
