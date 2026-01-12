# E2E Testing Framework - Generic Interface-Based Architecture

This package provides E2E-specific implementations of the generic orchestrator interfaces defined in `pkg/common/orchestrator`. The architecture is based on four core interfaces that enable flexible, testable implementations across different testing frameworks.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    cmd/osde2e/test/cmd.go                   │
│                   (Entry Point)                              │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                    e2e.Factory                               │
│              (Component Factory)                             │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                  e2e.Orchestrator                            │
│         (Coordinates workflow)                               │
└─────┬───────────┬───────────┬───────────┬───────────────────┘
      │           │           │           │
      ▼           ▼           ▼           ▼
┌──────────┐ ┌─────────┐ ┌──────────┐ ┌──────────┐
│Provisioner│ │Executor │ │Analyzer  │ │Reporter  │
│Interface │ │Interface│ │Interface │ │Interface │
└─────┬────┘ └────┬────┘ └────┬─────┘ └────┬─────┘
      │           │           │           │
      ▼           ▼           ▼           ▼
┌──────────┐ ┌─────────┐ ┌──────────┐ ┌──────────┐
│   OCM    │ │ Ginkgo  │ │    AI    │ │Composite │
│Provisioner│ │Executor │ │Analyzer  │ │Reporter  │
└──────────┘ └─────────┘ └──────────┘ └──────────┘
```

## Core Interfaces

All interfaces are defined in `pkg/common/orchestrator/` for maximum reusability across the codebase. These generic interfaces can be used by any testing framework, not just osde2e.

### 1. Provisioner Interface

**Responsibility**: Manage cluster infrastructure lifecycle

```go
type Provisioner interface {
    Provision(ctx context.Context) (*ClusterInfo, error)
    Destroy(ctx context.Context, cluster *ClusterInfo) error
    GetKubeconfig(ctx context.Context, cluster *ClusterInfo) ([]byte, error)
    Health(ctx context.Context, cluster *ClusterInfo) (*HealthStatus, error)
}
```

**Implementations**:
- `provision.OCMProvisioner` - OpenShift Cluster Manager based provisioning

### 2. Executor Interface

**Responsibility**: Execute tests/workloads against a cluster

```go
type Executor interface {
    Execute(ctx context.Context, target *ExecutionTarget) (*ExecutionResult, error)
}
```

**Key Features**:
- Generic `Execute` method (no hardcoded phases)
- Uses labels for flexible categorization: `{"phase": "install"}`, `{"upgrade": "true"}`
- Framework-agnostic result structure

**Implementations**:
- `execute.GinkgoExecutor` - Ginkgo test framework based execution

### 3. Analyzer Interface

**Responsibility**: Analyze test results and artifacts

```go
type Analyzer interface {
    Analyze(ctx context.Context, input *AnalysisInput) (*AnalysisResult, error)
    ShouldAnalyze(result *ExecutionResult) bool
}
```

**Implementations**:
- `analyze.AIAnalyzer` - AI-powered log analysis using LLM
- `analyze.NoOpAnalyzer` - No-op implementation for testing

### 4. Reporter Interface

**Responsibility**: Report test results in various formats

```go
type Reporter interface {
    Initialize(ctx context.Context) error
    Report(ctx context.Context, input *ReportInput) error
    Finalize(ctx context.Context) error
}
```

**Implementations**:
- `report.CompositeReporter` - Combines multiple reporters
- `report.ArtifactCollector` - Collects logs, must-gather, etc.

## Package Structure

```
pkg/common/orchestrator/       # Generic orchestrator package
├── interfaces.go              # 4 main interfaces
├── types.go                   # Shared types (ClusterInfo, ExecutionResult, etc.)
├── orchestrator.go            # Generic workflow orchestrator
├── factory.go                 # Generic factory helper
└── README.md                  # Interface documentation

pkg/e2e/                       # E2E-specific implementations
├── provision/                 # Provisioner implementations
│   ├── ocm.go                 # OCM-based provisioner
│   └── ocm_test.go
│
├── execute/                   # Executor implementations
│   ├── ginkgo.go              # Ginkgo-based executor
│   └── ginkgo_test.go
│
├── analyze/                   # Analyzer implementations
│   ├── ai.go                  # AI-powered analyzer
│   ├── noop.go                # No-op analyzer
│   └── analyzer_test.go
│
├── report/                    # Reporter implementations
│   ├── composite.go           # Composite reporter
│   └── reporter_test.go
│
├── orchestrator.go            # Workflow coordinator
├── orchestrator_test.go
├── factory.go                 # Component factory
└── e2e.go                     # Legacy entry point (deprecated)
```

## Usage

### Basic Usage (Recommended)

```go
import "github.com/openshift/osde2e/pkg/e2e"

func main() {
    ctx := context.Background()
    
    // Create E2E orchestrator (OCM + Ginkgo + AI + full reporting)
    orchestrator, err := e2e.NewOrchestrator()
    if err != nil {
        log.Fatalf("Failed to create orchestrator: %v", err)
    }
    
    // Run the E2E workflow
    exitCode := orchestrator.Run(ctx)
    os.Exit(exitCode)
}
```

### Custom Implementations

Create an orchestrator with custom component instances:

```go
import (
    "github.com/openshift/osde2e/pkg/e2e"
    "github.com/openshift/osde2e/pkg/e2e/provision"
    "github.com/openshift/osde2e/pkg/e2e/execute"
    "github.com/openshift/osde2e/pkg/e2e/analyze"
    "github.com/openshift/osde2e/pkg/e2e/report"
)

func main() {
    // Create specific component instances
    provisioner, _ := provision.NewOCMProvisioner()
    executor := execute.NewGinkgoExecutor()
    analyzer := analyze.NewNoOpAnalyzer()  // Disable AI analysis
    reporter := report.NewCompositeReporter()  // Custom reporter
    
    // Create orchestrator with custom components
    orch := e2e.NewOrchestratorWithComponents(
        provisioner, executor, analyzer, reporter,
    )
    
    exitCode := orch.Run(context.Background())
    os.Exit(exitCode)
}
```

### Legacy API (Deprecated)

```go
import "github.com/openshift/osde2e/pkg/e2e"

func main() {
    exitCode := e2e.RunTests(context.Background())
    os.Exit(exitCode)
}
```

## Key Benefits

1. **True Testability**: Each interface can be mocked independently
2. **Framework Flexibility**: Not tied to Ginkgo - could add pytest, shell scripts, anything
3. **Extensibility**: Easy to add new implementations (new cloud providers, test frameworks, reporters)
4. **Composability**: Combine multiple executors, reporters, analyzers
5. **Clarity**: Clean separation - each interface has single, clear responsibility
6. **No Implementation Leakage**: Interfaces reveal WHAT, not HOW

## Example Alternative Implementations

With these generic interfaces, you can easily create:

- **Different Executors**: Pytest executor, shell script executor, Terraform executor
- **Different Provisioners**: AWS EKS provisioner, GCP GKE provisioner, local kind provisioner
- **Different Analyzers**: Rule-based analyzer, external service analyzer, no-op analyzer
- **Different Reporters**: Slack reporter, email reporter, database reporter, S3 uploader

## Testing

Each package includes unit tests with mock implementations:

```bash
# Run all e2e framework tests
go test ./pkg/e2e/...

# Run specific package tests
go test ./pkg/e2e/core
go test ./pkg/e2e/orchestrator_test.go
```

Mock implementations are available in `pkg/e2e/orchestrator_test.go` for use in your own tests.

## Migration from Old Code

The old monolithic `e2e.go` has been preserved for backward compatibility but is deprecated. The function mapping is:

| Old Function | New Location | Interface Method |
|--------------|--------------|------------------|
| `beforeSuite()` | `provision/ocm.go` | `Provisioner.Provision()` |
| `installAddons()` | `provision/ocm.go` | Internal to Provision |
| `deleteCluster()` | `provision/ocm.go` | `Provisioner.Destroy()` |
| `cleanupAfterE2E()` | `provision/ocm.go` | Internal to Destroy |
| `runGinkgoTests()` | `execute/ginkgo.go` | `Executor.Execute()` |
| `runTestsInPhase()` | `execute/ginkgo.go` | Internal to Execute |
| `runLogAnalysis()` | `analyze/ai.go` | `Analyzer.Analyze()` |
| `getLogs()` | `report/composite.go` | Internal to Reporter |

## Design Principles

1. **Framework Agnostic**: No assumptions about Ginkgo, phases, or specific test structures
2. **Context-Driven**: All operations receive context for cancellation and deadlines
3. **Error Transparent**: Return detailed errors, let orchestrator decide handling
4. **Stateful Config, Stateless Execution**: Initialize with config, execute with runtime context
5. **Extensible via Composition**: Use Properties/Metadata maps for provider-specific data

## Contributing

When adding new implementations:

1. Implement the appropriate interface from `pkg/common/orchestrator`
2. Add your implementation to a new file in the relevant package
3. Create unit tests using the mock implementations
4. Update the factory if you want it to be a default option
5. Document your implementation in this README

