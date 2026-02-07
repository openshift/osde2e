package aggregator

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKrknAIAggregator_Collect(t *testing.T) {
	tempDir := t.TempDir()
	resultsDir := filepath.Join(tempDir, "results")
	reportsDir := filepath.Join(resultsDir, "reports")
	require.NoError(t, os.MkdirAll(reportsDir, 0o755))

	// Create test CSV files
	createKrknAITestFiles(t, resultsDir, reportsDir)

	ctx := context.Background()
	agg := NewKrknAIAggregator(ctx)
	data, err := agg.Collect(ctx, resultsDir)

	require.NoError(t, err)
	require.NotNil(t, data)

	// Verify summary
	assert.Equal(t, 5, data.Summary.TotalScenarios)
	assert.Equal(t, 4, data.Summary.SuccessfulScenarios)
	assert.Equal(t, 1, data.Summary.FailedScenarios)
	assert.Equal(t, 3, data.Summary.Generations) // 0, 1, 2

	// Verify top scenarios (sorted by fitness descending)
	assert.GreaterOrEqual(t, len(data.TopScenarios), 1)
	assert.Equal(t, 2.2, data.TopScenarios[0].FitnessScore)
	assert.Equal(t, "node-cpu-hog", data.TopScenarios[0].Scenario)

	// Verify failed scenarios
	assert.Equal(t, 1, len(data.FailedScenarios))
	assert.Equal(t, -1.0, data.FailedScenarios[0].KrknFailureScore)

	// Verify health check report
	assert.Greater(t, len(data.HealthCheckReport), 0)
	assert.Equal(t, "console", data.HealthCheckReport[0].ComponentName)

	// Verify config summary is populated with key sections
	assert.NotEmpty(t, data.ConfigSummary)
	assert.Contains(t, data.ConfigSummary, "=== Genetic Algorithm Parameters ===")
	assert.Contains(t, data.ConfigSummary, "generations: 20")
	assert.Contains(t, data.ConfigSummary, "=== Chaos Scenarios ===")
	assert.Contains(t, data.ConfigSummary, "node_cpu_hog")
}

func TestKrknAIAggregator_NonExistentDirectory(t *testing.T) {
	ctx := context.Background()
	agg := NewKrknAIAggregator(ctx)
	_, err := agg.Collect(ctx, "/non/existent/path")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "results directory does not exist")
}

func TestKrknAIAggregator_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()
	reportsDir := filepath.Join(tempDir, "reports")
	require.NoError(t, os.MkdirAll(reportsDir, 0o755))

	ctx := context.Background()
	agg := NewKrknAIAggregator(ctx)

	// Should not error on missing files, just return empty data
	data, err := agg.Collect(ctx, tempDir)

	require.NoError(t, err)
	require.NotNil(t, data)
	assert.Equal(t, 0, data.Summary.TotalScenarios)
}

func TestKrknAIAggregator_WithTopScenariosCount(t *testing.T) {
	tempDir := t.TempDir()
	resultsDir := filepath.Join(tempDir, "results")
	reportsDir := filepath.Join(resultsDir, "reports")
	require.NoError(t, os.MkdirAll(reportsDir, 0o755))

	createKrknAITestFiles(t, resultsDir, reportsDir)

	ctx := context.Background()
	agg := NewKrknAIAggregator(ctx).WithTopScenariosCount(2)
	data, err := agg.Collect(ctx, resultsDir)

	require.NoError(t, err)
	assert.LessOrEqual(t, len(data.TopScenarios), 2)
}

func TestKrknAIAggregator_SkipsPNGFiles(t *testing.T) {
	tempDir := t.TempDir()
	resultsDir := filepath.Join(tempDir, "results")
	reportsDir := filepath.Join(resultsDir, "reports")
	graphsDir := filepath.Join(reportsDir, "graphs")
	require.NoError(t, os.MkdirAll(graphsDir, 0o755))

	createKrknAITestFiles(t, resultsDir, reportsDir)

	// Create PNG files that should be skipped
	require.NoError(t, os.WriteFile(filepath.Join(graphsDir, "scenario_1.png"), []byte("fake png"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(graphsDir, "best_generation.PNG"), []byte("fake png"), 0o644))

	ctx := context.Background()
	agg := NewKrknAIAggregator(ctx)
	data, err := agg.Collect(ctx, resultsDir)

	require.NoError(t, err)

	// Verify PNG files are not in log artifacts
	for _, artifact := range data.LogArtifacts {
		assert.NotContains(t, artifact.Source, ".png", "PNG files should be skipped")
		assert.NotContains(t, artifact.Source, ".PNG", "PNG files should be skipped")
	}
}

func TestKrknAIAggregator_ParseScenarioResult(t *testing.T) {
	ctx := context.Background()
	agg := NewKrknAIAggregator(ctx)

	testCases := []struct {
		name     string
		record   []string
		expected ScenarioResult
		wantErr  bool
	}{
		{
			name: "valid record",
			record: []string{
				"6", "61", "node-cpu-hog",
				"chaos-duration=60 cpu-percentage=61",
				"0.0", "1.2", "0.0", "2.2",
			},
			expected: ScenarioResult{
				GenerationID:                 6,
				ScenarioID:                   61,
				Scenario:                     "node-cpu-hog",
				Parameters:                   "chaos-duration=60 cpu-percentage=61",
				HealthCheckFailureScore:      0.0,
				HealthCheckResponseTimeScore: 1.2,
				KrknFailureScore:             0.0,
				FitnessScore:                 2.2,
			},
			wantErr: false,
		},
		{
			name: "failed scenario with negative score",
			record: []string{
				"0", "6", "pod-scenarios",
				"namespace=openshift-monitoring",
				"0.0", "0.0", "-1.0", "-1.0",
			},
			expected: ScenarioResult{
				GenerationID:     0,
				ScenarioID:       6,
				Scenario:         "pod-scenarios",
				Parameters:       "namespace=openshift-monitoring",
				KrknFailureScore: -1.0,
				FitnessScore:     -1.0,
			},
			wantErr: false,
		},
		{
			name:    "invalid generation_id",
			record:  []string{"abc", "61", "node-cpu-hog", "params", "0.0", "1.2", "0.0", "2.2"},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := agg.parseScenarioRecord(tc.record)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expected.GenerationID, result.GenerationID)
			assert.Equal(t, tc.expected.ScenarioID, result.ScenarioID)
			assert.Equal(t, tc.expected.Scenario, result.Scenario)
			assert.Equal(t, tc.expected.FitnessScore, result.FitnessScore)
			assert.Equal(t, tc.expected.KrknFailureScore, result.KrknFailureScore)
		})
	}
}

func TestKrknAIAggregator_ProcessScenarios(t *testing.T) {
	ctx := context.Background()
	agg := NewKrknAIAggregator(ctx).WithTopScenariosCount(3)

	scenarios := []ScenarioResult{
		{GenerationID: 0, ScenarioID: 1, Scenario: "node-cpu-hog", FitnessScore: 2.0, KrknFailureScore: 0},
		{GenerationID: 0, ScenarioID: 2, Scenario: "node-memory-hog", FitnessScore: 1.5, KrknFailureScore: 0},
		{GenerationID: 1, ScenarioID: 3, Scenario: "pod-scenarios", FitnessScore: -1.0, KrknFailureScore: -1.0},
		{GenerationID: 1, ScenarioID: 4, Scenario: "node-io-hog", FitnessScore: 1.8, KrknFailureScore: 0},
		{GenerationID: 2, ScenarioID: 5, Scenario: "node-cpu-hog", FitnessScore: 2.2, KrknFailureScore: 0},
	}

	data := &KrknAIData{}
	agg.processScenarios(data, scenarios)

	// Verify summary
	assert.Equal(t, 5, data.Summary.TotalScenarios)
	assert.Equal(t, 4, data.Summary.SuccessfulScenarios)
	assert.Equal(t, 1, data.Summary.FailedScenarios)
	assert.Equal(t, 3, data.Summary.Generations)
	assert.Equal(t, 2.2, data.Summary.MaxFitnessScore)

	// Verify top scenarios are sorted by fitness descending
	require.Equal(t, 3, len(data.TopScenarios))
	assert.Equal(t, 2.2, data.TopScenarios[0].FitnessScore)
	assert.Equal(t, 2.0, data.TopScenarios[1].FitnessScore)
	assert.Equal(t, 1.8, data.TopScenarios[2].FitnessScore)

	// Verify failed scenarios
	assert.Equal(t, 1, len(data.FailedScenarios))
	assert.Equal(t, "pod-scenarios", data.FailedScenarios[0].Scenario)

	// Verify scenario types
	assert.Contains(t, data.Summary.ScenarioTypes, "node-cpu-hog")
	assert.Contains(t, data.Summary.ScenarioTypes, "node-memory-hog")
	assert.Contains(t, data.Summary.ScenarioTypes, "node-io-hog")
	assert.Contains(t, data.Summary.ScenarioTypes, "pod-scenarios")
}

func TestKrknAIAggregator_ConfigSummaryExtractsCorrectSections(t *testing.T) {
	tempDir := t.TempDir()
	resultsDir := filepath.Join(tempDir, "results")
	require.NoError(t, os.MkdirAll(resultsDir, 0o755))

	// Create config with keys in different order than expected (simulating updateKrknConfig reordering)
	configYAML := `scenario:
  network_scenarios:
    enable: false
  pod_scenarios:
    enable: true
health_checks:
  applications:
  - name: api
    url: https://api.example.com/health
generations: 15
fitness_function:
  query: sum(probe_success)
  type: range
  include_krkn_failure: true
population_size: 8
mutation_rate: 0.75
scenario_mutation_rate: 0.5
crossover_rate: 0.6
wait_duration: 90`

	require.NoError(t, os.WriteFile(filepath.Join(resultsDir, "krkn-ai.yaml"), []byte(configYAML), 0o644))

	ctx := context.Background()
	agg := NewKrknAIAggregator(ctx)
	data := &KrknAIData{}

	err := agg.collectConfigSummary(resultsDir, data)
	require.NoError(t, err)

	// Verify key sections extracted regardless of YAML key order
	assert.Contains(t, data.ConfigSummary, "generations: 15")
	assert.Contains(t, data.ConfigSummary, "population_size: 8")
	assert.Contains(t, data.ConfigSummary, "mutation_rate: 0.75")
	assert.Contains(t, data.ConfigSummary, "type: range")
	assert.Contains(t, data.ConfigSummary, "query: sum(probe_success)")
	assert.Contains(t, data.ConfigSummary, "Enabled: pod_scenarios")
	assert.Contains(t, data.ConfigSummary, "Disabled: network_scenarios")
	assert.Contains(t, data.ConfigSummary, "- api: https://api.example.com/health")
}

func createKrknAITestFiles(t *testing.T, resultsDir, reportsDir string) {
	// Create all.csv with sample data
	allCSV := `generation_id,scenario_id,scenario,parameters,health_check_failure_score,health_check_response_time_score,krkn_failure_score,fitness_score
0,1,node-cpu-hog,"chaos-duration=60 cpu-percentage=61",0.0,1.2,0.0,2.2
0,2,node-memory-hog,"chaos-duration=60 memory-consumption=49%",0.0,1.0,0.0,2.0
1,3,node-io-hog,"chaos-duration=60 io-block-size=3m",0.0,0.8,0.0,1.8
1,4,pod-scenarios,"namespace=openshift-monitoring",0.0,0.5,0.0,1.5
2,5,dns-outage,"chaos-duration=60 pod-name=test",0.0,0.0,-1.0,-1.0`

	require.NoError(t, os.WriteFile(filepath.Join(reportsDir, "all.csv"), []byte(allCSV), 0o644))

	// Create health_check_report.csv
	healthCSV := `scenario_id,component_name,min_response_time,max_response_time,average_response_time,success_count,failure_count
1,console,0.065,0.400,0.088,100,0
2,console,0.064,0.280,0.087,101,0
3,console,0.063,0.309,0.089,103,0
4,console,0.065,0.399,0.088,100,0
5,console,0.062,0.215,0.078,77,5`

	require.NoError(t, os.WriteFile(filepath.Join(reportsDir, "health_check_report.csv"), []byte(healthCSV), 0o644))

	// Create krkn-ai.yaml config
	configYAML := `kubeconfig_file_path: ./tmp/kubeconfig.yaml
generations: 20
population_size: 10
wait_duration: 120
fitness_function:
  query: sum(probe_success)
  type: range
scenario:
  pod_scenarios:
    enable: true
  node_cpu_hog:
    enable: true
  node_memory_hog:
    enable: true`

	require.NoError(t, os.WriteFile(filepath.Join(resultsDir, "krkn-ai.yaml"), []byte(configYAML), 0o644))
}
