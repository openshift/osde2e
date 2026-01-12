package report_test

import (
	"testing"

	"github.com/openshift/osde2e/pkg/e2e/report"
)

// TestNewCompositeReporter tests the creation of a new composite reporter
func TestNewCompositeReporter(t *testing.T) {
	reporter := report.NewCompositeReporter()
	
	if reporter == nil {
		t.Error("NewCompositeReporter returned nil")
	}
}

// TestNewArtifactCollector tests the creation of a new artifact collector
func TestNewArtifactCollector(t *testing.T) {
	collector := report.NewArtifactCollector()
	
	if collector == nil {
		t.Error("NewArtifactCollector returned nil")
	}
}

