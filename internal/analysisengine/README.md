# LLM-based Failure Analysis Engine

Analyzes test failures and artifacts using LLM (Gemini) to provide intelligent insights and root cause analysis.

## Workflow

```
┌─────────────┐    ┌──────────────┐    ┌─────────────┐    ┌──────────────┐
│ Test        │    │ Aggregator   │    │ PromptStore │    │ LLM Client   │
│ Artifacts   ├───▶│ Collects:    ├───▶│ Renders     ├───▶│ (Gemini)     │
│             │    │ • JUnit XML  │    │ templates   │    │ Analyzes     │
│ • Logs      │    │ • Log files  │    │ with data   │    │ with tools   │
│ • Results   │    │ • Failed     │    │ variables   │    │              │
│ • Failures  │    │   tests      │    │             │    │              │
└─────────────┘    └──────────────┘    └─────────────┘    └──────┬───────┘
                                                                 │
┌────────────────────────────────────────────────────────────────▼────────┐
│ Output: summary.yaml in llm-analysis/ directory                         │
│ • Analysis results  • Cluster info  • Metadata  • Original prompt       │
└─────────────────────────────────────────────────────────────────────────┘
```

## Components

- **Engine**: Orchestrates the analysis workflow
- **Config**: Analysis configuration (API keys, templates, cluster info)
- **ClusterInfo**: Cluster metadata (ID, provider, version, etc.)
- **Result**: Analysis output with summary and metadata

## Usage

```go
engine, err := analysisengine.New(ctx, &analysisengine.Config{
    ArtifactsDir:   "/path/to/artifacts",
    PromptTemplate: "default",
    APIKey:         os.Getenv("GEMINI_API_KEY"),
    ClusterInfo:    clusterInfo,
})
result, err := engine.Run(ctx)
```

## Output

Creates `llm-analysis/summary.yaml` with:
- LLM analysis and recommendations
- Cluster and failure context
- Examined artifacts count
- Complete prompt and response data
