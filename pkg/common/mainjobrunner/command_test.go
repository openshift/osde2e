package mainjobrunner

import "testing"

func TestCommand(t *testing.T) {
	// copy default runner
	def := *DefaultMainJobRunner
	r := &def

	_, err := r.Command()
	if err != nil {
		t.Fatalf("couldn't template default runner: %v", err)
	}
}
