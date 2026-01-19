# AI Agent Instructions - OSDe2e

## What This Is
End-to-end testing framework for Managed services for OSD/ROSA. It is currently applied as a test framework in ROSA nightly CI jobs, OSD operator CICD pipelines and on-demand test schedules such as gap analysis. 

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

## Interacting With The Code
- Always stick to existing design patterns in code
- Keep new code simple and concise
- Keep pull requests small
- Always reuse existing code
- Use go language best practices
- Prefer implementing existing or new interfaces, don't write extensive procedural logic
- Extend test helper functionality in https://github.com/openshift/osde2e-common, not here

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
