# Krkn-AI Chaos Testing — osde2e Command

Krkn-AI is an AI-powered chaos testing framework integrated into osde2e. It uses a genetic algorithm to discover, evolve, and execute chaos scenarios against OpenShift clusters, then analyzes the results with an LLM.

[Krkn-AI](https://github.com/krkn-chaos/krkn-ai) leverages an evolutionary algorithm to identify chaos scenarios that impact the stability of your cluster and applications:

1. **Initial Population** — Creates random chaos scenarios based on your configuration
2. **Fitness Evaluation** — Runs each scenario and measures system response using Prometheus metrics
3. **Selection** — Identifies the most effective scenarios based on fitness scores
4. **Evolution** — Creates new scenarios through crossover and mutation
5. **Health Monitoring** — Continuously monitors application health during experiments
6. **Iteration** — Repeats the process across multiple generations to find optimal scenarios

## Overview

The osde2e integration wraps this in four phases:

1. **Discover** — The krkn-ai container scans the target namespace to identify pods, nodes, and potential chaos scenarios, outputting a `krkn-ai.yaml` config file.
2. **Update YAML** — osde2e merges user-provided config (scenarios, generations, population, fitness query, health checks) into the discovered `krkn-ai.yaml`.
3. **Run** — The container executes the evolved scenarios using the genetic algorithm (evolving across generations), producing CSV results and health check reports.
4. **Analyze** — osde2e optionally runs LLM-powered analysis on the results to generate a human-readable summary.

## Prerequisites

- An existing OpenShift cluster (ROSA or OSD) accessible via OCM
- A container runtime (`podman` or `docker`) available in the execution environment
- OCM credentials (`OCM_CLIENT_ID` / `OCM_CLIENT_SECRET`)
- AWS credentials (`AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY`) and `AWS_REGION`
- (Optional) `GEMINI_API_KEY` for LLM-powered log analysis

## Basic Usage

```bash
osde2e krkn-ai \
  --cluster-id <CLUSTER_ID> \
  --configs rosa,sts,stage \
  --skip-destroy-cluster \
  --log-analysis-enable
```

## CLI Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--configs` | Comma-separated list of built-in configs (e.g. `rosa,sts,stage`) | `""` |
| `--cluster-id`, `-i` | Existing OCM cluster ID to run chaos tests against | **required** |
| `--environment`, `-e` | OCM environment (`stage`, `int`) | from config |
| `--kube-config`, `-k` | Path to local kubeconfig file | `""` |
| `--skip-destroy-cluster` | Preserve the cluster after testing | `false` |
| `--skip-must-gather` | Skip must-gather collection after the chaos run | `false` |
| `--log-analysis-enable` | Enable LLM-powered analysis of chaos test results | `false` |
| `--custom-config` | Path to a custom YAML config file | `""` |
| `--secret-locations` | Comma-separated list of secret directory locations | `""` |

## Built-in Configs

The `--configs` flag loads YAML config files that set provider and environment settings:

| Config | Sets |
|--------|------|
| `rosa` | `provider: rosa` |
| `sts` | `rosa.STS: true` |
| `stage` | `ocm.env: stage`, `rosa.env: stage` |
| `int` | `ocm.env: int`, `rosa.env: int` |

Typical usage: `--configs rosa,sts,stage` for a ROSA STS cluster in the stage environment.

## Environment Variables

### Cluster and Provider

| Variable | Description | Default |
|----------|-------------|---------|
| `CLUSTER_ID` | Existing cluster ID (required) | `""` |
| `ROSA_ENV` / `OCM_ENV` | OCM environment (overrides config) | from `--configs` |
| `AWS_REGION` | AWS region — required for provider initialization (aliases: `ROSA_AWS_REGION`, `CLOUD_PROVIDER_REGION`) | `""` |
| `AWS_ACCESS_KEY_ID` | AWS access key | `""` |
| `AWS_SECRET_ACCESS_KEY` | AWS secret key | `""` |

### Chaos Targeting

| Variable | Description | Default |
|----------|-------------|---------|
| `KRKN_NAMESPACE` | Target namespace for chaos testing | `default` |
| `KRKN_POD_LABEL` | Label selector for targeting pods (e.g. `app=myservice`) | `""` |
| `KRKN_NODE_LABEL` | Label selector for targeting nodes | `kubernetes.io/hostname` |
| `KRKN_SKIP_POD_NAME` | Regex pattern to skip specific pods | `""` |
| `KRKN_SCENARIOS` | Comma-separated list of scenarios to enable (empty = all discovered) | `""` |

### Genetic Algorithm

| Variable | Description | Default |
|----------|-------------|---------|
| `KRKN_GENERATIONS` | Number of generations to evolve scenarios | `2` |
| `KRKN_POPULATION` | Population size per generation (minimum 2) | `2` |

### Observability

| Variable | Description | Default |
|----------|-------------|---------|
| `KRKN_FITNESS_QUERY` | PromQL query for the fitness function | `""` |
| `KRKN_HEALTH_CHECK` | Health check endpoints in `name=url` format, comma-separated | `""` |
| `KRKN_TOP_SCENARIOS_COUNT` | Number of top scenarios to include in the analysis report | `10` |

### LLM Analysis

| Variable | Description | Default |
|----------|-------------|---------|
| `GEMINI_API_KEY` | API key for the Gemini LLM service | `""` |
| `LLM_MODEL` | LLM model to use | `gemini-3.1-pro-preview` |

## Examples

### Local Development

```bash
export OCM_CLIENT_ID="..."
export OCM_CLIENT_SECRET="..."
export AWS_ACCESS_KEY_ID="..."
export AWS_SECRET_ACCESS_KEY="..."
export AWS_REGION="us-east-1"
export GEMINI_API_KEY="..."

osde2e krkn-ai \
  --cluster-id 2abc123def456 \
  --configs rosa,sts,stage \
  --skip-destroy-cluster \
  --skip-must-gather \
  --log-analysis-enable
```

### Targeting a Specific Namespace

```bash
export KRKN_NAMESPACE="openshift-console"
export KRKN_POD_LABEL="app=console"
export KRKN_GENERATIONS="3"
export KRKN_POPULATION="4"

osde2e krkn-ai \
  --cluster-id 2abc123def456 \
  --configs rosa,sts,stage \
  --skip-destroy-cluster
```

### VSCode / Cursor Debugger

Create or update `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "krkn-ai",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/osde2e/main.go",
      "env": {
        "CLUSTER_ID": "<your-cluster-id>",
        "OCM_CLIENT_ID": "<your-ocm-client-id>",
        "OCM_CLIENT_SECRET": "<your-ocm-client-secret>",
        "AWS_ACCESS_KEY_ID": "<your-aws-key>",
        "AWS_SECRET_ACCESS_KEY": "<your-aws-secret>",
        "AWS_REGION": "us-east-1",
        "GEMINI_API_KEY": "<your-gemini-key>",
        "ROSA_ENV": "stage",
        "KRKN_NAMESPACE": "openshift-console",
        "KRKN_GENERATIONS": "2",
        "KRKN_POPULATION": "2",
        "REPORT_DIR": "/tmp/osde2e-report",
        "SHARED_DIR": "/tmp/osde2e-shared"
      },
      "args": [
        "krkn-ai",
        "--configs", "rosa,sts,stage",
        "--skip-destroy-cluster",
        "--skip-must-gather",
        "--log-analysis-enable"
      ]
    }
  ]
}
```

## Execution Flow

```
osde2e krkn-ai
  │
  ├─ PreProcess     Validate config, check container runtime
  ├─ Provision      Load cluster context (kubeconfig from OCM)
  ├─ Execute
  │    ├─ Discover  Run krkn-ai container in discover mode
  │    │             → scans target namespace, outputs krkn-ai.yaml
  │    ├─ Update    Merge user config (scenarios, generations, fitness
  │    │             query, health checks) into discovered krkn-ai.yaml
  │    └─ Run       Run krkn-ai container in run mode (privileged)
  │                  → executes chaos scenarios via genetic algorithm
  │                  → outputs CSV results and health check reports
  ├─ AnalyzeLogs    (if --log-analysis-enable) LLM analysis of results
  ├─ PostProcess    Must-gather collection (if not skipped)
  ├─ Report         Generate HTML report
  └─ Cleanup        Destroy cluster (if not skipped)
```

## Output

After a run, artifacts are split across two directories:

**Report directory** (`REPORT_DIR`, default `/tmp/osde2e-report`):

| File | Description |
|------|-------------|
| `reports/all.csv` | Raw results from all scenario executions |
| `reports/health_check_report.csv` | Health check results per scenario |
| `report.html` | HTML summary report with top scenarios |
| `llm-analysis/` | (if enabled) LLM analysis output |
| `must-gather/` | (if not skipped) OpenShift must-gather artifacts |

**Shared directory** (`SHARED_DIR`, default `/tmp/osde2e-shared`):

| File | Description |
|------|-------------|
| `kubeconfig` | Cluster kubeconfig fetched from OCM |
| `krkn-ai.yaml` | The discovered/merged chaos configuration |

## Troubleshooting

### "no container runtime found: install podman or docker"

The krkn-ai command requires `podman` or `docker` in `$PATH` to launch the krkn-ai sub-container. In CI (Prow), the base image does not include a container runtime, so krkn-ai cannot run in Prow PR checks. Use the [Jenkins job](Krkn-AI-Jenkins-Job.md) instead for CI execution.

### "aws variables were not set (access key id, secret access key, region)"

The ROSA provider requires AWS credentials and a region to initialize, even when reusing an existing cluster. Ensure `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, and `AWS_REGION` are all set.

### "exec: oc: executable file not found in $PATH"

The `oc` binary is needed for must-gather. If unavailable, must-gather is skipped gracefully with a warning. Use `--skip-must-gather` to suppress the warning entirely.

### "discover mode failed"

Check that `CLUSTER_ID` is valid and the cluster is accessible. Verify that OCM credentials are correct and the cluster is in a `ready` state. Ensure the OCM environment matches where the cluster lives (stage vs int).

### Health check URL validation errors

`KRKN_HEALTH_CHECK` URLs must be cluster API `/health` endpoints (e.g. `https://api.cluster.example.com:6443/healthz`). External URLs are rejected to prevent accidental data exfiltration from CI environments.

## Related

- [krkn-ai Jenkins job](Krkn-AI-Jenkins-Job.md)
- [krkn-ai documentation](https://krkn-chaos.dev/docs/krkn_ai/) (upstream)
- [krkn-ai container image](https://quay.io/repository/krkn-chaos/krkn-ai) (upstream)
- [osde2e ops-sop](https://github.com/openshift/ops-sop/tree/master/v4/howto/osde2e) (internal)
