package execute_test

import (
	"testing"

	"github.com/openshift/osde2e/pkg/e2e/execute"
)

// TestNewGinkgoExecutor tests the creation of a new Ginkgo executor
func TestNewGinkgoExecutor(t *testing.T) {
	executor := execute.NewGinkgoExecutor()
	
	if executor == nil {
		t.Error("NewGinkgoExecutor returned nil")
	}
}

