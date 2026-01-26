# AI Agent Instructions 

## What This Is
Osde2e is End-to-end testing framework for Managed services for OSD/ROSA. 

# Agent Rules
- **Cross-reference osde2e-common**: Always download and analyze https://github.com/openshift/osde2e-common for existing core functionality and helper utilities before implementing new features. All core client functionality and helper updates MUST be made in osde2e-common to make them available across all consumer repositories. Only implement functionality in this repository if it's specific to osde2e testing logic.
- **Reuse code**:Always first analyze pkg/common to extend existing core and helper utilities before implementing new features
- Always use h.Client for k8s object crud operations
- Propose commits in github.com/openshift/osde2e-common where changes to core clients are needed
- Always ask for confirmation before adding new dependencies to go.mod
- **Adhere to project architecture**: Strictly follow the structure in pkg/
- Prefer implementing or extending interfaces, don't write extensive procedural logic
- Keep code simple and concise 
- Use go language best practices

## Core Test Workflow
1. Load config (CLI flags → env vars → custom YAML → defaults)
2. Provision cluster (or use existing via CLUSTER_ID)
3. Health check (optional)
4. Run tests
5. Upgrade (optional)
6. Cleanup (optional)

## Key Files
- `pkg/common/config/config.go` - All configuration options (START HERE)
- `cmd/osde2e/main.go` - Entry point
- `pkg/common/providers/` - Cloud provider implementations (OCM, ROSA)
- `pkg/common/cluster/healthchecks/` - Health validation logic
- `internal/llm/` - LLM/AI integration (Gemini)

## Common Patterns

### Configuration
- Everything in `config.go`: const key + env var + default
- Access via `viper.GetString(config.SomeKey)`
- Precedence: CLI > Env > Custom YAML > Default

### Providers
- Interface: `pkg/common/spi/`
- Registered in `main.go`


### Authoring new platform component tests
- To add new component tests: Use boilerplate (README https://github.com/openshift/boilerplate/blob/master/boilerplate/openshift/golang-osd-e2e/README.md), not here

### Logging
- Use `logr`/`klog`, not `fmt.Println`
- Structured logging preferred


## Quick Tips
1. Config changes? Edit `config.go` only
2. Provider logic? Use SPI abstraction
3. Integration test failures? Check credentials/env vars
4. Always use `gofumpt`, not `gofmt`
5. Check git status before committing

## Environment
- See `config.go` for complete list

## Architecture
```
osde2e
├── cmd/osde2e/          # CLI commands (provision, test, cleanup, krknai)
├── pkg/common/          # Core logic (config, providers, helpers)
├── internal/            # LLM analysis (llm, sanitizer, prompts)
└── test/                # Standalone Ginkgo test suites
```

 
## Before You Commit
```bash
gofumpt -w .        # Format (not gofmt!)
make build      # Compile
go test ./... -v    # Test (integration tests need credentials)
```

## E2E Testing instructions
- Do not allow e2e test unless following env vars are set: AD_HOC_TEST_IMAGES|CLUSTER_ID|OCM_CLIENT_ID|OCM_CLIENT_SECRET|AWS_ACCESS_KEY_ID|AWS_SECRET_ACCESS_KEY
- To run via cli:
```bash
go run cmd/osde2e/main.go test  --skip-health-check --skip-must-gather --skip-destroy-cluster --configs=rosa,sts,stage,ad-hoc-image"
````
- To run via IDE debugger, if IDE is VSCode, use configs/local/example-launch.json; if IDE is GoLand, use configs/local/example-e2e.run.xml

## Unit Testing instructions
- Add or update for unit tests for concrete implementation changes and new functionality, even if nobody asked.
- Fix any test or type errors until the whole suite is green.

## PR instructions
- Title format: [<jira-ID>] <Title>
