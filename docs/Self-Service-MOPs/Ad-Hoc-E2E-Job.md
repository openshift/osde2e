# Ad-hoc E2E Test Run

> **Note:** For comprehensive testing instructions, see [Testing with OSDe2e](../testing-with-osde2e.md).

## Quick Reference

Ad-hoc testing allows you to run osde2e tests without modifying code, useful for validating new regions, instance types, or specific issues.

### Prerequisites

- Connect to Red Hat VPN
- Access to [Jenkins parameterized job](https://ci.int.devshift.net/blue/organizations/jenkins/osde2e-parameterized-job/activity)

### Quick Start

1. Navigate to https://ci.int.devshift.net/blue/organizations/jenkins/osde2e-parameterized-job/activity
2. Click "Login" (upper-right corner)
3. Click "Run" to create a new osde2e build
4. Configure parameters (see examples below)
5. Click "Run" to initiate the build

### Common Config Examples

**AWS:**
- ROSA sanity (stage): `rosa,stage,sanity`
- ROSA full e2e (prod): `rosa,prod,e2e-suite`
- Region: e.g., `me-central-1`
- Instance Type: e.g., `x1e.xlarge` (or leave empty for default)

**GCP:**
- GCP sanity (stage): `gcp,stage,sanity`
- GCP full e2e (prod): `gcp,prod,e2e-suite`
- Region: e.g., `southamerica-west1`
- Instance Type: e.g., `custom-8-32768` (or leave empty for default)
- OCM_CCS: Check for CCS (project: `osde2e-ccs`), uncheck for non-CCS

### Troubleshooting

- Review Jenkins build logs
- Contact [@sd-cicd-team](https://redhat-internal.slack.com/admin/user_groups) in [#sd-cicd](https://redhat-internal.slack.com/archives/CMK13BP4J)

## Additional Resources

- [Jenkins Job Definition](https://gitlab.cee.redhat.com/service/app-interface/-/blob/master/resources/jenkins/osde2e/job-templates.yaml)
- [Jenkins Job Link](https://ci.int.devshift.net/blue/organizations/jenkins/osde2e-parameterized-job/activity)

## Complete Documentation

For comprehensive instructions:
- [Testing with OSDe2e](../testing-with-osde2e.md) - Testing overview and guides
- [Running OSDe2e Tests](../run-osde2e-tests.md) - Running tests on existing clusters
- [Configuration Reference](../Config.md) - Environment variables and CLI flags
