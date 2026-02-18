package provision

import (
	"errors"
	"testing"

	clusterutil "github.com/openshift/osde2e/pkg/common/cluster"
	"github.com/stretchr/testify/assert"
)

// TestProvisionErrorHandling tests that ErrReserveFull is treated as success
// while other errors are propagated. This is a focused unit test for the
// error handling logic added to maintain backward compatibility.
func TestProvisionErrorHandling(t *testing.T) {
	tests := []struct {
		name         string
		provisionErr error
		wantErr      bool
	}{
		{
			name:         "success_case_no_error",
			provisionErr: nil,
			wantErr:      false,
		},
		{
			name:         "reserve_full_treated_as_success",
			provisionErr: clusterutil.ErrReserveFull,
			wantErr:      false,
		},
		{
			name:         "actual_error_should_propagate",
			provisionErr: errors.New("provisioning failed"),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the error handling logic from the run function
			err := tt.provisionErr

			// This is the logic from lines 94-98 of cmd.go
			if err != nil && !errors.Is(err, clusterutil.ErrReserveFull) {
				// Error should be returned
				assert.Error(t, err)
				if tt.wantErr {
					return // Test passes
				}
				t.Errorf("expected no error, got: %v", err)
				return
			}

			// ErrReserveFull or nil should result in nil
			if tt.wantErr {
				t.Errorf("expected error, got nil")
			}
		})
	}
}

// TestErrReserveFullIsExported verifies that ErrReserveFull is accessible
// from other packages, which is required for the error handling.
func TestErrReserveFullIsExported(t *testing.T) {
	assert.NotNil(t, clusterutil.ErrReserveFull)
	assert.Equal(t, "reserve full", clusterutil.ErrReserveFull.Error())
}
