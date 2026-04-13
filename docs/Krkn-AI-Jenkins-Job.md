# Krkn-AI Jenkins Job

The krkn-ai Jenkins job runs AI-powered chaos testing against an existing OpenShift cluster. It wraps the `osde2e krkn-ai` command inside a container with access to a rootless Podman socket, enabling the chaos sub-container to run on the Jenkins agent.

[Krkn-AI](https://github.com/krkn-chaos/krkn-ai) leverages an evolutionary algorithm to identify chaos scenarios that impact the stability of your cluster and applications:

1. **Initial Population** — Creates random chaos scenarios based on your configuration
2. **Fitness Evaluation** — Runs each scenario and measures system response using Prometheus metrics
3. **Selection** — Identifies the most effective scenarios based on fitness scores
4. **Evolution** — Creates new scenarios through crossover and mutation
5. **Health Monitoring** — Continuously monitors application health during experiments
6. **Iteration** — Repeats the process across multiple generations to find optimal scenarios

## Job URL

**[ci.int.devshift.net/view/osde2e/job/osde2e-krknai-parameterized-job](https://ci.int.devshift.net/view/osde2e/job/osde2e-krknai-parameterized-job/)**

## Quick Start

1. Navigate to the [job page](https://ci.int.devshift.net/view/osde2e/job/osde2e-krknai-parameterized-job/)
2. Click **Build with Parameters**
3. Fill in:
   - `CLUSTER_ID`: Your target cluster ID (required)
   - `CONFIGS`: `rosa,sts,stage` (for stage) or `rosa,sts,int` (for int)
   - `KRKN_NAMESPACE`: Target namespace (default: `openshift-console`)
4. Click **Build**
5. Artifacts are archived under the build's **Build Artifacts** section

## Parameters

| Parameter | Description | Default | Required |
|-----------|-------------|---------|----------|
| `CLUSTER_ID` | Cluster ID to run chaos scenarios against | — | Yes |
| `CONFIGS` | osde2e configs to load (controls provider, STS, environment) | `rosa,sts,stage` | Yes |
| `SKIP_MUST_GATHER` | Skip must-gather collection | `false` | — |
| `KRKN_NAMESPACE` | Target namespace for chaos testing | `openshift-console` | — |
| `KRKN_POD_LABEL` | Label selector for targeting pods (e.g. `service`) | `""` | — |
| `KRKN_NODE_LABEL` | Label selector for targeting nodes | `kubernetes.io/hostname` | — |
| `KRKN_SKIP_POD_NAME` | Pattern to skip specific pods | `""` | — |
| `KRKN_FITNESS_QUERY` | PromQL query for the fitness function used in genetic evolution of the scenario, e.g. `sum(kube_pod_container_status_restarts_total)` | `""` | — |
| `KRKN_SCENARIOS` | Comma-separated list of scenarios to enable | `node-cpu-hog,node-memory-hog` | — |
| `KRKN_GENERATIONS` | Genetic algorithm generations | `10` | — |
| `KRKN_POPULATION` | Population size per generation (minimum 2) | `5` | — |
| `KRKN_HEALTH_CHECK` | Health check endpoints in `name=url` format | `""` | — |
| `KRKN_TOP_SCENARIOS_COUNT` | Top scenarios to include in the analysis report | `10` | — |

### Choosing the Right CONFIGS

The `CONFIGS` parameter controls the OCM environment and provider settings:

| Cluster type | CONFIGS value |
|--------------|---------------|
| ROSA STS in **stage** | `rosa,sts,stage` (default) |
| ROSA STS in **int** | `rosa,sts,int` |

Each config name maps to a YAML file in the osde2e repository that sets viper keys:
- `rosa` → `provider: rosa`
- `sts` → `rosa.STS: true`
- `stage` → `ocm.env: stage`, `rosa.env: stage`
- `int` → `ocm.env: int`, `rosa.env: int`

## Job Execution Flow

The Jenkins job executes `scripts/parameterized-job-krknai.sh`, which:

```
Jenkins Agent
  │
  ├─ 1. Start rootless Podman socket (systemctl --user start podman.socket)
  ├─ 2. Create report/shared directories on host (/tmp/osde2e-report, /tmp/shared-dir)
  ├─ 3. podman create osde2e container
  │       - Mount Podman socket into container
  │       - Mount report/shared dirs as volumes
  │       - Pass Jenkins params + Vault secrets as env vars
  │       - Image: quay.io/redhat-services-prod/osde2e-cicada-tenant/osde2e:latest
  ├─ 4. podman start (runs: osde2e krkn-ai --configs $CONFIGS --skip-destroy-cluster --log-analysis-enable)
  │       │
  │       └─ Inside osde2e container:
  │            ├─ Validate config, connect to OCM
  │            ├─ Fetch kubeconfig for cluster
  │            ├─ Run krkn-ai discover (via mounted Podman socket → launches krkn-ai sub-container)
  │            ├─ Update discovered krkn-ai.yaml with user config (scenarios, generations, fitness query)
  │            ├─ Run krkn-ai execute (chaos scenarios via genetic algorithm)
  │            ├─ LLM analysis of results (if enabled)
  │            ├─ Must-gather (if not skipped)
  │            └─ Generate HTML report
  │
  ├─ 5. Copy report from /tmp/osde2e-report to ${WORKSPACE}/osde2e-report
  └─ 6. Jenkins archives osde2e-report/** as build artifacts
```

### Why Podman Socket?

The krkn-ai chaos engine runs as a separate container image (`quay.io/krkn-chaos/krkn-ai`). Since the osde2e binary itself runs inside a container, it needs access to a container runtime to launch the krkn-ai sub-container. The script mounts the host's rootless Podman socket into the osde2e container, allowing `podman run` commands inside to actually execute on the Jenkins agent.

## Vault Secrets

AWS credentials, OCM access, and the Gemini API key are automatically injected from Vault — no user configuration needed. See the [osde2e ops-sop](https://github.com/openshift/ops-sop/tree/master/v4/howto/osde2e) for Vault secret details.

## Build Artifacts

After each run, the following artifacts are archived from `osde2e-report/`:

| File | Description |
|------|-------------|
| `reports/all.csv` | Raw results from all scenario executions |
| `reports/health_check_report.csv` | Health check results per scenario |
| `report.html` | HTML summary report with top scenarios |
| `llm-analysis/` | LLM analysis output |
| `must-gather/` | OpenShift must-gather (if not skipped) |

## Build Retention

The last 15 builds are retained. Download artifacts from older builds before they are rotated out.

## Troubleshooting

### Job fails with "aws variables were not set"

The ROSA provider requires `AWS_REGION` to initialize the AWS session. The script passes `AWS_REGION="us-east-1"` as a default. If this error occurs, check that the environment variable is reaching the container.

### Job fails with "no container runtime found"

The Podman socket failed to start on the Jenkins agent. Check `systemctl --user status podman.socket` in the console output. This can happen if the agent doesn't have Podman installed or the user session is not properly initialized.

### Job fails with "discover mode failed"

- Verify the `CLUSTER_ID` exists in the OCM environment selected by `CONFIGS`
- A stage cluster requires `CONFIGS=rosa,sts,stage`; an int cluster requires `CONFIGS=rosa,sts,int`
- Check that the cluster is in a `ready` state

### Artifacts are empty

If artifacts are missing from the build, check whether the report directory path matches between the container and the publisher. The publisher archives `osde2e-report/**`.

## App-Interface Configuration

The Jenkins job is defined in two places in app-interface:

- **Job definition**: `data/services/osde2e/cicd/ci-int/jobs.yaml` — maps the job name to `scripts/parameterized-job-krknai.sh`
- **Job template**: `resources/jenkins/osde2e/job-templates.yaml` — defines parameters, Vault secrets, and publishers under the `osde2e-krknai-parameterized-job` template ID

To modify parameters, Vault secrets, or retention policy, update the job template in app-interface and submit an MR.

## Related

- [osde2e krkn-ai command documentation](Krkn-AI-Chaos-Testing.md)
- [krkn-ai documentation](https://krkn-chaos.dev/docs/krkn_ai/) (upstream)
- [krkn-ai container image](https://quay.io/repository/krkn-chaos/krkn-ai) (upstream)
- [osde2e ops-sop](https://github.com/openshift/ops-sop/tree/master/v4/howto/osde2e) (internal)
