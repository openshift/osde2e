package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/openshift/osde2e/internal/aggregator"
	"google.golang.org/genai"
)

type getArtifactsTool struct{}

func (t *getArtifactsTool) name() string {
	return "get_artifacts"
}

func (t *getArtifactsTool) description() string {
	return "Collects and returns specific artifacts and metadata from the current osde2e test run. Supports filtering by artifact type: metadata, failed_tests, build_logs, cluster_logs, mustgather, test_results, or all"
}

func (t *getArtifactsTool) schema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"artifact_type": {
				Type:        genai.TypeString,
				Description: "The type of artifacts to return. Options: metadata, failed_tests, build_logs, cluster_logs, mustgather, test_results, all",
				Enum:        []string{"metadata", "failed_tests", "build_logs", "cluster_logs", "mustgather", "test_results", "all"},
			},
		},
		Required: []string{"artifact_type"},
	}
}

func (t *getArtifactsTool) execute(ctx context.Context, params map[string]any) (any, error) {

	artifactType, err := extractString(params, "artifact_type")
	if err != nil {
		return nil, err
	}
	fmt.Println("Executing get_artifacts with params: ", params)

	// TODO: provide the report directory here
	reportDir := "ci_artifacts"

	// Create a logger for the aggregator service
	// In a real implementation, this would come from the main application
	logger := logr.Discard()

	// Create the aggregator service
	service := aggregator.NewService(logger)

	// Collect artifacts from the report directory
	data, err := service.Collect(ctx, reportDir)
	if err != nil {
		return nil, fmt.Errorf("failed to collect artifacts from %s: %w", reportDir, err)
	}

	// Filter and return only the requested artifact type
	result, err := t.filterArtifacts(data, artifactType)
	if err != nil {
		return nil, err
	}

	// Convert to JSON for clean serialization
	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize artifacts data: %w", err)
	}

	return string(jsonData), nil
}

func (t *getArtifactsTool) filterArtifacts(data *aggregator.AggregatedData, artifactType string) (any, error) {
	switch artifactType {
	case "metadata":
		return map[string]any{
			"clusterId":      data.Metadata.ClusterID,
			"clusterVersion": data.Metadata.ClusterVersion,
			"provider":       data.Metadata.Provider,
			"cloudProvider":  data.Metadata.CloudProvider,
			"region":         data.Metadata.Region,
			"jobName":        data.Metadata.JobName,
			"jobId":          data.Metadata.JobID,
			"phase":          data.Metadata.Phase,
			"environment":    data.Metadata.Environment,
			"properties":     data.Metadata.Properties,
		}, nil

	case "test_results":
		return map[string]any{
			"totalTests":   data.TestResults.TotalTests,
			"passedTests":  data.TestResults.PassedTests,
			"failedTests":  data.TestResults.FailedTests,
			"skippedTests": data.TestResults.SkippedTests,
			"errorTests":   data.TestResults.ErrorTests,
			"duration":     data.TestResults.Duration.String(),
			"suiteCount":   data.TestResults.SuiteCount,
		}, nil

	case "failed_tests":
		return data.FailedTests, nil

	case "cluster_logs":
		return data.ClusterLogs, nil

	case "mustgather":
		return data.MustGatherData, nil

	case "build_logs":
		return data.BuildLogs, nil

	case "all":
		return map[string]any{
			"metadata": map[string]any{
				"clusterId":      data.Metadata.ClusterID,
				"clusterVersion": data.Metadata.ClusterVersion,
				"provider":       data.Metadata.Provider,
				"cloudProvider":  data.Metadata.CloudProvider,
				"region":         data.Metadata.Region,
				"jobName":        data.Metadata.JobName,
				"jobId":          data.Metadata.JobID,
				"phase":          data.Metadata.Phase,
				"environment":    data.Metadata.Environment,
				"properties":     data.Metadata.Properties,
			},
			"testResults": map[string]any{
				"totalTests":   data.TestResults.TotalTests,
				"passedTests":  data.TestResults.PassedTests,
				"failedTests":  data.TestResults.FailedTests,
				"skippedTests": data.TestResults.SkippedTests,
				"errorTests":   data.TestResults.ErrorTests,
				"duration":     data.TestResults.Duration.String(),
				"suiteCount":   data.TestResults.SuiteCount,
			},
			"failedTests":    data.FailedTests,
			"clusterLogs":    data.ClusterLogs,
			"mustGatherData": data.MustGatherData,
			"buildLogs":      data.BuildLogs,
			"collectionTime": data.CollectionTime.Format("2006-01-02T15:04:05Z07:00"),
		}, nil

	default:
		return nil, fmt.Errorf("unsupported artifact type: %s. Supported types are: metadata, failed_tests, build_logs, cluster_logs, mustgather, test_results, all", artifactType)
	}
}

func extractString(params map[string]any, key string) (string, error) {
	val, ok := params[key]
	if !ok {
		return "", fmt.Errorf("parameter '%s' is required", key)
	}

	str, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("parameter '%s' must be a string, got %T", key, val)
	}

	return str, nil
}
