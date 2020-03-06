package gate

import (
	"fmt"
	"log"

	"github.com/openshift/osde2e/pkg/common/report"
)

// AnalyzeReport takes a given report file, gives a plaintext report on it, and returns true if the report indicates a release is viable, false otherwise.
func AnalyzeReport(input string) (bool, error) {
	report, err := report.ReadGateReportFromFile(input)

	if err != nil {
		return false, fmt.Errorf("error trying to read the report: %v", err)
	}

	log.Printf("The report ran tests over the following versions:")
	for _, version := range report.Versions {
		log.Printf("* %s", version)
	}

	log.Printf("\n")

	if report.Viable {
		log.Printf("Release is viable!")
	} else {
		log.Printf("The following failing tests were detected:")
		for _, failingTest := range report.FailingTests {
			log.Printf("* %s", failingTest)
		}
	}

	return report.Viable, nil
}
