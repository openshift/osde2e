---
name: dashboard-dev
description: Guide for contributing to and deploying the Delivery Dashboard
allowed-tools: [Bash, Read, Grep, Glob, Write, Edit, TodoWrite]
---

# Delivery Dashboard Development Skill

## Purpose

Help developers contribute to, run locally, and deploy the Delivery Dashboard — a web UI showing operator pipeline status across stage and integration environments, backed by SQLite, SQS, and S3.

---

## Getting Started: Fork the Source Branch

The dashboard lives on the `feat/delivery-dashboard` branch of:
```
https://github.com/ritmun/osde2e
```

Fork that repo on GitHub, then:

```bash
git clone git@github.com:<your-github-username>/osde2e.git
cd osde2e
git remote add upstream git@github.com:ritmun/osde2e.git
git fetch upstream
git checkout -b feat/delivery-dashboard upstream/feat/delivery-dashboard
```

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
  deploy.sh                # full deploy to OpenShift cluster
  run-local.sh             # run locally
  verify-build.sh          # sanity check binary + templates
```

---

## Local Development

Build:
```bash
GOFLAGS="-mod=mod" go build -o out/osde2e ./cmd/osde2e/
```

Run locally against a SQLite file:
```bash
./out/osde2e dashboard --db=./dashboard.db --port=8080
```

Open: http://localhost:8080/dashboard/deliverables

Or use the local script:
```bash
./scripts/dashboard/run-local.sh
```

---

## Environment Policy

> **IMPORTANT: Development and testing must only target stage/non-production clusters.**
>
> The cluster `rh-hp-delivery` is **production**. Deployments to it are handled exclusively by the CI/CD pipeline — never manually.
>
> Use a personal or stage OpenShift cluster for all dev/test work.

---

## Deploying to Your Own OpenShift Cluster

### Prerequisites

- `oc` CLI installed and logged in: `oc login <cluster-url>`
- Cluster must be able to pull from `registry.access.redhat.com` (UBI images)
- Two secrets pre-created in the `delivery-dashboard` namespace:
  - `aws-credentials` — keys: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`
  - `ocm-credentials` — keys: `OCM_CLIENT_ID`, `OCM_CLIENT_SECRET`
- An SQS queue URL receiving S3 event notifications for osde2e log uploads

### Create secrets

```bash
oc new-project delivery-dashboard 2>/dev/null || oc project delivery-dashboard

oc create secret generic aws-credentials \
  --from-literal=AWS_ACCESS_KEY_ID=<your-key-id> \
  --from-literal=AWS_SECRET_ACCESS_KEY=<your-secret>

oc create secret generic ocm-credentials \
  --from-literal=OCM_CLIENT_ID=<your-client-id> \
  --from-literal=OCM_CLIENT_SECRET=<your-client-secret>
```

### Deploy

```bash
SQS_QUEUE_URL=https://sqs.us-east-1.amazonaws.com/<account>/<queue> \
  ./scripts/dashboard/deploy.sh
```

The script:
1. Builds `osde2e` binary for `linux/amd64`
2. Builds a container image inside the cluster via OpenShift BuildConfig (`dashboard.Dockerfile`)
3. Applies ConfigMap, Deployment (emptyDir + RollingUpdate), Service, and Route manifests
4. Waits for rollout
5. Prints the dashboard URL

Route is named `live` so URL will be:
```
https://live-delivery-dashboard.apps.<your-cluster-domain>/dashboard/deliverables
```

### When to rebuild vs re-apply

| Change type | Action needed |
|-------------|--------------|
| Go source / templates | Re-run `deploy.sh` (new build + rollout) |
| ConfigMap / env vars | `oc apply` the manifest only, pod restarts automatically |
| Route / Service | `oc apply` the manifest only, no restart needed |

---

## Common Development Tasks

- **Add a new page**: create template in `server/templates/`, add handler in `server.go`, register route in `setupRoutes()`, add nav link in `base.html`
- **Add a data query**: add method to `store/store.go`, add model to `models/types.go`
- **Change nav highlighting**: set `ActivePage` key in handler's data map, match it in `base.html`
- **Check logs**: `oc logs -f deployment/delivery-dashboard -n delivery-dashboard`
- **Check pod status**: `oc get pods -n delivery-dashboard`
- **Trigger new build**: `oc start-build delivery-dashboard -n delivery-dashboard --follow`
- **Rolling restart**: `oc rollout restart deployment/delivery-dashboard -n delivery-dashboard`

---

## Architecture

- **Ingestion**: SQS listener polls for S3 event notifications; each event points to a test result JSON in S3, downloaded and parsed into `pipeline_runs` SQLite table
- **Backfill**: on startup with `--backfill`, server scans S3 bucket directly for historical results (~5s typical)
- **LLM analysis**: stored in `llm_analysis` column as JSON; parsed to extract `root_cause` and `recommendations`
- **Storage**: single SQLite file at `/data/dashboard.db`, mounted via `emptyDir` (repopulated from S3 on each start)
- **Templates**: standard Go `html/template`, server-side rendered, no JS framework