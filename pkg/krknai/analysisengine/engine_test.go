package analysisengine

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/openshift/osde2e/internal/analysisengine"
	"github.com/openshift/osde2e/internal/llm"
	"github.com/openshift/osde2e/internal/llm/tools"
	"github.com/openshift/osde2e/internal/prompts"
	"github.com/openshift/osde2e/internal/reporter"
	krknAggregator "github.com/openshift/osde2e/pkg/krknai/aggregator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// mockLLMClient implements llm.LLMClient for testing.
type mockLLMClient struct {
	response *llm.AnalysisResult
	err      error
}

func (m *mockLLMClient) Analyze(_ context.Context, _ string, _ *llm.AnalysisConfig, _ *tools.Registry) (*llm.AnalysisResult, error) {
	return m.response, m.err
}

func TestNew_ValidConfig(t *testing.T) {
	// New requires a real Gemini API key to create the client,
	// so we test validation logic only
	ctx := context.Background()

	_, err := New(ctx, &Config{
		BaseConfig: analysisengine.BaseConfig{
			ArtifactsDir: "/some/dir",
		},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GEMINI_API_KEY is required")

	_, err = New(ctx, &Config{
		BaseConfig: analysisengine.BaseConfig{
			APIKey: "fake-key",
		},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "results directory is required")
}

func TestEmbeddedPromptTemplate(t *testing.T) {
	// Verify the embedded prompt template loads correctly
	data, err := krknaiTemplatesFS.ReadFile("prompts/krknai.yaml")
	require.NoError(t, err)
	assert.Contains(t, string(data), "system_prompt")
	assert.Contains(t, string(data), "user_prompt")
	assert.Contains(t, string(data), "chaos engineering")
}

func TestWriteSummary(t *testing.T) {
	tempDir := t.TempDir()

	engine := &Engine{
		config: &Config{
			BaseConfig: analysisengine.BaseConfig{ArtifactsDir: tempDir},
		},
	}

	result := &analysisengine.Result{
		Status:  "completed",
		Content: "Test analysis content",
		Prompt:  "Test prompt",
		Metadata: map[string]any{
			"analysis_type":   "krknai",
			"total_scenarios": 5,
		},
	}

	data := &krknAggregator.KrknAIData{
		Summary: krknAggregator.KrknAISummary{
			TotalScenarioCount:      5,
			SuccessfulScenarioCount: 4,
			FailedScenarioCount:     1,
			Generations:             3,
			MaxFitnessScore:         2.2,
			AvgFitnessScore:         1.8,
			ScenarioTypes:           []string{"node-cpu-hog", "pod-scenarios"},
		},
		TopScenarios: []krknAggregator.ScenarioResult{
			{ScenarioID: 1, Scenario: "node-cpu-hog", FitnessScore: 2.2},
		},
		FailedScenarios: []krknAggregator.ScenarioResult{
			{ScenarioID: 5, Scenario: "dns-outage", KrknFailureScore: -1.0},
		},
	}

	err := engine.writeSummary(result, data)
	require.NoError(t, err)

	// Verify summary file exists
	summaryPath := filepath.Join(tempDir, analysisDirName, summaryFileName)
	_, err = os.Stat(summaryPath)
	require.NoError(t, err)

	// Verify summary content
	content, err := os.ReadFile(summaryPath)
	require.NoError(t, err)

	var summary map[string]any
	require.NoError(t, yaml.Unmarshal(content, &summary))

	assert.Equal(t, "krknai", summary["analysis_type"])
	assert.Equal(t, "completed", summary["status"])
	assert.Equal(t, "Test analysis content", summary["response"])

	runSummary, ok := summary["run_summary"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, 5, runSummary["total_scenarios"])
	assert.Equal(t, 4, runSummary["successful_scenarios"])
	assert.Equal(t, 1, runSummary["failed_scenarios"])
}

func TestRun_WithMockLLM(t *testing.T) {
	// Create temp results directory with test data
	tempDir := t.TempDir()
	reportsDir := filepath.Join(tempDir, "reports")
	require.NoError(t, os.MkdirAll(reportsDir, 0o755))

	createTestResultFiles(t, tempDir, reportsDir)

	ctx := context.Background()

	// Build engine with mock LLM client
	agg := krknAggregator.NewKrknAIAggregator(ctx)

	promptStore := newTestPromptStore(t)

	mockClient := &mockLLMClient{
		response: &llm.AnalysisResult{
			Content: `{"cluster_resilience_assessment": "Good", "recommendations": ["Add more CPU"]}`,
		},
	}

	engine := &Engine{
		config: &Config{
			BaseConfig: analysisengine.BaseConfig{ArtifactsDir: tempDir, APIKey: "fake-key"},
		},
		aggregator:       agg,
		promptStore:      promptStore,
		llmClient:        mockClient,
		reporterRegistry: newTestReporterRegistry(),
	}

	result, err := engine.Run(ctx)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify result
	assert.Equal(t, "completed", result.Status)
	assert.Contains(t, result.Content, "cluster_resilience_assessment")
	assert.NotEmpty(t, result.Prompt)

	// Verify metadata
	assert.Equal(t, "krknai", result.Metadata["analysis_type"])
	assert.Equal(t, 5, result.Metadata["total_scenarios"])
	assert.Equal(t, 4, result.Metadata["successful_scenarios"])
	assert.Equal(t, 1, result.Metadata["failed_scenarios"])

	// Verify summary file was written
	summaryPath := filepath.Join(tempDir, analysisDirName, summaryFileName)
	_, err = os.Stat(summaryPath)
	assert.NoError(t, err)
}

func TestRun_LLMFailure(t *testing.T) {
	tempDir := t.TempDir()
	reportsDir := filepath.Join(tempDir, "reports")
	require.NoError(t, os.MkdirAll(reportsDir, 0o755))

	createTestResultFiles(t, tempDir, reportsDir)

	ctx := context.Background()
	agg := krknAggregator.NewKrknAIAggregator(ctx)
	promptStore := newTestPromptStore(t)

	mockClient := &mockLLMClient{
		err: assert.AnError,
	}

	engine := &Engine{
		config: &Config{
			BaseConfig: analysisengine.BaseConfig{ArtifactsDir: tempDir, APIKey: "fake-key"},
		},
		aggregator:       agg,
		promptStore:      promptStore,
		llmClient:        mockClient,
		reporterRegistry: newTestReporterRegistry(),
	}

	_, err := engine.Run(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "LLM analysis failed")
}

func TestRun_MissingResults(t *testing.T) {
	ctx := context.Background()
	agg := krknAggregator.NewKrknAIAggregator(ctx)
	promptStore := newTestPromptStore(t)

	engine := &Engine{
		config: &Config{
			BaseConfig: analysisengine.BaseConfig{ArtifactsDir: "/nonexistent/path", APIKey: "fake-key"},
		},
		aggregator:       agg,
		promptStore:      promptStore,
		llmClient:        &mockLLMClient{},
		reporterRegistry: newTestReporterRegistry(),
	}

	_, err := engine.Run(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to collect krkn-ai results")
}

// newTestPromptStore creates a prompt store using the embedded krkn-ai templates.
func newTestPromptStore(t *testing.T) *prompts.PromptStore {
	t.Helper()
	templatesFS, err := fs.Sub(krknaiTemplatesFS, "prompts")
	require.NoError(t, err)
	store, err := prompts.NewPromptStore(templatesFS)
	require.NoError(t, err)
	return store
}

// newTestReporterRegistry creates an empty reporter registry for testing.
func newTestReporterRegistry() *reporter.ReporterRegistry {
	return reporter.NewReporterRegistry()
}

func createTestResultFiles(t *testing.T, resultsDir, reportsDir string) {
	allCSV := `generation_id,scenario_id,scenario,parameters,health_check_failure_score,health_check_response_time_score,krkn_failure_score,fitness_score
0,1,node-cpu-hog,"chaos-duration=60 cpu-percentage=61",0.0,1.2,0.0,2.2
0,2,node-memory-hog,"chaos-duration=60 memory-consumption=49%",0.0,1.0,0.0,2.0
1,3,node-io-hog,"chaos-duration=60 io-block-size=3m",0.0,0.8,0.0,1.8
1,4,pod-scenarios,"namespace=openshift-monitoring",0.0,0.5,0.0,1.5
2,5,dns-outage,"chaos-duration=60 pod-name=test",0.0,0.0,-1.0,-1.0`

	require.NoError(t, os.WriteFile(filepath.Join(reportsDir, "all.csv"), []byte(allCSV), 0o644))

	healthCSV := `scenario_id,component_name,min_response_time,max_response_time,average_response_time,success_count,failure_count
1,console,0.065,0.400,0.088,100,0
2,console,0.064,0.280,0.087,101,0`

	require.NoError(t, os.WriteFile(filepath.Join(reportsDir, "health_check_report.csv"), []byte(healthCSV), 0o644))

	configYAML := `generations: 20
population_size: 10
wait_duration: 120
fitness_function:
  query: sum(probe_success)
  type: range
scenario:
  pod_scenarios:
    enable: true
  node_cpu_hog:
    enable: true`

	require.NoError(t, os.WriteFile(filepath.Join(resultsDir, "krkn-ai.yaml"), []byte(configYAML), 0o644))
}
