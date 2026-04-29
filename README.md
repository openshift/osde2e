# OSDE2E Documentation Summary

## Table of Contents

- [Overview](#overview)
- [Core Components](#core-components)
    - [1. Testing Framework](#1-testing-framework-osde2e-testingmd)
    - [2. Test Execution Methods](#2-test-execution-methods)
        - [A. Ginkgo Test Images](#a-ginkgo-test-images-test-harnessesmd)
        - [B. Writing Tests](#b-writing-tests-writing-testsmd)
    - [3. Periodic Jobs](#3-ci-jobs-ci-jobsmd)
    - [5. Self-Service Operations](#5-self-service-operations)
        - [Gap Analysis Testing](#gap-analysis-testing-adhoc-osde2e-testingmd-ad-hoc-e2e-jobmd)
        - [Instance Type Enablement](#instance-type-enablement-instance-type-enablementmd)
        - [Region Enablement](#region-enablement-region-enablementmd)
        - [Testgrid Pipeline Integration](#testgrid-pipeline-integration-adding-testgrid-pipelines-through-ci-operatormd)
    - [6. Running Tests](#6-running-tests)
        - [Local Testing](#local-testing-run-osde2e-testsmd-testing-with-osde2emd)
    - [7. Slack Notifications](#7-slack-notifications)
    - [8. Configuration](#8-configuration-configmd)
- [Workflow Summary](#workflow-summary)
    - [Typical Test Flow](#typical-test-flow)
    - [Integration Points](#integration-points)
- [Key Resources](#key-resources)

## Overview

OSDE2E (OpenShift Dedicated End-to-End) is a comprehensive test framework for qualifying new versions of OpenShift in managed environments. It facilitates testing for Managed OpenShift platforms (OSD, ROSA, ROSA HCP, ARO), OSD Operators, and Addons.

OSDE2E integrates with OpenShift's CI/CD infrastructure to provide continuous validation of new OpenShift releases and serves as a critical actor in the release gating process. The framework supports various testing scenarios from ad-hoc testing to automated periodic validation, with robust cluster provisioning, artifact collection and reporting capabilities.

## Core Components

### 1. Testing Framework (OSDE2E-Testing.md)

**Primary Use Cases:**
- Managed OpenShift platforms (OSD, ROSA, ROSA HCP)
- OSD Operators running on Managed OpenShift
- Addons integration testing with OpenShift versions

Test results serve as gating signals for promotion between environments.

### 2. Test Execution Methods

#### A. Ginkgo Test Images (Test-Harnesses.md)
Standalone Ginkgo e2e test images run on test pods. Three types available:

**Operator Ad-hoc Test Image:**
- Uses `openshift/golang-osd-operator-osde2e` boilerplate
- Test structure in operator repo under `/test/e2e/`
- Automated publishing via CI/CD pipelines
- Integrated with Prow for automated testing

**Addon Test Harness:**
- For OpenShift addon components
- Requires `ADDON_IDS` environment parameter
- Addon installation before test execution

#### B. Writing Tests (Writing-Tests.md)

**Best Practices:**
- Follow Kubernetes best practices guide for e2e tests
- Use Ginkgo and Gomega frameworks
- Leverage osde2e-common module to reduce code duplication
- Use e2e-framework for cluster interfacing
- Apply labels/tags for test classification
- Focus test cases on specific scope
- Cover both positive and negative cases
- Include proper error messages for debugging

**Example Repositories:**
- Managed Upgrade Operator Tests
- OCM Agent Operator Tests
- RBAC Permissions Operator Tests

### 3. CI Jobs (CI-Jobs.md)

**SD CICD Periodic Jobs:**
- ROSA BYOVPC Proxy Install/Post Install
- OSD AWS Upgrade suites (Y-1 to Y, Z-1 to Z, Y to Y+1)
- OSD AWS SREP Operator Informing Suite

**TRT Nightly Periodic Jobs:**
- Validates Managed OpenShift for new nightly OCP builds
- Provides informing signal to releases
- Covers OSD (AWS/GCP) and ROSA (Classic STS/HCP)
- Supports OCP versions 4.10-4.14


**Adding Jobs:**
- PR to release repo for periodic job
- PR to continuous-release-jobs repo for signal notification
- Auto-included after 24 hours
  **Removing Jobs:**
- PR to release repo to remove job
- PR to continuous-release-jobs repo to remove signal


These jobs send alerts to #hcm-cicd-alerts Slack channel.


### 5. Self-Service Operations

#### Gap Analysis Testing (adhoc-osde2e-testing.md, Ad-Hoc-E2E-Job.md)
- Jenkins parameterized job for on-demand testing
- Available at ci.int.devshift.net
- Supports AWS and GCP testing
- Custom configuration via environment variables
- Useful for region/instance type enablement

#### Instance Type Enablement (Instance-Type-Enablement.md)

**Prerequisites:**
- Instance type enabled in Stage via OCM/ROSA CLIs
- Quota verification
- Pricing enablement

**Process:**
1. Create PR in release repo for osde2e jobs
2. Configure job with new instance type
3. Run `make jobs` to generate prowgen jobs
4. Merge and monitor results in Prow
5. Validate with 3 consecutive successful runs

#### Region Enablement (Region-Enablement.md)

**Prerequisites:**
- Region enabled in AWS account
- SDA team enables region for ocm account

**Common Issues:**
- AMI availability errors (report to BU)
- Quota errors (request quota increase)

**Process:**
- Create periodic Prow job
- Or run ad-hoc Jenkins job
- Follow region enablement SOP

#### Testgrid Pipeline Integration (Adding-Testgrid-Pipelines-Through-Ci-Operator.md)
- Integration with ci-operator for testgrid pipelines
- Jobs added to redhat-openshift-osd dashboard
- Custom 'osde2e' tag for identification
- Must be prowgen job
- Add to _allow-list.yaml in release repo
- Auto-updates every 24 hours

### 6. Running Tests

#### Local Testing (run-osde2e-tests.md, testing-with-osde2e.md)

**On Existing Cluster:**
```bash
# OCM credentials
export OCM_CLIENT_ID=<id>
export OCM_CLIENT_SECRET=<secret>

# AWS credentials (for ROSA/OSD on AWS)
export AWS_ACCESS_KEY_ID=<access-key>
export AWS_SECRET_ACCESS_KEY=<secret-key>
export AWS_REGION=<region>  # e.g., us-east-1

# Optional: For BYO-VPC clusters
export AWS_VPC_SUBNET_IDS=<subnet-ids>

# Optional: For ROSA using AWS profile
export AWS_PROFILE=<profile-name>

./osde2e test --cluster-id ${CLUSTERID} --configs rosa,e2e-suite,stage
```

**Cluster Upgrades:**
```bash
# OCM and AWS credentials (same as above)
export OCM_CLIENT_ID=<id>
export OCM_CLIENT_SECRET=<secret>
export AWS_ACCESS_KEY_ID=<access-key>
export AWS_SECRET_ACCESS_KEY=<secret-key>

# Upgrade configuration
export CLUSTER_ID=<cluster-id>
export UPGRADE_MANAGED=true
export UPGRADE_TO_LATEST_Z=true  # Or UPGRADE_TO_LATEST, UPGRADE_TO_LATEST_Y, UPGRADE_RELEASE_NAME

ocm login --url stg   
./osde2e test --cluster-id ${CLUSTER_ID} --configs rosa,e2e-suite,stage
```

**Key Points:**
- Supports OCM-driven upgrades via managed-upgrade-operator
- Verify cluster health pre and post-upgrade
- Specify target version with `UPGRADE_RELEASE_NAME` or use latest flags

**On New Cluster:**
```bash
# OCM credentials
export OCM_CLIENT_ID=<id>
export OCM_CLIENT_SECRET=<secret>

# AWS credentials
export AWS_ACCESS_KEY_ID=<access-key>
export AWS_SECRET_ACCESS_KEY=<secret-key>
export AWS_REGION=<region>

# Optional cluster configuration
export CLUSTER_VERSION=openshift-v4.14.0

./osde2e test --configs rosa,e2e-suite,stage
```

### 7. Slack Notifications

OSDe2e can send AI-powered failure analysis to Slack when tests fail. Each test suite can notify a different Slack channel with failure details, analysis, and logs.

#### Setup

**1. Get Your Channel ID**

Right-click your channel → **View channel details** → copy the channel ID (starts with `C`, e.g., `C06HQR8HN0L`)

**2. Configure Test Suites**

Set `TEST_SUITES_YAML` with your test images, webhook URLs, and Slack channel IDs:

```bash
export TEST_SUITES_YAML='
- image: quay.io/openshift/osde2e-tests:latest
  slackWebhook: https://hooks.slack.com/workflows/T.../A.../...
  slackChannel: C06HQR8HN0L
- image: quay.io/openshift/custom-tests:v1.0
  slackWebhook: https://hooks.slack.com/workflows/T.../B.../...
  slackChannel: C07ABC123XY
'
```

**3. Enable Notifications**

Enable Slack notifications in your config:

```yaml
tests:
  enableSlackNotify: true
logAnalysis:
  enableAnalysis: true
```

#### What You'll Receive

When tests fail, you'll get a threaded Slack message with:
1. **Main message**: Test suite info (what failed)
2. **Reply 1**: AI analysis (why it failed)
3. **Reply 2**: Links to persisted logs and junit results (evidence)
4. **Reply 3**: Cluster details (for debugging)

For implementation details, see [internal/reporter/README.md](../internal/reporter/README.md).

### 8. Configuration (Config.md)

**Environment Variables:**

**Cluster Related:**
- CLUSTER_ID, OSD_ENV, CLOUD_PROVIDER_ID
- CLOUD_PROVIDER_REGION, CLUSTER_VERSION
- SKIP_DESTROY_CLUSTER, MULTI_AZ

**ROSA Specific:**
- ROSA_ENV, ROSA_STS, ROSA_REPLICAS

**Hypershift:**
- Hypershift (boolean for HostedCluster)

**OCM:**
- OCM_COMPUTE_MACHINE_TYPE, OCM_CCS
- OCM_FLAVOUR, OCM_ADDITIONAL_LABELS

**Upgrade:**
- UPGRADE_TO_LATEST, UPGRADE_TO_LATEST_Z/Y
- UPGRADE_RELEASE_NAME, UPGRADE_IMAGE

**Test Execution:**
- GINKGO_SKIP, GINKGO_FOCUS (only for monorepo tests)
- ADDON_IDS_AT_CREATION, ADDONS_IDS

**Command Line Flags:**
- `--cluster-id`: Test existing cluster
- `--configs`: Comma-separated built-in configs
- `--skip-destroy-cluster`: Retain cluster after test
- `--skip-health-check`: Skip health checks
- `--skip-tests`: Skip matching tests

**Config Values:**
- Environments: int, stage, prod
- Providers: aws, gcp, ocm, rosa
- Test Suites: e2e-suite, informing-suite, openshift-suite
- Special: dry-run, skip-health-checks, upgrade-to-latest

## Workflow Summary

### Typical Test Flow
1. **Setup**: Configure environment variables or use built-in configs
2. **Cluster Provisioning**: Create new or use existing cluster
3. **Health Checks**: Validate cluster health (optional)
4. **Test Execution**: Run Ginkgo test suites via ad-hoc test images
5. **Metrics Collection**: Send results to Prometheus
6. **Upgrade Testing**: Optional cluster upgrade validation
7. **Cleanup**: Delete cluster (optional) and collect must-gather

### Integration Points
- **Prow**: Automated CI/CD testing
- **Testgrid**: Result visualization
- **Prometheus**: Metrics storage and querying
- **OCM**: Cluster lifecycle management
- **TRT**: OpenShift release gating

## Key Resources

**Documentation:**
- Ad-hoc Test Image Example: github.com/openshift/osde2e-example-test-harness
- OSDE2E Common: github.com/openshift/osde2e-common
- Release Repo: github.com/openshift/release
- E2E Framework: github.com/kubernetes-sigs/e2e-framework

**Dashboards:**
- Jenkins jobs: https://ci.int.devshift.net/view/osde2e/
- Progressive delivery rollouts: https://inscope.corp.redhat.com/catalog
- SD CICD TestGrid: testgrid.k8s.io/redhat-openshift-osd
- Prow: prow.ci.openshift.org
- Prometheus: prometheus.app-sre-prod-01.devshift.net

**Communication:**
- Slack: #hcm-delivery