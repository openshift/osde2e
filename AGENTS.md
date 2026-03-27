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

## Configuration
- **Primary file**: `pkg/common/config/config.go` - All configuration options (START HERE)
- **Pattern**: const key + env var + default
- **Access**: `viper.GetString(config.SomeKey)`
- **Precedence**: CLI flags → env vars → custom YAML → defaults
- **Required env vars**: OCM_CLIENT_ID, OCM_CLIENT_SECRET, AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY
- **Optional env vars**:
  - `CLUSTER_ID` - Reuse existing cluster
  - `AD_HOC_TEST_IMAGES` - Override test images (comma-separated list)
  - `TEST_SUITES_YAML` - YAML list of test suites with optional slack channels (format: `- image: ...\n  slackChannel: ...`)

## Key Files & Components
- `cmd/osde2e/test/cmd.go` - CLI entry point
- `pkg/e2e/e2e.go` - Main orchestrator
- `pkg/common/cluster/clusterutil.go` - Cluster lifecycle
- `pkg/common/providers/` - Cloud provider implementations (OCM, ROSA)
- `pkg/common/cluster/healthchecks/` - Health validation logic
- `pkg/common/executor/executor.go` - Ad-hoc executor pods
- `pkg/common/runner/runner.go` - Runner pods
- `pkg/e2e/adhoctestimages/adhoctestimages.go` - Ad-hoc test suite driver
- `internal/llm/` - LLM/AI integration (Gemini)

## Common Patterns

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

## Architecture
```
osde2e
├── cmd/osde2e/          # CLI commands (provision, test, cleanup, krknai)
├── pkg/common/          # Core logic (config, providers, helpers)
├── internal/            # LLM analysis (llm, sanitizer, prompts)
└── test/                # Standalone Ginkgo test suites
```

 
## Before You Commit
- gofumpt all changed files
- run unit tests except test/ folder
- make build
- update README and AGENTS.md for changes made


## Testing Instructions

### E2E Tests
- **CLI**: `go run cmd/osde2e/main.go test --skip-health-check --skip-must-gather --skip-destroy-cluster --configs=rosa,sts,stage,ad-hoc-image`
- **IDE debugger**: VSCode (use `configs/local/example-launch.json`), GoLand (use `configs/local/example-e2e.run.xml`)

### Unit Tests
- Add or update unit tests for concrete implementation changes and new functionality
- Fix any test or type errors until the whole suite is green

## E2E-Suite Execution Flow

### Overview
The e2e-suite execution follows a multi-stage pipeline from test command initialization through cluster provisioning, test execution in isolated pods, multi-level artifact gathering, and cleanup.

### Execution Pipeline

1. **Command Initialization** (`cmd/osde2e/test/cmd.go:160-216`)
   - Load configuration (CLI flags → env vars → YAML → defaults)
   - Call `e2e.RunTests()` orchestrator

2. **Orchestrator Setup** (`pkg/e2e/e2e.go:42-97`)
   ```
   NewOrchestrator() → Provision() → Execute() → AnalyzeLogs() →
   PostProcessCluster() → Report() → Cleanup()
   ```

3. **Cluster Provisioning** (`pkg/common/cluster/clusterutil.go`)
   - Load/reuse existing cluster or provision new via OCM provider
   - Run health checks (CVO, nodes, operators, certs, daemonsets)
   - Retrieve kubeconfig and configure cluster access

4. **Test Execution** (Two Patterns)

   **Pattern A: Runner Pods** (`pkg/common/runner/runner.go`) - Traditional OpenShift test suites
   - Get test image from ImageStream
   - Create Job pod with git init containers for repo cloning
   - Stream logs from all containers to `{reportDir}/{phase}/containerLogs/`
   - Wait for completion with timeout

   **Pattern B: Ad-Hoc Executor Pods** (`pkg/common/executor/executor.go`) - Modern test suites
   - Create isolated namespace per test suite
   - Deploy 2-container pod:
     - `e2e-suite`: Runs test image, writes results to `/test-run-results`
     - `pause-for-artifacts`: Keeps pod alive (`tail -f /dev/null`) for artifact collection
   - Inject cluster metadata as env vars (OCM_CLUSTER_ID, CLOUD_PROVIDER_ID, etc.)
   - Wait for e2e-suite container completion

5. **Multi-Level Artifact Gathering**

   **Level 1: Pod Logs** (`pkg/common/runner/service.go:95-135`)
   - Stream logs from all containers to individual files

   **Level 2: Test Suite Results** (`pkg/common/executor/executor.go:302-398`)
   - Fetch pod logs from e2e-suite container
   - Execute `tar cf - /test-run-results` in pause container
   - Stream and extract tar archive via SPDY protocol
   - Process JUnit XML results

   **Level 3: Cluster Diagnostics** (`pkg/common/cluster/clusterutil.go`)
   - Run `oc adm must-gather` → `{reportDir}/must-gather`
   - Inspect cluster state (projects, OLM)

   **Level 4: Reports & Analysis** (`pkg/e2e/e2e.go:579-643`)
   - Generate JUnit XML: `{phaseDir}/junit_{suffix}.xml`
   - Create Konflux JSON report (if configured)
   - Run AI-powered log analysis on failures → `{reportDir}/analysis/`

6. **Reporting & Notifications** (`pkg/e2e/e2e.go:271-302`)
   - Upload all artifacts to S3 with presigned URLs
   - Send Slack notifications with links:
     - Per-test-image notifications (ad-hoc) with analysis results
     - General failure notifications with aggregated results

7. **Cleanup** (`pkg/e2e/e2e.go:559-571`)
   - Delete test namespaces (executor pods)
   - Run must-gather and cluster inspection
   - Update OCM cluster properties
   - Delete cluster (unless `--skip-destroy-cluster` flag set)

### Configuration Levels

**Level 1: Main osde2e Test Command** (CLI/Environment)
- `configs/ad-hoc-image.yaml`: Test suite images, timeouts, Slack channels
- See "Configuration" section above for env vars and precedence

**Level 2: Executor E2E-Suite Job** (Pod - Injected by `pkg/common/executor/executor.go:186-268`)
- `OCM_CLUSTER_ID`: Target cluster identifier
- `OCM_ENV`: OCM environment (stage/production)
- `CLOUD_PROVIDER_ID`: Cloud provider (aws/gcp/azure)
- `CLOUD_PROVIDER_REGION`: Cluster region
- `OCM_CCS`: Customer Cloud Subscription flag
- `GINKGO_NO_COLOR`: Disable colored output
- Shared volume: EmptyDir at `/test-run-results` (e2e-suite writes, pause container serves)

## Exception Tests (Non-Executor Pattern)

A few legacy tests exist directly in `pkg/e2e/` that run without test-suite executor jobs. These are exceptions to the standard execution pattern.

### How They Work

**Direct Execution**
- Tests live in `pkg/e2e/` directory (e.g., `workloads`, `verify`, `operators`)
- Run as part of the main osde2e Ginkgo suite
- Execute directly in the runner pod without spawning executor pods
- Test output and results are fetched directly from the runner pod
- No need for artifact collection via pause containers

**Key Differences from Executor Pattern**
- No isolated namespace per test
- No 2-container pod architecture (e2e-suite + pause)
- No tar-based artifact fetching via SPDY
- Results captured directly through Ginkgo's reporting mechanisms

**Future Direction**
- Users should prefer test suites (executor pattern) for new tests
- These direct tests are maintained for backward compatibility
- New platform component tests should use boilerplate pattern instead

**Files**: `pkg/e2e/workloads/`, `pkg/e2e/verify/`, `pkg/e2e/operators/`

## Secret Management

Secrets from vault volumes are automatically loaded and propagated from the top-level osde2e pod to second-level test suite executor pods.

### How It Works

**In Prow Jobs & Progressive Delivery**
- Vault secrets are mounted as volumes via `--secret-file-locations` flag (e.g., `/secrets/vault`)
- At startup, osde2e reads all files from these mounted directories
- Each file becomes a secret: filename becomes the key, file contents become the value
- Secrets are stored in viper configuration and combined into a passthrough map

**Propagation to Test Pods**
- Top-level osde2e pod collects all secrets into a single map
- For each test suite execution, a Kubernetes Secret named `ci-secrets` is created in the test namespace
- The executor pod references this secret via `envFrom.secretRef`
- All secrets automatically become environment variables in the test suite container
- Test suites access secrets as standard environment variables

**Files**: `pkg/common/load/load.go` (vault loading), `pkg/common/executor/executor.go:158-177` (pod injection)

## PR instructions
- Title format: [<jira-ID>] <Title>
