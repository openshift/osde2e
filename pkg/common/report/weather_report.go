package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/openshift/osde2e/pkg/common/aws"
	"github.com/openshift/osde2e/pkg/common/templates"
)

var markdownTemplate *template.Template

var sdReportTemplate *template.Template

func init() {
	var err error

	markdownTemplate, err = templates.LoadTemplate("/assets/reports/markdown.template")

	if err != nil {
		panic(fmt.Sprintf("error loading markdown template: %v", err))
	}

	sdReportTemplate, err = templates.LoadTemplate("/assets/reports/sd-report.template")

	if err != nil {
		panic(fmt.Sprintf("error loading SD report template: %v", err))
	}
}

// WeatherReport is the weather report.
type WeatherReport struct {
	ReportDate time.Time   `json:"reportDate"`
	Provider   string      `json:"provider"`
	Jobs       []JobReport `json:"jobs"`

	// We want the sort interface so that we can sort jobs and produce stable, comparable reports.
	sort.Interface `json:"-"`
}

// JobReport is a report for an individual job.
type JobReport struct {
	Name         string        `json:"name"`
	Viable       bool          `json:"viable"`
	Color        string        `json:"color"`
	JobIDsReport []JobIDReport `json:"jobIDsReport"`
	Versions     []string      `json:"versions"`
	PassRate     float64       `json:"passRate"`
	FailingTests []string      `json:"failingTests,omitempty"`
}

// JobIDReport combines the job ID, pass rate, and a color for the job run together.
type JobIDReport struct {
	JobID          int64    `json:"jobID"`
	PassRate       float64  `json:"passRate"`
	JobColor       string   `json:"jobColor"`
	InstallVersion string   `json:"installVersion"`
	UpgradeVersion string   `json:"upgradeVersion"`
	FailingTests   []string `json:"failingTests,omitempty"`
}

// Len is the number of jobs in the weather report.
func (w WeatherReport) Len() int {
	return len(w.Jobs)
}

// Less reports whether the element with index i should sort before the element with index j.
func (w WeatherReport) Less(i, j int) bool {
	return w.Jobs[i].Name < w.Jobs[j].Name
}

// Swap swaps the elements with indexes i and j.
func (w WeatherReport) Swap(i, j int) {
	w.Jobs[i], w.Jobs[j] = w.Jobs[j], w.Jobs[i]
}

// ToJSON will convert the weather report into a JSON object.
func (w WeatherReport) ToJSON() ([]byte, error) {
	jsonReport, err := json.MarshalIndent(w, "", "  ")

	if err != nil {
		return nil, fmt.Errorf("error while marshaling report into JSON: %v", err)
	}

	return append(jsonReport, '\n'), nil
}

// WriteJSON will write a JSON encoded version of the weather report to the supplied output.
// Output will behave in a way specified by the createWriter function.
func (w WeatherReport) WriteJSON(output string) error {
	jsonReport, err := w.ToJSON()

	if err != nil {
		return fmt.Errorf("error while generating JSON: %v", err)
	}

	if strings.HasPrefix(output, "s3") {
		aws.WriteToS3(output, jsonReport)
	} else {
		writer, err := createWriter(output)
		if err != nil {
			return fmt.Errorf("error while creating writer: %v", err)
		}
		defer writer.Close()

		_, err = writer.Write(append(jsonReport, '\n'))

		if err != nil {
			return fmt.Errorf("error while writing report to output: %v", err)
		}
	}

	return nil
}

// ToMarkdown will convert the weather report into a Markdown rendering.
func (w WeatherReport) ToMarkdown() ([]byte, error) {
	markdownReportBuffer := new(bytes.Buffer)
	if err := markdownTemplate.ExecuteTemplate(markdownReportBuffer, markdownTemplate.Name(), w); err != nil {
		return nil, fmt.Errorf("error while creating markdown report: %v", err)
	}

	return append(markdownReportBuffer.Bytes(), '\n'), nil
}

// ToSDReport will convert the weather report into a service delivery markdown report.
func (w WeatherReport) ToSDReport() ([]byte, error) {
	sdReportBuffer := new(bytes.Buffer)
	if err := sdReportTemplate.ExecuteTemplate(sdReportBuffer, sdReportTemplate.Name(), w); err != nil {
		return nil, fmt.Errorf("error while creating SD report: %v", err)
	}

	return append(sdReportBuffer.Bytes(), '\n'), nil
}

// WriteMarkdown will write a markdown version of the weather report to the supplied output.
// Output will behave in a way specified by the createWriter function.
func (w WeatherReport) WriteMarkdown(output string) error {
	markdownReport, err := w.ToMarkdown()

	if err != nil {
		return fmt.Errorf("error while generating markdown: %v", err)
	}

	return w.writeRawReport(output, markdownReport)
}

// WriteSDReport will write an SD report version of the weather report to the supplied output.
// Output will behave in a way specified by the createWriter function.
func (w WeatherReport) WriteSDReport(output string) error {
	sdReport, err := w.ToSDReport()

	if err != nil {
		return fmt.Errorf("error while generating SD report: %v", err)
	}

	return w.writeRawReport(output, sdReport)
}

func (w WeatherReport) writeRawReport(output string, report []byte) error {
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
