package templates

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// WriteReport will write the report to a byte array for later processing.
func WriteReport(data interface{}, reportName string, reportType string) ([]byte, error) {
	report := new(bytes.Buffer)
	if reportType == "json" {
		reportBytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("error while marshaling report into JSON: %v", err)
		}

		report.Write(reportBytes)
	} else {
		template, err := cache.getReportTemplate(reportName, reportType)
		if err != nil {
			return nil, fmt.Errorf("error loading specified template: %v", err)
		}

		if err := template.ExecuteTemplate(report, template.Name(), data); err != nil {
			return nil, fmt.Errorf("error while creating %s report: %v", reportName, err)
		}
	}

	return append(report.Bytes(), '\n'), nil
}
