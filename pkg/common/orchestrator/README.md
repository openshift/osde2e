# Orchestrator - Generic Testing Framework Interfaces

This package provides generic, framework-agnostic interfaces for orchestrating end-to-end testing workflows. These interfaces are designed to be reusable across different testing frameworks and can work with any cluster provider or test framework.

## Overview

The orchestrator package defines four core interfaces that separate concerns in a testing workflow, plus a generic `Orchestrator` that coordinates these components:

1. **Provisioner** - Manages cluster infrastructure lifecycle
2. **Executor** - Runs tests or workloads against clusters
3. **Analyzer** - Analyzes test results and failures
4. **Reporter** - Reports results in various formats
5. **Orchestrator** - Coordinates the workflow using the four interfaces

## Core Interfaces

### 1. Provisioner

Manages cluster infrastructure lifecycle (provision, access, destroy):

```go
type Provisioner interface {
    Provision(ctx context.Context) (*ClusterInfo, error)
    Destroy(ctx context.Context, cluster *ClusterInfo) error
    GetKubeconfig(ctx context.Context, cluster *ClusterInfo) ([]byte, error)
    Health(ctx context.Context, cluster *ClusterInfo) (*HealthStatus, error)
}
```

**Design Principles:**
- Provider-agnostic (works with OCM, EKS, GKE, kind, etc.)
- Extensions via `ClusterInfo.Properties` map
- No assumptions about addons or provider-specific concepts

### 2. Executor

Runs tests or workloads against a target cluster:

```go
type Executor interface {
    Execute(ctx context.Context, target *ExecutionTarget) (*ExecutionResult, error)
}
```

**Design Principles:**
- Framework-agnostic (Ginkgo, pytest, shell scripts, etc.)
- Single `Execute` method - no hardcoded phases or test types
- Flexible categorization via `ExecutionTarget.Labels`
- Generic result structure works with any test framework

**Example Labels:**
```go
target := &ExecutionTarget{
    Labels: map[string]string{
        "phase":  "install",
        "suite":  "e2e",
        "type":   "smoke",
    },
}
```

### 3. Analyzer

Performs post-execution analysis of failures:

```go
type Analyzer interface {
    Analyze(ctx context.Context, input *AnalysisInput) (*AnalysisResult, error)
    ShouldAnalyze(result *ExecutionResult) bool
}
```

**Design Principles:**
- Not tied to specific analysis methods (AI, rule-based, ML, etc.)
- Conditional execution via `ShouldAnalyze`
- Extensions via `AnalysisResult.Metadata` map

### 4. Reporter

Handles test result reporting:

```go
type Reporter interface {
    Initialize(ctx context.Context) error
    Report(ctx context.Context, input *ReportInput) error
    Finalize(ctx context.Context) error
}
```

**Design Principles:**
- Format-agnostic (JUnit, HTML, Slack, databases, etc.)
- Supports composite pattern for multiple reporters
- Lifecycle methods for setup and cleanup

## Key Types

### ClusterInfo

Represents cluster connection information:

```go
type ClusterInfo struct {
    ID         string
    Name       string
    Provider   string                 // e.g., "ocm", "eks", "gke"
    Region     string
    Version    string
    Kubeconfig []byte
    Properties map[string]interface{} // Provider-specific extensions
}
```

### ExecutionTarget

Defines what and where to execute:

```go
type ExecutionTarget struct {
    Cluster    *ClusterInfo
    Kubeconfig []byte
    Labels     map[string]string // Flexible categorization
    Timeout    time.Duration
}
```

### ExecutionResult

Generic test execution outcome:

```go
type ExecutionResult struct {
    Success   bool
    StartTime time.Time
    EndTime   time.Time
    Summary   *ResultSummary
    Artifacts []Artifact
    Metadata  map[string]interface{} // Framework-specific data
}
```

## Usage Examples

### Implementing a Custom Provisioner

```go
type MyCloudProvisioner struct {
    // Your cloud provider client
}

func (p *MyCloudProvisioner) Provision(ctx context.Context) (*orchestrator.ClusterInfo, error) {
    // Provision cluster using your cloud provider
    cluster := createCluster()
    
    return &orchestrator.ClusterInfo{
        ID:         cluster.ID,
        Name:       cluster.Name,
        Provider:   "mycloud",
        Kubeconfig: cluster.Kubeconfig,
        Properties: map[string]interface{}{
            "instanceType": cluster.InstanceType,
            "customField":  cluster.CustomData,
        },
    }, nil
}
```

### Implementing a Custom Executor

```go
type PytestExecutor struct {
    testDir string
}

func (e *PytestExecutor) Execute(ctx context.Context, target *orchestrator.ExecutionTarget) (*orchestrator.ExecutionResult, error) {
    // Run pytest tests
    result := runPytest(e.testDir, target.Kubeconfig)
    
    return &orchestrator.ExecutionResult{
        Success:   result.ExitCode == 0,
        StartTime: result.StartTime,
        EndTime:   result.EndTime,
        Summary: &orchestrator.ResultSummary{
            Total:  result.TotalTests,
            Passed: result.PassedTests,
            Failed: result.FailedTests,
        },
        Artifacts: result.Artifacts,
        Metadata: map[string]interface{}{
            "framework": "pytest",
            "version":   "7.0.0",
        },
    }, nil
}
```

## Using the Orchestrator

### Direct Component Construction

When you have component instances, simply construct the orchestrator:

```go
import "github.com/openshift/osde2e/pkg/common/orchestrator"

func main() {
    // Create your component implementations
    provisioner := MyProvisioner{}
    executor := MyExecutor{}
    analyzer := MyAnalyzer{}
    reporter := MyReporter{}

    // Create orchestrator with components
    orch := orchestrator.NewOrchestratorWithComponents(
        provisioner, executor, analyzer, reporter,
    )
    
    // Run the workflow
    exitCode := orch.Run(context.Background())
    os.Exit(exitCode)
}
```

### Via E2E Factory

The `pkg/e2e` package provides a simple factory for E2E orchestrators:

```go
import "github.com/openshift/osde2e/pkg/e2e"

// Create E2E orchestrator (OCM + Ginkgo + AI analysis + full reporting)
orch, err := e2e.NewOrchestrator()
if err != nil {
    log.Fatal(err)
}
exitCode := orch.Run(context.Background())

// Or create with custom components
provisioner, _ := provision.NewOCMProvisioner()
executor := execute.NewGinkgoExecutor()
analyzer := analyze.NewNoOpAnalyzer()  // Disable AI analysis
reporter := report.NewCompositeReporter()

orch := e2e.NewOrchestratorWithComponents(provisioner, executor, analyzer, reporter)
```

## Integration with E2E Framework

The `pkg/e2e` package provides concrete implementations of these interfaces:

- **provision.OCMProvisioner** - OpenShift Cluster Manager provisioning
- **execute.GinkgoExecutor** - Ginkgo-based test execution
- **analyze.AIAnalyzer** - AI-powered log analysis
- **report.CompositeReporter** - Multi-format reporting

See `pkg/e2e/README.md` for details on the E2E implementations.

## Design Benefits

1. **Reusability** - Interfaces can be used by any testing framework
2. **Testability** - Each interface can be mocked independently
3. **Flexibility** - No hardcoded assumptions about frameworks or providers
4. **Extensibility** - Extensions via Properties and Metadata maps
5. **Composability** - Combine multiple implementations using composition patterns
6. **No Import Cycles** - Clean package structure in common/

## Exit Codes

Standard exit codes for orchestrator workflows:

```go
const (
    SuccessExitCode = 0
    FailureExitCode = 1
)
```

## Future Extensions

This interface design supports future enhancements such as:

- **Different Cloud Providers**: AWS EKS, GCP GKE, Azure AKS, local kind
- **Different Test Frameworks**: pytest, Robot Framework, shell scripts, Terraform
- **Different Analyzers**: Rule-based, ML models, external services
- **Different Reporters**: Slack, email, databases, S3, custom dashboards

## Contributing

When defining new interfaces or types:

1. Keep interfaces small and focused (single responsibility)
2. Use context for cancellation and timeouts
3. Provide extensibility via Properties/Metadata maps
4. Document expected behavior and edge cases
5. Ensure backward compatibility when modifying existing interfaces

