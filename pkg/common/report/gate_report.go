package report

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/openshift/osde2e/pkg/common/aws"
)

// GateReport is the gating report.
type GateReport struct {
	Viable       bool     `json:"viable"`
	Versions     []string `json:"versions"`
	FailingTests []string `json:"failingTests,omitempty"`
}

// ToOutput will write a JSON encoded version of the gate report to the supplied output.
// Output will behave in a way specified by the createWriter function.
func (g *GateReport) ToOutput(output string) error {

	jsonReport, err := json.MarshalIndent(g, "", "  ")

	if err != nil {
		return fmt.Errorf("error while marshaling report into JSON: %v", err)
	}

	jsonReport = append(jsonReport, '\n')

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

// ReadGateReportFromFile will unmarshal a JSON representation of the GateReport from a file and return the report.
func ReadGateReportFromFile(input string) (*GateReport, error) {
	var data []byte
	var err error

	if strings.HasPrefix(input, "s3") {
		data, err = aws.ReadFromS3(input)

		if err != nil {
			return nil, err
		}
	} else {
		data, err = ioutil.ReadFile(input)

		if err != nil {
			return nil, fmt.Errorf("error while reading report file: %v", err)
		}
	}

	report := &GateReport{}
	err = json.Unmarshal(data, report)

	if err != nil {
		return nil, fmt.Errorf("error while unmarshaling report file: %v", err)
	}

	return report, nil
}
