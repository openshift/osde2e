package aggregator

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	internalAggregator "github.com/openshift/osde2e/internal/aggregator"
	"gopkg.in/yaml.v3"
)

const (
	// Default file paths relative to results directory
	allCSVPath               = "reports/all.csv"
	healthCheckReportCSVPath = "reports/health_check_report.csv"
	configYAMLPath           = "krkn-ai.yaml"

	// Top scenarios to include in summary
	defaultTopScenariosCount = 10
)

// KrknAIAggregator collects and parses krkn-ai chaos test results.
type KrknAIAggregator struct {
	logger            logr.Logger
	topScenariosCount int
}

// KrknAIData holds aggregated krkn-ai results with minimal context.
type KrknAIData struct {
	Summary           KrknAISummary                 `json:"summary"`
	TopScenarios      []ScenarioResult              `json:"topScenarios"`
	FailedScenarios   []ScenarioResult              `json:"failedScenarios"`
	HealthCheckReport []HealthCheckResult           `json:"healthCheckReport"`
	LogArtifacts      []internalAggregator.LogEntry `json:"logArtifacts"`
	ConfigSummary     string                        `json:"configSummary,omitempty"`
}

// KrknAISummary provides high-level statistics about the chaos test run.
type KrknAISummary struct {
	TotalScenarioCount      int      `json:"totalScenarioCount"`
	SuccessfulScenarioCount int      `json:"successfulScenarioCount"`
	FailedScenarioCount     int      `json:"failedScenarioCount"`
	Generations             int      `json:"generations"`
	MaxFitnessScore         float64  `json:"maxFitnessScore"`
	AvgFitnessScore         float64  `json:"avgFitnessScore"`
	ScenarioTypes           []string `json:"scenarioTypes"`
}

// ScenarioResult represents a single chaos scenario execution result.
type ScenarioResult struct {
	GenerationID                 int     `json:"generationId"`
	ScenarioID                   int     `json:"scenarioId"`
	Scenario                     string  `json:"scenario"`
	Parameters                   string  `json:"parameters"`
	HealthCheckFailureScore      float64 `json:"healthCheckFailureScore"`
	HealthCheckResponseTimeScore float64 `json:"healthCheckResponseTimeScore"`
	KrknFailureScore             float64 `json:"krknFailureScore"`
	FitnessScore                 float64 `json:"fitnessScore"`
}

// HealthCheckResult represents health check metrics for a scenario.
type HealthCheckResult struct {
	ScenarioID          int     `json:"scenarioId"`
	ComponentName       string  `json:"componentName"`
	MinResponseTime     float64 `json:"minResponseTime"`
	MaxResponseTime     float64 `json:"maxResponseTime"`
	AverageResponseTime float64 `json:"averageResponseTime"`
	SuccessCount        int     `json:"successCount"`
	FailureCount        int     `json:"failureCount"`
}

// NewKrknAIAggregator creates a new aggregator for krkn-ai results.
func NewKrknAIAggregator(ctx context.Context) *KrknAIAggregator {
	return &KrknAIAggregator{
		logger:            logr.FromContextOrDiscard(ctx),
		topScenariosCount: defaultTopScenariosCount,
	}
}

// WithTopScenariosCount sets the number of top scenarios to include.
func (a *KrknAIAggregator) WithTopScenariosCount(count int) *KrknAIAggregator {
	a.topScenariosCount = count
	return a
}

// Collect gathers krkn-ai results from the specified directory.
func (a *KrknAIAggregator) Collect(ctx context.Context, resultsDir string) (*KrknAIData, error) {
	a.logger.Info("collecting krkn-ai results", "resultsDir", resultsDir)

	if _, err := os.Stat(resultsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("results directory does not exist: %s", resultsDir)
	}

	data := &KrknAIData{}
	var collectionErrors []string

	// Collect scenario results from all.csv
	scenarios, err := a.collectScenarioResults(resultsDir)
	if err != nil {
		errMsg := fmt.Sprintf("failed to collect scenario results: %v", err)
		a.logger.Error(err, "failed to collect scenario results")
		collectionErrors = append(collectionErrors, errMsg)
	} else {
		a.processScenarios(data, scenarios)
	}

	// Collect health check report
	if err := a.collectHealthCheckReport(resultsDir, data); err != nil {
		errMsg := fmt.Sprintf("failed to collect health check report: %v", err)
		a.logger.Error(err, "failed to collect health check report")
		collectionErrors = append(collectionErrors, errMsg)
	}

	// Collect config summary
	if err := a.collectConfigSummary(resultsDir, data); err != nil {
		a.logger.Info("config file not found or unreadable", "error", err)
		// Not critical - continue without config
	}

	// Collect log artifacts for LLM tool access
	if err := a.collectLogArtifacts(resultsDir, data); err != nil {
		errMsg := fmt.Sprintf("failed to collect log artifacts: %v", err)
		a.logger.Error(err, "failed to collect log artifacts")
		collectionErrors = append(collectionErrors, errMsg)
	}

	a.logger.Info("completed krkn-ai artifact collection",
		"totalScenarios", data.Summary.TotalScenarioCount,
		"failedScenarios", data.Summary.FailedScenarioCount,
		"topScenarios", len(data.TopScenarios),
		"errors", len(collectionErrors))

	return data, nil
}

// collectScenarioResults parses all.csv and returns scenario results.
func (a *KrknAIAggregator) collectScenarioResults(resultsDir string) ([]ScenarioResult, error) {
	csvPath := filepath.Join(resultsDir, allCSVPath)
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", allCSVPath, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file is empty or has no data rows")
	}

	// Skip header row
	var scenarios []ScenarioResult
	for i, record := range records[1:] {
		if len(record) < 8 {
			a.logger.Info("skipping malformed row", "row", i+2, "columns", len(record))
			continue
		}

		scenario, err := a.parseScenarioRecord(record)
		if err != nil {
			a.logger.Info("failed to parse row", "row", i+2, "error", err)
			continue
		}
		scenarios = append(scenarios, scenario)
	}

	return scenarios, nil
}

// parseScenarioRecord parses a single CSV row into ScenarioResult.
func (a *KrknAIAggregator) parseScenarioRecord(record []string) (ScenarioResult, error) {
	generationID, err := strconv.Atoi(record[0])
	if err != nil {
		return ScenarioResult{}, fmt.Errorf("invalid generation_id: %w", err)
	}

	scenarioID, err := strconv.Atoi(record[1])
	if err != nil {
		return ScenarioResult{}, fmt.Errorf("invalid scenario_id: %w", err)
	}

	healthCheckFailureScore, _ := strconv.ParseFloat(record[4], 64)
	healthCheckResponseTimeScore, _ := strconv.ParseFloat(record[5], 64)
	krknFailureScore, _ := strconv.ParseFloat(record[6], 64)
	fitnessScore, _ := strconv.ParseFloat(record[7], 64)

	return ScenarioResult{
		GenerationID:                 generationID,
		ScenarioID:                   scenarioID,
		Scenario:                     record[2],
		Parameters:                   record[3],
		HealthCheckFailureScore:      healthCheckFailureScore,
		HealthCheckResponseTimeScore: healthCheckResponseTimeScore,
		KrknFailureScore:             krknFailureScore,
		FitnessScore:                 fitnessScore,
	}, nil
}

// processScenarios analyzes scenarios and populates summary, top, and failed lists.
func (a *KrknAIAggregator) processScenarios(data *KrknAIData, scenarios []ScenarioResult) {
	if len(scenarios) == 0 {
		return
	}

	// Calculate summary statistics
	var totalFitness float64
	maxGen := 0
	scenarioTypes := make(map[string]struct{})
	var failed []ScenarioResult

	for _, s := range scenarios {
		if s.GenerationID > maxGen {
			maxGen = s.GenerationID
		}
		scenarioTypes[s.Scenario] = struct{}{}

		// KrknFailureScore of -1 indicates scenario failure
		if s.KrknFailureScore < 0 {
			failed = append(failed, s)
		} else {
			totalFitness += s.FitnessScore
		}
	}

	successCount := len(scenarios) - len(failed)

	// Build scenario types list
	types := make([]string, 0, len(scenarioTypes))
	for t := range scenarioTypes {
		types = append(types, t)
	}
	sort.Strings(types)

	// Sort by fitness score descending to get top scenarios
	sorted := make([]ScenarioResult, len(scenarios))
	copy(sorted, scenarios)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].FitnessScore > sorted[j].FitnessScore
	})

	// Get top N scenarios (excluding failed ones)
	var topScenarios []ScenarioResult
	for _, s := range sorted {
		if s.KrknFailureScore >= 0 && len(topScenarios) < a.topScenariosCount {
			topScenarios = append(topScenarios, s)
		}
	}

	// Calculate max and average fitness (excluding failed)
	var maxFitness, avgFitness float64
	if successCount > 0 {
		avgFitness = totalFitness / float64(successCount)
		if len(topScenarios) > 0 {
			maxFitness = topScenarios[0].FitnessScore
		}
	}

	data.Summary = KrknAISummary{
		TotalScenarioCount:      len(scenarios),
		SuccessfulScenarioCount: successCount,
		FailedScenarioCount:     len(failed),
		Generations:             maxGen + 1, // 0-indexed
		MaxFitnessScore:         maxFitness,
		AvgFitnessScore:         avgFitness,
		ScenarioTypes:           types,
	}
	data.TopScenarios = topScenarios
	data.FailedScenarios = failed
}

// collectHealthCheckReport parses health_check_report.csv.
func (a *KrknAIAggregator) collectHealthCheckReport(resultsDir string, data *KrknAIData) error {
	csvPath := filepath.Join(resultsDir, healthCheckReportCSVPath)
	file, err := os.Open(csvPath)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", healthCheckReportCSVPath, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) < 2 {
		return nil // Empty file is OK
	}

	// Skip header row
	for i, record := range records[1:] {
		if len(record) < 7 {
			a.logger.Info("skipping malformed health check row", "row", i+2)
			continue
		}

		result, err := a.parseHealthCheckRecord(record)
		if err != nil {
			a.logger.Info("failed to parse health check row", "row", i+2, "error", err)
			continue
		}
		data.HealthCheckReport = append(data.HealthCheckReport, result)
	}

	return nil
}

// parseHealthCheckRecord parses a single health check CSV row.
func (a *KrknAIAggregator) parseHealthCheckRecord(record []string) (HealthCheckResult, error) {
	scenarioID, err := strconv.Atoi(record[0])
	if err != nil {
		return HealthCheckResult{}, fmt.Errorf("invalid scenario_id: %w", err)
	}

	minRT, _ := strconv.ParseFloat(record[2], 64)
	maxRT, _ := strconv.ParseFloat(record[3], 64)
	avgRT, _ := strconv.ParseFloat(record[4], 64)
	successCount, _ := strconv.Atoi(record[5])
	failureCount, _ := strconv.Atoi(record[6])

	return HealthCheckResult{
		ScenarioID:          scenarioID,
		ComponentName:       record[1],
		MinResponseTime:     minRT,
		MaxResponseTime:     maxRT,
		AverageResponseTime: avgRT,
		SuccessCount:        successCount,
		FailureCount:        failureCount,
	}, nil
}

// collectConfigSummary parses krkn-ai.yaml and extracts relevant config sections.
func (a *KrknAIAggregator) collectConfigSummary(resultsDir string, data *KrknAIData) error {
	configPath := filepath.Join(resultsDir, configYAMLPath)
	content, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var cfg map[string]interface{}
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return fmt.Errorf("failed to parse krkn-ai config: %w", err)
	}

	data.ConfigSummary = formatConfigSummary(cfg)
	return nil
}

// formatConfigSummary extracts key sections from config, excluding verbose cluster_components.
func formatConfigSummary(cfg map[string]interface{}) string {
	var sb strings.Builder

	// GA parameters
	sb.WriteString("=== Genetic Algorithm Parameters ===\n")
	for _, key := range []string{"generations", "population_size", "wait_duration", "mutation_rate", "scenario_mutation_rate", "crossover_rate"} {
		if v, ok := cfg[key]; ok {
			sb.WriteString(fmt.Sprintf("%s: %v\n", key, v))
		}
	}

	// Fitness function
	if ff, ok := cfg["fitness_function"].(map[string]interface{}); ok {
		sb.WriteString("\n=== Fitness Function ===\n")
		for k, v := range ff {
			sb.WriteString(fmt.Sprintf("%s: %v\n", k, v))
		}
	}

	// Enabled scenarios
	if scenarios, ok := cfg["scenario"].(map[string]interface{}); ok {
		sb.WriteString("\n=== Enabled Chaos Scenarios ===\n")
		var enabled []string
		for name, v := range scenarios {
			if m, ok := v.(map[string]interface{}); ok && m["enable"] == true {
				enabled = append(enabled, name)
			}
		}
		sort.Strings(enabled)
		sb.WriteString(strings.Join(enabled, ", ") + "\n")
	}

	// Health check targets (just names/URLs, not full config)
	if hc, ok := cfg["health_checks"].(map[string]interface{}); ok {
		if apps, ok := hc["applications"].([]interface{}); ok && len(apps) > 0 {
			sb.WriteString("\n=== Health Check Targets ===\n")
			for _, app := range apps {
				if m, ok := app.(map[string]interface{}); ok {
					sb.WriteString(fmt.Sprintf("- %v: %v\n", m["name"], m["url"]))
				}
			}
		}
	}

	return sb.String()
}

// collectLogArtifacts walks the results directory and catalogs available files.
func (a *KrknAIAggregator) collectLogArtifacts(resultsDir string, data *KrknAIData) error {
	// Get absolute path for the results directory
	absResultsDir, err := filepath.Abs(resultsDir)
	if err != nil {
		absResultsDir = resultsDir
	}

	return filepath.Walk(absResultsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue on error
		}

		// Skip directories and hidden files
		if info.IsDir() || strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(info.Name()))
		// Skip PNG files (not useful for text analysis) and CSV files (already parsed)
		if ext == ".png" || ext == ".csv" {
			return nil
		}

		lineCount := 0
		if content, err := os.ReadFile(path); err == nil {
			lineCount = strings.Count(string(content), "\n")
			if len(content) > 0 && !strings.HasSuffix(string(content), "\n") {
				lineCount++
			}
		}

		// Use absolute path so read_file tool can find the file
		data.LogArtifacts = append(data.LogArtifacts, internalAggregator.LogEntry{
			Source:    path,
			LineCount: lineCount,
		})

		return nil
	})
}
