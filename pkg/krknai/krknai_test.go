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
	}{
		{"Mode", config.KrknAI.Mode, "discover"},
		{"Namespace", config.KrknAI.Namespace, "default"},
		{"PodLabel", config.KrknAI.PodLabel, ""},
		{"NodeLabel", config.KrknAI.NodeLabel, "kubernetes.io/hostname"},
		{"SkipPodName", config.KrknAI.SkipPodName, ""},
		{"Verbose", config.KrknAI.Verbose, "2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := viper.GetString(tt.key)
			if value != tt.expected {
				t.Errorf("viper.GetString(%q) = %q, want %q", tt.key, value, tt.expected)
			}
		})
	}
}
