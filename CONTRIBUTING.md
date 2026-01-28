# Contributing to OSDe2e

Thank you for your interest in contributing to OSDe2e! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Code Standards](#code-standards)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Commit Message Guidelines](#commit-message-guidelines)
- [Pull Request Process](#pull-request-process)
- [Architecture Guidelines](#architecture-guidelines)
- [Common Tasks](#common-tasks)

## Code of Conduct

By participating in this project, you agree to abide by the [OpenShift Code of Conduct](https://github.com/openshift/community/blob/master/CODE_OF_CONDUCT.md). Please treat everyone with respect and create a welcoming environment for all contributors.

## Getting Started

### Prerequisites

Before you begin, ensure you have:

1. **Go workspace** running the minimal version defined in [go.mod](go.mod)
2. **OCM Service Account credentials**
   - `OCM_CLIENT_ID` - OCM service account client ID
   - `OCM_CLIENT_SECRET` - OCM service account client secret
3. **Adequate quota** for deploying clusters (verify at [ocm-resources](https://gitlab.cee.redhat.com/service/ocm-resources/))
4. **Git** and **Make** installed
5. **AWS credentials** (if testing ROSA/OSD clusters)
   - `AWS_ACCESS_KEY_ID`
   - `AWS_SECRET_ACCESS_KEY`

### Setting Up Your Development Environment

1. **Fork and clone the repository:**

```bash
git clone https://github.com/YOUR_USERNAME/osde2e.git
cd osde2e
```

2. **Add upstream remote:**

```bash
git remote add upstream https://github.com/openshift/osde2e.git
```

3. **Install dependencies:**

```bash
go mod tidy
```

4. **Build the project:**

```bash
make build
```

5. **Verify the build:**

```bash
./out/osde2e --help
```

## Development Workflow

### Creating a New Branch

Always create a feature branch from the latest `main`:

```bash
git checkout main
git pull upstream main
git checkout -b feature/SDCICD-1234-your-feature-name
```

**Branch naming convention:**
- Feature: `feature/SDCICD-XXXX`
- Bug fix: `fix/SDCICD-XXXX`
- Documentation: `docs/description`

### Making Changes

- Keep changes focused: One feature or fix per PR
- Follow code standards (see [Code Standards](#code-standards))
- Write tests: Add unit tests for new functionality
- Update documentation: Keep README and docs in sync

### Testing Your Changes

Before committing, always:

```bash
# Format code (IMPORTANT: use gofumpt, not gofmt!)
gofumpt -w .

# Build
make build

# Run unit tests
go test ./... -v

# Run linting
make lint
```

### Running E2E Tests

E2E tests require specific environment variables:

```bash
# Required environment variables
export OCM_CLIENT_ID="your-client-id"
export OCM_CLIENT_SECRET="your-client-secret"
export AWS_ACCESS_KEY_ID="your-aws-key"
export AWS_SECRET_ACCESS_KEY="your-aws-secret"
export CLUSTER_ID="existing-cluster-id"  # Or let osde2e create one
export AD_HOC_TEST_IMAGES="quay.io/your-test-image:tag"  # For ad-hoc tests

# Run via CLI
go run cmd/osde2e/main.go test \
  --skip-health-check \
  --skip-must-gather \
  --skip-destroy-cluster \
  --configs=rosa,sts,stage,ad-hoc-image
```

**IDE-specific configurations:**
- **VSCode**: Use `configs/local/example-launch.json`
- **GoLand**: Use `configs/local/example-e2e.run.xml`

## Code Standards

### Go Best Practices

- Follow [Effective Go](https://golang.org/doc/effective_go.html) conventions
- Always use `gofumpt`, NOT `gofmt`
- Use `logr`/`klog` for logging, NOT `fmt.Println`
- Always check and handle errors appropriately
- Pass `context.Context` for cancellation and timeouts

### Project-Specific Guidelines

#### Configuration Management

All configuration must go through the config system in `pkg/common/config/config.go`:

```go
// 1. Define constant
const MyNewSetting = "my.new.setting"

// 2. Set environment variable mapping
viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

// 3. Provide default value
viper.SetDefault(MyNewSetting, "default-value")
```

**Precedence**: CLI flags > Env vars > Custom YAML > Defaults

#### Cross-Repository Integration

- **Cross-reference osde2e-common**: Always check [osde2e-common](https://github.com/openshift/osde2e-common) for existing functionality before implementing new features
- **Core clients go in osde2e-common**: All core client functionality MUST be in osde2e-common to make it available across all consumer repositories
- **Reuse code**: Analyze `pkg/common` to extend existing utilities before creating new ones
- **Use h.Client**: Always use `h.Client` for k8s object CRUD operations
- **Propose commits in osde2e-common** where changes to core clients are needed

#### Code Organization

```
osde2e/
├── cmd/osde2e/  # CLI commands (provision, test, cleanup, krknai)
├── pkg/common/  # Core logic (config, providers, helpers)
├── internal/    # Internal packages (llm, sanitizer, prompts)
└── test/        # Standalone Ginkgo test suites
```

- **Adhere to project architecture**: Strictly follow the structure in `pkg/`
- **Prefer interfaces**: Implement or extend interfaces, don't write extensive procedural logic
- **Keep code simple and concise**

### Code Style

```go
// Good: Clear, concise, with proper error handling
func ProcessCluster(ctx context.Context, clusterID string) error {
    logger := log.FromContext(ctx)

    cluster, err := getCluster(ctx, clusterID)
    if err != nil {
        return fmt.Errorf("failed to get cluster: %w", err)
    }

    logger.Info("Processing cluster", "id", cluster.ID, "state", cluster.State)
    return nil
}

// Bad: No error handling, uses fmt.Println
func ProcessCluster(clusterID string) {
    cluster := getCluster(clusterID)
    fmt.Println("Processing", cluster.ID)
}
```

## Testing

### Unit Tests

- Add or update unit tests for concrete implementation changes and new functionality, even if nobody asked
- Aim for >80% coverage on new code
- Use table-driven tests for multiple scenarios

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid input", "test", "TEST", false},
        {"empty input", "", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := MyFunction(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("MyFunction() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("MyFunction() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Integration Tests

Integration tests require credentials and should be run in CI or with proper setup:

```bash
# Run all tests
go test ./... -v

# Run specific package
go test ./pkg/common/cluster -v

# Run with coverage
go test ./... -cover
```

### Test Organization

- Unit tests: `*_test.go` alongside implementation files
- Integration tests: May require build tags or separate directories
- E2E tests: In `test/` directory using Ginkgo framework

## Submitting Changes

### Before You Commit

```bash
gofumpt -w .        # Format (not gofmt!)
make build          # Compile
go test ./... -v    # Test
git status          # Check git status
```

### Commit Message Guidelines

Follow the conventional commit format:

```
[JIRA-ID] type: short description

Longer description if needed, explaining:
- Why this change is necessary
- What problem it solves
- Any breaking changes or side effects

Fixes: SDCICD-1234
```

**Commit types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**

```
[SDCICD-1234] feat: add LLM-based failure analysis

Implements LLM-powered analysis of test failures using Gemini API.
Includes data sanitization and Slack notifications.

Fixes: SDCICD-1234
```

```
[SDCICD-5678] fix: resolve cluster health check timeout

Increases timeout from 30s to 60s to accommodate slower cluster responses
in stage environment.

Fixes: SDCICD-5678
```

### Commit Best Practices

- Atomic commits: Each commit should be a logical unit
- Descriptive messages: Explain WHY, not just WHAT
- Reference issues: Always include JIRA ID in brackets

## Pull Request Process

### Creating a Pull Request

1. **Push your branch:**

```bash
git push origin feature/SDCICD-1234-your-feature
```

2. **Open PR** on GitHub with the template below

3. **PR Title Format:**

```
[SDCICD-1234] Brief description of change
```

### PR Description Template

```markdown
## Summary
Brief description of what this PR does.

## JIRA
[SDCICD-1234](https://issues.redhat.com/browse/SDCICD-1234)

## Changes
- Added feature X
- Updated configuration for Y
- Fixed bug in Z

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests passed
- [ ] Manual testing completed

## Documentation
- [ ] README updated (if needed)
- [ ] Code comments added
- [ ] AGENTS.md updated (if needed)

## Checklist
- [ ] Code follows project standards
- [ ] Tests pass locally
- [ ] Formatted with gofumpt
- [ ] No linter errors
- [ ] Breaking changes documented
```

### PR Review Process

- Automated checks: CI must pass (tests, linting, build)
- Code review: At least one approval from CODEOWNERS other than author
- Address feedback: Respond to all review comments
- Keep updated: Rebase on main if needed

```bash
# Rebase on latest main
git fetch upstream
git rebase upstream/main
git push -f origin feature/SDCICD-1234-your-feature
```

### Merging

- Squash and merge is preferred for feature branches
- Ensure clean commit history on main branch
- Delete branch after merge

## Architecture Guidelines

### Core Test Workflow

1. Load config (CLI flags → env vars → custom YAML → defaults)
2. Provision cluster (or use existing via CLUSTER_ID)
3. Health check (optional)
4. Run tests
5. Upgrade (optional)
6. Cleanup (optional)

### Key Files

- `pkg/common/config/config.go` - All configuration options (START HERE)
- `cmd/osde2e/main.go` - Entry point
- `pkg/common/providers/` - Cloud provider implementations (OCM, ROSA)
- `pkg/common/cluster/healthchecks/` - Health validation logic
- `internal/llm/` - LLM/AI integration (Gemini)

### Common Patterns

**Configuration:**
- Everything in `config.go`: const key + env var + default
- Access via `viper.GetString(config.SomeKey)`
- Precedence: CLI > Env > Custom YAML > Default

**Providers:**
- Interface: `pkg/common/spi/`
- Registered in `main.go`

**Authoring new platform component tests:**
- Use [boilerplate](https://github.com/openshift/boilerplate/blob/master/boilerplate/openshift/golang-osd-e2e/README.md), not here
- See [OSDE2E Test Harness](https://github.com/openshift/osde2e-example-test-harness) for examples

## Common Tasks

### Adding a New Configuration Option

- Add constant in `pkg/common/config/config.go`
- Set default value
- Add environment variable mapping
- Update documentation in `docs/Config.md`
- Add CLI flag if needed

### Adding a New Test

See [Writing Tests](/docs/Writing-Tests.md) for detailed guidelines.

Quick checklist:
- Use Ginkgo framework
- Add appropriate labels
- Include informative test descriptions
- Handle cleanup in `AfterEach`

### Adding a New Provider

- Implement the SPI interface in `pkg/common/spi/`
- Add implementation in `pkg/common/providers/`
- Register in `cmd/osde2e/main.go`
- Add tests and documentation

### Updating Dependencies

```bash
go get -u github.com/example/package@v1.2.3  # Update specific dependency
go mod tidy                                    # Tidy up
```

**Always ask for confirmation before adding new dependencies to go.mod**

### Working with osde2e-common

When changes affect shared functionality:

```bash
# 1. In osde2e go.mod, use replace directive
replace github.com/openshift/osde2e-common => /local/path/to/osde2e-common

# 2. Test locally
go test ./...

# 3. Remove replace directive before committing
```

Then:
- Propose PR in osde2e-common
- Update osde2e after osde2e-common merge

## Getting Help

- Slack: Join `#sd-cicd` on CoreOS Slack
- GitHub Issues: Open an issue for bugs or feature requests
- Documentation: Check [docs/](/docs/) directory
- AGENTS.md: AI agent instructions and project overview

## Additional Resources

- [AGENTS.md](AGENTS.md) - AI agent instructions and project overview
- [Writing Tests](/docs/Writing-Tests.md)
- [Configuration Reference](/docs/Config.md)
- [Log Analysis System](/docs/Log-Analysis-System.md)
- [Testing with OSDe2e](/docs/testing-with-osde2e.md) - Testing overview and guides
- [Running OSDe2e Tests](/docs/run-osde2e-tests.md) - Running tests on clusters
- [Ad-hoc Testing](/docs/adhoc-osde2e-testing.md) - Ad-hoc test procedures
- [CI Jobs](/docs/CI-Jobs.md)
- [Test Harness Guide](https://github.com/openshift/osde2e-example-test-harness)
- [OpenShift Documentation](https://docs.openshift.com/)

---

Thank you for contributing to OSDe2e! Your efforts help improve the quality and reliability of Managed OpenShift services.
