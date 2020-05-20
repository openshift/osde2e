package main

import (
	"fmt"
	"testing"
)

func TestRetryer(t *testing.T) {
	eventualSuccessCounter := 0

	tests := []struct {
		Name     string
		Function func() error
		Attempts int
		Success  bool
	}{
		{
			Name: "immediate success",
			Function: func() error {
				return nil
			},
			Attempts: 1,
			Success:  true,
		},
		{
			Name: "immediate failure",
			Function: func() error {
				return fmt.Errorf("failure")
			},
			Attempts: 3,
			Success:  false,
		},
		{
			Name: "eventual success",
			Function: func() error {
				if eventualSuccessCounter < 2 {
					eventualSuccessCounter = eventualSuccessCounter + 1
					return fmt.Errorf("failure")
				}

				return nil
			},
			Attempts: 3,
			Success:  true,
		},
	}

	for _, test := range tests {
		retryer := retryer()
		retryer.Tries = 3
		err := retryer.Do(test.Function)

		if test.Attempts != retryer.Attempts() {
			t.Fatalf("Test %s: number of expected attempts did not match: expected %d, got %d", test.Name, test.Attempts, retryer.Attempts())
		}

		if (err == nil) != test.Success {
			if test.Success {
				t.Fatalf("Test %s did not succeed as expected", test.Name)
			} else {
				t.Fatalf("Test %s did not fail as expected", test.Name)
			}
		}
	}
}
