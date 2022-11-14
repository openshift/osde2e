package reporting

import (
	"fmt"
	"sort"
	"strings"

	"github.com/openshift/osde2e/pkg/common/aws"
	"github.com/openshift/osde2e/pkg/reporting/reporters"
	"github.com/openshift/osde2e/pkg/reporting/spi"
	"github.com/openshift/osde2e/pkg/reporting/templates"
)

// ListReporters will list all possible reporters.
func ListReporters() []string {
	return spi.ListReporters()
}

// ListReportTypes will list all possible report types for a report.
func ListReportTypes(reportName string) ([]string, error) {
	reporters := ListReporters()

	foundReportName := false
	for _, reporter := range reporters {
		if reportName == reporter {
			foundReportName = true
		}
	}

	if !foundReportName {
		return nil, fmt.Errorf("unable to find report %s", reportName)
	}

	// We always support JSON.
	reportTypes := append(templates.ListTemplates(reportName), "json")

	sort.Sort(sort.StringSlice(reportTypes))
	return reportTypes, nil
}

// GenerateReport will generate the specified report and return it.
func GenerateReport(reporterName string, reportType string) ([]byte, error) {
	reporter, err := reporters.GetReporter(reporterName)
	if err != nil {
		return nil, fmt.Errorf("error getting reporter: %v", err)
	}

	rawReport, err := reporter.GenerateReport(reportType)
	if err != nil {
		return nil, fmt.Errorf("error generating report: %v", err)
	}

	return rawReport, nil
}

// WriteReport will write the raw report to a given output.
func WriteReport(report []byte, output string) error {
	if strings.HasPrefix(output, "s3") {
		aws.WriteToS3(output, report)
	} else {
		writer, err := createWriter(output)
		if err != nil {
			return fmt.Errorf("error while creating writer: %v", err)
		}
		defer writer.Close()

		_, err = writer.Write(append(report, '\n'))

		if err != nil {
			return fmt.Errorf("error while writing report to output: %v", err)
		}
	}

	return nil
}
