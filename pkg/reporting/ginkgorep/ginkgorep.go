/*
Package ginkgorep provides custom ginkgo reporters.
*/
package ginkgorep

import (
	"fmt"

	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"
)

// PhaseReporter wraps a normal JUnitReporter with the ability to specify
// a testing phase in the names of the generated XML tests.
type PhaseReporter struct {
	*reporters.JUnitReporter
	phase string
}

func NewPhaseReporter(phase, filename string) PhaseReporter {
	return PhaseReporter{
		phase:         phase,
		JUnitReporter: reporters.NewJUnitReporter(filename),
	}
}

var _ reporters.Reporter = PhaseReporter{}

func (r PhaseReporter) SpecDidComplete(specSummary *types.SpecSummary) {
	// Inject the phase into the name of the test case.
	const testCaseNameIndex = 1
	if len(specSummary.ComponentTexts) > testCaseNameIndex {
		text := fmt.Sprintf("[%s] %s", r.phase, specSummary.ComponentTexts[testCaseNameIndex])
		specSummary.ComponentTexts[testCaseNameIndex] = text
	}
	r.JUnitReporter.SpecDidComplete(specSummary)
}
