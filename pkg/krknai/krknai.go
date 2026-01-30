// Package krknai provides an orchestrator implementation for Kraken AI-powered chaos testing.
package krknai

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-logr/logr"
	"github.com/openshift/osde2e-common/pkg/clients/openshift"
	"github.com/openshift/osde2e-common/pkg/clients/prometheus"
	"github.com/openshift/osde2e/pkg/common/cluster"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/orchestrator"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/spi"
	"gopkg.in/yaml.v3"
)

const (
	// DefaultKrknAIImage is the default container image for Kraken AI chaos testing.
	// This value is also set as the viper default in config.KrknAI.Image.
	DefaultKrknAIImage = "quay.io/krkn-chaos/krkn-ai:latest"

	// Container mount paths
	containerMountPath   = "/mount"
	containerResultsPath = "/krknresults/"

	// File names
	kubeconfigFileName = "kubeconfig"
	krknConfigFileName = "krkn-ai.yaml"
)

// KrknAI implements the orchestrator.Orchestrator interface for Kraken AI chaos testing.
type KrknAI struct {
	provider spi.Provider
	result   *orchestrator.Result
}

// New creates a new KrknAI orchestrator instance.
func New(ctx context.Context) (orchestrator.Orchestrator, error) {
	provider, err := providers.ClusterProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster provider: %w", err)
	}

	return &KrknAI{
		provider: provider,
		result: &orchestrator.Result{
			ExitCode: config.Success,
		},
	}, nil
}

// Provision prepares the test environment by provisioning or reusing a cluster,
// loading kubeconfig, and installing required addons.
func (k *KrknAI) Provision(ctx context.Context) error {
	log.Println("Starting cluster provisioning")

	// Load cluster context (kubeconfig and cluster ID)
	if err := cluster.LoadClusterContext(); err != nil {
		return fmt.Errorf("failed to load cluster context: %w", err)
	}

	// Provision or reuse cluster
	cl, err := cluster.ProvisionOrReuseCluster(k.provider)
	if err != nil {
		return fmt.Errorf("failed to provision cluster: %w", err)
	}

	k.result.ClusterID = cl.ID()

	return nil
}

// Execute runs the configured test suites including chaos testing scenarios.
// The execution flow: discover mode -> update YAML -> run mode
func (k *KrknAI) Execute(ctx context.Context) error {
	k.result.TestsPassed = true
	viper.Set(config.Cluster.Passing, k.result.TestsPassed)

	if !viper.GetBool(config.DryRun) {
		// Step 1: Run discover mode to identify chaos targets
		log.Println("Krkn-ai discover mode")
		if err := k.runKrknContainer(ctx, config.KrknAIModeDiscover); err != nil {
			return k.handleExecutionError(fmt.Errorf("discover mode failed: %w", err))
		}

		// Step 2: Update the YAML config with discovered targets (skip in dry-run mode)
		log.Println("Updating config with discovered targets")
		if err := k.updateKrknConfig(); err != nil {
			return k.handleExecutionError(fmt.Errorf("failed to update config: %w", err))
		}

		// Step 3: Run run mode with the updated config
		log.Println("Krkn-ai run mode")
		if err := k.runKrknContainer(ctx, config.KrknAIModeRun); err != nil {
			return k.handleExecutionError(fmt.Errorf("run mode failed: %w", err))
		}
	} else {
		log.Println("Krkn-ai dry mode finished")
	}

	log.Println("krkn-ai execution completed")
	return nil
}

// handleExecutionError sets the failure state and returns the error
func (k *KrknAI) handleExecutionError(err error) error {
	k.result.ExitCode = config.Failure
	viper.Set(config.Cluster.Passing, false)
	return err
}

// runKrknContainer executes the Krkn-ai container using podman or docker with the specified mode.
func (k *KrknAI) runKrknContainer(ctx context.Context, mode string) error {
	runtime, err := detectContainerRuntime()
	if err != nil {
		return err
	}

	// Build base container arguments (common to both modes)
	args := []string{"run", "--rm", "--net=host"}

	// Add volume mounts
	args = append(args,
		"-v", fmt.Sprintf("%s:%s:Z", viper.GetString(config.SharedDir), containerMountPath),
		"-v", fmt.Sprintf("%s:%s:Z", viper.GetString(config.ReportDir), containerResultsPath),
	)

	// Add common environment variables
	args = append(args,
		"-e", fmt.Sprintf("MODE=%s", mode),
		"-e", fmt.Sprintf("KUBECONFIG=%s/%s", containerMountPath, kubeconfigFileName),
		"-e", fmt.Sprintf("VERBOSE=%s", config.KrknAIVerboseLevel),
	)

	// Add mode-specific flags and environment variables
	if mode == config.KrknAIModeRun {
		// Run mode: privileged flag, config file, results output, and Prometheus token
		args = append(args, "--privileged")
		args = append(args,
			"-e", fmt.Sprintf("CONFIG_FILE=%s/%s", containerMountPath, krknConfigFileName),
			"-e", fmt.Sprintf("OUTPUT_DIR=%s", containerResultsPath),
		)

		// Fetch Prometheus token from cluster
		log.Println("Fetching Prometheus token from cluster")
		promToken, err := k.getPrometheusToken(ctx)
		if err != nil {
			log.Printf("Warning - failed to fetch Prometheus token: %v", err)
			log.Println("Continuing without Prometheus token")
		} else {
			args = append(args, "-e", fmt.Sprintf("PROMETHEUS_TOKEN=%s", promToken))
		}
	} else {
		// Discover mode: namespace/pod/node targeting
		args = append(args,
			"-e", fmt.Sprintf("OUTPUT_DIR=%s", containerMountPath),
			"-e", fmt.Sprintf("NAMESPACE=%s", viper.GetString(config.KrknAI.Namespace)),
			"-e", fmt.Sprintf("POD_LABEL=%s", viper.GetString(config.KrknAI.PodLabel)),
		)

		if nodeLabel := viper.GetString(config.KrknAI.NodeLabel); nodeLabel != "" {
			args = append(args, "-e", fmt.Sprintf("NODE_LABEL=%s", nodeLabel))
		}
		if skipPodName := viper.GetString(config.KrknAI.SkipPodName); skipPodName != "" {
			args = append(args, "-e", fmt.Sprintf("SKIP_POD_NAME=%s", skipPodName))
		}
	}

	// Add the image name
	args = append(args, DefaultKrknAIImage)

	log.Printf("Executing command: %s %v", runtime, args)

	cmd := exec.CommandContext(ctx, runtime, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("container execution failed: %w", err)
	}

	log.Printf("Container output:\n%s", stdout.String())
	if stderr.Len() > 0 {
		log.Printf("Container stderr:\n%s", stderr.String())
	}

	return nil
}

// getPrometheusToken retrieves a token for the prometheus-k8s service account from the cluster.
func (k *KrknAI) getPrometheusToken(ctx context.Context) (string, error) {
	// Get kubeconfig from shared dir
	sharedDir := viper.GetString(config.SharedDir)
	kubeconfigPath := filepath.Join(sharedDir, kubeconfigFileName)

	// Create openshift client from kubeconfig
	client, err := openshift.NewFromKubeconfig(kubeconfigPath, logr.Discard())
	if err != nil {
		return "", fmt.Errorf("failed to create openshift client: %w", err)
	}

	// Use osde2e-common prometheus package to create the token
	return prometheus.GetPrometheusToken(ctx, client)
}

// updateKrknConfig updates the Krkn-ai output YAML with values from viper config.
func (k *KrknAI) updateKrknConfig() error {
	sharedDir := viper.GetString(config.SharedDir)
	fitnessQuery := viper.GetString(config.KrknAI.FitnessQuery)
	scenarios := viper.GetString(config.KrknAI.Scenarios)

	// Skip if no config values to update
	if fitnessQuery == "" && scenarios == "" {
		return nil
	}

	// Find YAML file in the shared directory
	yamlFile := filepath.Join(sharedDir, krknConfigFileName)
	if _, err := os.Stat(yamlFile); os.IsNotExist(err) {
		return fmt.Errorf("no file named %s found in %s", krknConfigFileName, sharedDir)
	}

	// Read the YAML file
	data, err := os.ReadFile(yamlFile)
	if err != nil {
		return fmt.Errorf("failed to read Krkn-ai config file: %w", err)
	}

	// Parse YAML into a map
	var cfg map[string]interface{}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse Krkn-ai config file: %w", err)
	}

	// Update fitness_function.query if set
	if fitnessQuery != "" {
		if ff, ok := cfg["fitness_function"].(map[string]interface{}); ok {
			ff["query"] = fitnessQuery
			log.Printf("Updated fitness_function.query to: %s", fitnessQuery)
		}
	}

	// Update scenarios if set
	// If the user has set a list of scenarios, enable all of them
	// TODO: Add a way to disable scenarios not selected by user
	if scenarios != "" {
		enabledScenarios := make(map[string]bool)
		for _, s := range strings.Split(scenarios, ",") {
			enabledScenarios[strings.TrimSpace(s)] = true
		}

		if scenarioCfg, ok := cfg["scenario"].(map[string]interface{}); ok {
			for name, val := range scenarioCfg {
				if scenarioMap, ok := val.(map[string]interface{}); ok {
					scenarioMap["enable"] = enabledScenarios[name]
				}
			}
			log.Printf("Updated scenarios: %v", scenarios)
		}
	}

	// Write updated YAML back
	updatedData, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal updated config: %w", err)
	}

	if err := os.WriteFile(yamlFile, updatedData, 0o644); err != nil {
		return fmt.Errorf("failed to write updated config: %w", err)
	}

	log.Printf("Config file updated: %s", yamlFile)
	return nil
}

// detectContainerRuntime finds an available container runtime (podman or docker).
func detectContainerRuntime() (string, error) {
	// Check for podman first
	if path, err := exec.LookPath("podman"); err == nil {
		return path, nil
	}

	// Fall back to docker
	if path, err := exec.LookPath("docker"); err == nil {
		return path, nil
	}

	return "", fmt.Errorf("no container runtime found: install podman or docker")
}

// AnalyzeLogs performs AI-powered log analysis when tests fail,
// providing insights into failure root causes.
func (k *KrknAI) AnalyzeLogs(ctx context.Context, testErr error) error {
	log.Println("Analyzing logs for failure insights")

	// TODO: Implement Kraken AI-specific log analysis
	// This could include:
	// - Correlating chaos events with failures
	// - AI-powered root cause analysis
	// - Generating remediation suggestions

	log.Printf("Log analysis completed for error: %v", testErr)
	return nil
}

// Report generates test reports and collects diagnostic data.
func (k *KrknAI) Report(ctx context.Context) error {
	log.Println("Generating test reports")

	// TODO: Implement chaos test reporting
	// This should include:
	// - Chaos experiment results
	// - Cluster resilience metrics
	// - Recovery time statistics

	log.Println("Report generation completed")
	return nil
}

// Cleanup performs post-test cleanup including resource cleanup and
// optionally destroys the cluster based on configuration.
func (k *KrknAI) Cleanup(ctx context.Context) error {
	log.Println("Starting cleanup")

	// Delete cluster if configured
	if err := cluster.DeleteCluster(k.provider); err != nil {
		k.result.Errors = append(k.result.Errors, err)
		return fmt.Errorf("failed to delete cluster: %w", err)
	}

	log.Println("Cleanup completed")
	return nil
}

// PostProcessCluster performs optional post-processing on the cluster
// after test execution but before cleanup.
func (k *KrknAI) PostProcessCluster(ctx context.Context) error {
	// TODO: Implement post-processing logic
	// This could include:
	// - Collecting chaos experiment artifacts
	// - Updating cluster metadata
	// - Extending cluster expiration if needed

	return nil
}

// Result returns the outcome of the test run including exit code and status.
func (k *KrknAI) Result() *orchestrator.Result {
	return k.result
}
