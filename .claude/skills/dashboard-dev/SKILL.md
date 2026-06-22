---
name: dashboard-dev
description: Guide for contributing to and deploying the Delivery Dashboard
allowed-tools: [Bash, Read, Grep, Glob, Write, Edit, TodoWrite]
---

# Delivery Dashboard Development Skill

## Purpose

Help developers contribute to, run locally, and deploy the Delivery Dashboard — a web UI showing operator pipeline status across stage and integration environments, backed by SQLite, SQS, and S3.

---

## Codebase Layout

```
pkg/dashboard/
  models/types.go          # data models (PipelineRun, FailureGroup, etc.)
  store/store.go           # SQLite queries
  server/server.go         # HTTP handlers and routes
  server/templates/        # Go HTML templates
    base.html              # nav, layout
    operators.html         # deliverables/pipelines page
    pipeline-detail.html   # per-operator history
    analysis.html          # failure grouping by AI root cause
    usage.html             # infra/clusters page
cmd/osde2e/dashboard/      # CLI entry point (flags, wiring)
scripts/dashboard/
  deploy.sh                # local dev deploy to OpenShift cluster
  verify-build.sh          # sanity check binary + templates
configs/local/
  dashboard-build/         # podman build context (Dockerfile committed, binary gitignored)
```

Manifests live in the adjacent **hp-delivery-apps** repo:
```
delivery-dashboard/
  base/                    # Deployment + Service
  overlays/
    local/                 # personal dev cluster (gitignored, manually provisioned secrets)
    stage/                 # vault ExternalSecrets
    prod/                  # vault ExternalSecrets
```

---

## Local Development (native, no container)

```bash
make dashboard
```

Builds the binary and runs it at http://localhost:8080/dashboard/deliverables against `./dashboard.db`.


## Deploying to Your Own OpenShift Cluster

### Prerequisites

- `podman login quay.io`
- `oc login <cluster-url>`
- hp-delivery-apps repo cloned adjacent to this repo
- Secrets pre-created in the target namespace (see hp-delivery-apps/delivery-dashboard/README.md)

### Create secrets (local overlay — vault handles stage/prod automatically)

```bash
oc create secret generic osde2e-ocm-credentials \
  --from-literal=ocm-client-id=<id> \
  --from-literal=ocm-client-secret=<secret> \
  -n <namespace>

oc create secret generic osde2e-aws-credentials \
  --from-literal=aws-access-key-id=<key> \
  --from-literal=aws-secret-access-key=<secret> \
  -n <namespace>
```

### Set SQS_QUEUE_URL

Edit `hp-delivery-apps/delivery-dashboard/overlays/local/configmap.yaml` directly — it is gitignored.

### Deploy

```bash
DASHBOARD_QUAY_IMAGE=quay.io/<your-username>/delivery-dashboard:latest \
  QUAY_EXPIRE=26w \
  ./scripts/dashboard/deploy.sh
```

The script:
1. Checks required secrets exist (fails fast if not)
2. Compiles linux/amd64 binary → `configs/local/dashboard-build/osde2e`
3. Builds slim image via podman and pushes to quay
4. Applies `kustomize build overlays/local | oc apply`
5. Waits for rollout, prints URL

Route URL: `https://live-<namespace>.apps.<cluster-domain>/dashboard/deliverables`

### When to rebuild vs re-apply

| Change type | Action |
|-------------|--------|
| Go source / templates | Re-run `deploy.sh` |
| ConfigMap / env vars | Edit overlay configmap, `kustomize build \| oc apply -f -` |
| Route / Service | Same as above, no restart needed |

---

## Common Development Tasks

- **Add a new page**: template in `server/templates/`, handler in `server.go`, route in `setupRoutes()`, nav link in `base.html`
- **Add a data query**: method in `store/store.go`, model in `models/types.go`
- **Check logs**: `oc logs -f deployment/delivery-dashboard -n <namespace>`
- **Check pod status**: `oc get pods -n <namespace>`
- **Rolling restart**: `oc rollout restart deployment/delivery-dashboard -n <namespace>`

---

## Architecture

- **Pipeline data**: SQS listener polls for S3 event notifications; each event points to a test result JSON, downloaded and parsed into `pipeline_runs` SQLite table
- **Pipeline Backfill**: on startup with `--backfill`, scans S3 bucket directly for historical results
- **Pipeline LLM analysis**: stored in `llm_analysis` column as JSON; parsed to extract `root_cause` and `recommendations`
- **OCM data**: collectors query OCM API for cluster reserves, usage metrics, and environment status (stage/int/prod)
- **Local Storage**: single SQLite file at `/data/dashboard.db`, mounted via `emptyDir` (repopulated from S3 + OCM on each start)
- **UI Templates**: standard Go `html/template`, server-side rendered, no JS framework