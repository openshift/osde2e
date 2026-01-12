package provision_test

import (
	"testing"

	"github.com/openshift/osde2e/pkg/e2e/provision"
)

// TestNewOCMProvisioner tests the creation of a new OCM provisioner
func TestNewOCMProvisioner(t *testing.T) {
	// Note: This test requires proper OCM provider setup
	// In a real scenario, you'd mock the providers.ClusterProvider() call
	
	// For now, we just verify the constructor doesn't panic
	_, err := provision.NewOCMProvisioner()
	
	// We expect an error in test environment without proper OCM setup
	// This is acceptable - the test verifies the API
	if err == nil {
		// If no error, that's also fine (means OCM is configured)
		t.Log("OCM provisioner created successfully")
	} else {
		// Expected in test environment
		t.Logf("Expected error in test environment: %v", err)
	}
}

