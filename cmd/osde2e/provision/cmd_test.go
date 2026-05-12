package provision

import (
	"errors"
	"fmt"
	"testing"

	clusterutil "github.com/openshift/osde2e/pkg/common/cluster"
)

func TestHandleProvisionError(t *testing.T) {
	tests := []struct {
		name        string
		inputError  error
		expectError bool
	}{
		{
			name:        "success returns nil",
			inputError:  nil,
			expectError: false,
		},
		{
			name:        "ErrReserveFull returns nil (treated as success)",
			inputError:  clusterutil.ErrReserveFull,
			expectError: false,
		},
		{
			name:        "wrapped ErrReserveFull returns nil",
			inputError:  fmt.Errorf("failed to set up or retrieve cluster: %w", clusterutil.ErrReserveFull),
			expectError: false,
		},
		{
			name:        "OIDC error returns error (SDCICD-1752)",
			inputError:  errors.New("failed to set up or retrieve cluster: could not launch cluster: create oidc config failed"),
			expectError: true,
		},
		{
			name:        "IAM permission error returns error (SDCICD-1752)",
			inputError:  errors.New("failed to set up or retrieve cluster: AccessDenied"),
			expectError: true,
		},
		{
			name:        "cluster never ready returns error",
			inputError:  errors.New("cluster never became ready: timeout"),
			expectError: true,
		},
		{
			name:        "health check failure returns error",
			inputError:  errors.New("cluster failed health check: nodes not ready"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handleProvisionError(tt.inputError)

			if tt.expectError && result == nil {
				t.Errorf("expected error but got nil for input: %v", tt.inputError)
			}
			if !tt.expectError && result != nil {
				t.Errorf("expected nil but got error: %v", result)
			}
		})
	}
}

