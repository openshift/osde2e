package krknai

import (
	"testing"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
)

func TestDetectContainerRuntime(t *testing.T) {
	runtime, err := detectContainerRuntime()
	// This test will pass if either podman or docker is installed
	// If neither is installed, it should return an error
	if err != nil {
		t.Logf("No container runtime found (expected in CI without containers): %v", err)
		return
	}

	if runtime == "" {
		t.Error("detectContainerRuntime() returned empty string without error")
	}

	t.Logf("Detected container runtime: %s", runtime)
}

func TestDefaultKrknAIImage(t *testing.T) {
	expected := "quay.io/krkn-chaos/krkn-ai:latest"
	if DefaultKrknAIImage != expected {
		t.Errorf("DefaultKrknAIImage = %q, want %q", DefaultKrknAIImage, expected)
	}
}

func TestKrknAIViperConfig(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
		mode     string // "discover" or "run"
	}{
		// Discover mode specific fields
		{"Namespace", config.KrknAI.Namespace, "default", "discover"},
		{"PodLabel", config.KrknAI.PodLabel, "", "discover"},
		{"NodeLabel", config.KrknAI.NodeLabel, "kubernetes.io/hostname", "discover"},
		{"SkipPodName", config.KrknAI.SkipPodName, "", "discover"},

		// Run mode specific fields (FitnessQuery, Scenarios)
		{"FitnessQuery", config.KrknAI.FitnessQuery, "", "run"},
		{"Scenarios", config.KrknAI.Scenarios, "", "run"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := viper.GetString(tt.key)
			if value != tt.expected {
				t.Errorf("viper.GetString(%q) = %q, want %q (mode: %s)", tt.key, value, tt.expected, tt.mode)
			}
		})
	}
}
