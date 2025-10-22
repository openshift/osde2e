package tools

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/openshift/osde2e/internal/aggregator"
	"google.golang.org/genai"
)

type mustGatherTool struct {
	omcClient *OMCClient
}

// newMustGatherTool creates a new must-gather tool with the specified tar file
func newMustGatherTool(ctx context.Context, mustGatherTarPath string) (*mustGatherTool, error) {
	client := NewOMCClient()
	if err := client.Initialize(ctx, mustGatherTarPath); err != nil {
		return nil, fmt.Errorf("failed to initialize OMC client: %w", err)
	}

	return &mustGatherTool{
		omcClient: client,
	}, nil
}

func (t *mustGatherTool) Name() string {
	return "must_gather"
}

func (t *mustGatherTool) Description() string {
	return "Check for unhealthy operators in the OpenShift cluster using must-gather data. " +
		"This tool identifies core platform operators and add-on operators that are not in a healthy state. " +
		"Use this to quickly identify operator-related issues that might be causing cluster problems."
}

func (t *mustGatherTool) Schema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"get_operator_health": {
				Type:        genai.TypeString,
				Description: "Scope of operator health check to perform. Options: 'core' (core platform operators), 'addons' (add-on operators), 'all' (both core and add-on operators)",
			},
		},
		Required: []string{"get_operator_health"},
	}
}

func (t *mustGatherTool) Execute(ctx context.Context, params map[string]any, data *aggregator.AggregatedData) (any, error) {
	// Set up context-based cleanup
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start cleanup goroutine that responds to context cancellation
	go func() {
		<-ctx.Done()
		if err := t.Cleanup(); err != nil {
			log.Printf("Warning: Failed to cleanup must-gather tool: %v", err)
		}
	}()

	// Extract get_operator_health parameter
	scope, err := extractString(params, "get_operator_health")
	if err != nil {
		return nil, err
	}

	// Validate scope
	validScopes := []string{"core", "addons", "all"}
	if !containsString(validScopes, scope) {
		return nil, fmt.Errorf("invalid scope '%s'. Must be one of: %v", scope, validScopes)
	}

	var unhealthyOperators []string
	var checkResults []string

	// Check core platform operators
	if scope == "core" || scope == "all" {
		coreOperators, err := t.checkCoreOperators(ctx)
		if err != nil {
			log.Printf("Warning: Failed to check core operators: %v", err)
			checkResults = append(checkResults, fmt.Sprintf("Failed to check core operators: %v", err))
		} else {
			unhealthyOperators = append(unhealthyOperators, coreOperators...)
			checkResults = append(checkResults, "Checked core operators")
		}
	}

	// Check add-on operators
	if scope == "addons" || scope == "all" {
		addonOperators, err := t.checkAddonOperators(ctx)
		if err != nil {
			log.Printf("Warning: Failed to check addon operators: %v", err)
			checkResults = append(checkResults, fmt.Sprintf("Failed to check addon operators: %v", err))
		} else {
			unhealthyOperators = append(unhealthyOperators, addonOperators...)
			checkResults = append(checkResults, "Checked addon operators")
		}
	}

	// Structure the response
	result := map[string]any{
		"scope":               scope,
		"unhealthy_operators": unhealthyOperators,
		"total_unhealthy":     len(unhealthyOperators),
		"check_results":       checkResults,
		"success":             true,
	}

	if len(unhealthyOperators) == 0 {
		result["status"] = "All operators healthy"
	} else {
		result["status"] = fmt.Sprintf("Found %d unhealthy operators", len(unhealthyOperators))
	}

	return result, nil
}

// Cleanup cleans up resources used by the tool
func (t *mustGatherTool) Cleanup() error {
	if t.omcClient != nil {
		return t.omcClient.Cleanup()
	}
	return nil
}

// checkCoreOperators checks for unhealthy core platform operators
func (t *mustGatherTool) checkCoreOperators(ctx context.Context) ([]string, error) {
	// Execute: omc get clusteroperators --no-headers
	output, err := t.omcClient.ExecuteCommand(ctx, "get clusteroperators --no-headers")
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster operators: %w", err)
	}

	var unhealthyOperators []string
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse operator status: NAME VERSION AVAILABLE PROGRESSING DEGRADED SINCE MESSAGE
		// Healthy operators should have: True False False
		fields := strings.Fields(line)
		if len(fields) >= 5 {
			operatorName := fields[0]
			available := fields[2]
			progressing := fields[3]
			degraded := fields[4]

			// Check if operator is unhealthy (not "True False False")
			if available != "True" || progressing != "False" || degraded != "False" {
				status := fmt.Sprintf("%s (Available: %s, Progressing: %s, Degraded: %s)",
					operatorName, available, progressing, degraded)
				unhealthyOperators = append(unhealthyOperators, status)
				log.Printf("Found unhealthy core operator: %s", status)
			}
		}
	}

	return unhealthyOperators, nil
}

// checkAddonOperators checks for unhealthy add-on operators
func (t *mustGatherTool) checkAddonOperators(ctx context.Context) ([]string, error) {
	// First get all subscriptions
	output, err := t.omcClient.ExecuteCommand(ctx, "get subscriptions --all-namespaces --no-headers")
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions: %w", err)
	}

	var unhealthyOperators []string
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse subscription line: NAMESPACE NAME PACKAGE SOURCE CHANNEL
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			namespace := fields[0]
			subscriptionName := fields[1]

			// Check CSV status in this namespace
			csvOutput, err := t.omcClient.ExecuteCommand(ctx, fmt.Sprintf("get csv -n %s --no-headers", namespace))
			if err != nil {
				log.Printf("Warning: Failed to check CSV in namespace %s: %v", namespace, err)
				continue
			}

			// Look for failed CSVs (not "Succeeded")
			csvLines := strings.Split(strings.TrimSpace(csvOutput), "\n")
			for _, csvLine := range csvLines {
				csvLine = strings.TrimSpace(csvLine)
				if csvLine == "" {
					continue
				}

				csvFields := strings.Fields(csvLine)
				if len(csvFields) >= 2 {
					csvName := csvFields[0]
					phase := csvFields[1]

					// Check for actual failure phases, not just "not Succeeded"
					failurePhases := []string{"Failed", "Pending", "Installing", "Replacing", "Deleting"}
					if containsString(failurePhases, phase) {
						status := fmt.Sprintf("%s (%s) - CSV: %s, Phase: %s",
							subscriptionName, namespace, csvName, phase)
						unhealthyOperators = append(unhealthyOperators, status)
						log.Printf("Found unhealthy addon operator: %s", status)
					}
				}
			}
		}
	}

	return unhealthyOperators, nil
}

// containsString checks if a slice contains a string
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// findMustGatherTar searches for must-gather tar files in the log artifacts
func findMustGatherTar(logArtifacts []aggregator.LogEntry) string {
	for _, artifact := range logArtifacts {
		fileName := strings.ToLower(artifact.Source)

		// Look for must-gather tar files
		if strings.Contains(fileName, "must-gather") &&
			(strings.HasSuffix(fileName, ".tar") ||
				strings.HasSuffix(fileName, ".tar.gz") ||
				strings.HasSuffix(fileName, ".tgz")) {
			return artifact.Source
		}

		// Also look for generic must-gather.tar pattern
		if strings.HasSuffix(fileName, "must-gather.tar") ||
			strings.HasSuffix(fileName, "must-gather.tar.gz") {
			return artifact.Source
		}
	}

	return ""
}
