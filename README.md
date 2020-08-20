# OSDe2e

[![GoDoc](https://godoc.org/github.com/openshift/osde2e?status.svg)](https://godoc.org/github.com/openshift/osde2e)

Comprehensive testing solution for Service Delivery

## Purpose
Provide a standard for testing every aspect of the Openshift Dedicated product. Use data derived from tests to inform release and product decisions.

## Setup 

Log into OCM, then go here to obtain an [OpenShift Offline Token].

A properly setup Go workspace using **Go 1.13+ is required**.

Install dependencies:
```
# Install dependencies
$ go mod tidy
# Copy them to a vendor dir
$ go mod vendor
```

Set OCM_TOKEN environment variable:
```
$ export OCM_TOKEN=<token from step 1>
```

## The `osde2e` command

The `osde2e` command is the root command that executes all functionality within the osde2e repo through a number of subcommands.

### Running tests locally

To run osde2e locally, first build the binary (do this after all changes) by running `make build`. The resulting binaries will be in `./out/`.

Once built, you can invoke osde2e by running `./out/osde2e` and osde2ectl by running `./out/osde2ectl`.

A common workflow is having a local script that combines these steps and the config. Example:

```bash
#!/usr/bin/env bash
make build

GINKGO_SKIP="" \
CLEAN_CHECK_RUNS="3" \
POLLING_TIMEOUT="5" \
OCM_TOKEN="[OCM token here]" \
./out/osde2e test --configs "prod,e2e-suite"
```

Please note: Do not commit or push any local scripts into osde2e.

## Configuration

There are many options to drive an osde2e run. Please refer to the [config package] for the most up to date config options. While golang, each option is well documented and includes the environment variable name for the option (where applicable.)

### Composable configs

OSDe2e comes with a number of [configs] that can be passed to the `osde2e test` command using the -configs argument. These can be strung together in a comma separated list to create a more complex scenario for testing.

```
$ osde2e test --configs prod,e2e-suite,conformance-suite
```

This will create a cluster on production (using the default version) that will run both the end to end suite and the Kubernetes conformance tests.

#### Using environment variables

Any config option can be passed in using environment variables. Please refer to the [config package] for exact environment variable names.

Example of spinning up a hosted-OSD instance and testing against it

```
OCM_TOKEN=$(cat ~/.ocm-token) \
OSD_ENV=prod \
CLUSTER_NAME=my-name-osd-test \
MAJOR_TARGET=4 \
MINOR_TARGET=2 \
osde2e test
``` 

These can be combined with the composable configs mentioned in the previous section as well.

```
OCM_TOKEN=$(cat ~/.ocm-token) \
MAJOR_TARGET=4 \
MINOR_TARGET=2 \
osde2e test --configs prod,e2e-suite
``` 

#### Using a custom YAML config

The composable configs consist of a number of small YAML files that can all be loaded together. Rather than use these built in configs, you can also elect to build your own custom YAML file and provide that using the `--custom-config` parameter.

```
osde2e test --custom-config ./osde2e.yaml
```

#### Via the command-line

Some configuration settings are also exposed as command-line parameters. A full list can be displayed by providing `--help` after the command.

An example is included below:

```shell
osde2e test --cluster-id 1ddkj9cr9j908gdlb1q5v6ld4b7ina5m \
    --provider stage \
    --skip-health-check \
    --focus-tests "RBAC Operator"
```

##### Full custom YAML config example
```
dryRun: false
cluster:
 name: jsica-test
 multiAZ: true
ocm:
 debug: false
 token: [Redacted]
 env: stage
tests:
 testsToRun:
 - '[Suite: e2e]'
```

#### Order of precedence

Config options are currently parsed by loading defaults, attempting to load environment variables, attempting to load composable configs, and finally attempting to load config data from the custom YAML file. There are instances where you may want to have most of your config in a custom YAML file while keeping one or two sensitive config options as environment variables (OCM Token)

### Testing against non OSD clusters

It is possible to test against non-OSD clusters by specifying a kubeconfig to test against.

```
TEST_KUBECONFIG=~/.kube/config \
osde2e test --configs prod --custom-config .osde2e.yaml
```
*Note: You must skip certain Operator tests that only exist in a hosted OSD instance. This can be skipped by skipping the operators test suite.*

## Different Test Types
Core tests and Operator tests reside within the OSDe2e repo and are maintained by the CICD team. The tests are written and compiled as part of the OSDe2e project. 
* Core Tests
* OpenShift Conformance
* Scale testing
* OC Must Gather
* Verify 
  * All pods are healthy or successful
  * ImageStreams exist
  * Project creation possible
  * Ingress to console possible
* [Operator tests]

Third-party (Addon) tests are built as containers that spin up and report back results to OSDe2e. These containers are built and maintained by external groups looking to get CI signal for their product within OSD. The definition of a third-party test is maintained within the `managed-tenants` repo and is returned via the Add-Ons API.

For more information please see the [Addon Testing Guide]

## Operator Testing
Much like the different phases of operators laid out on OperatorHub, Operator tests using OSDe2e falls under one of a few categories:

**Basic Testing**
This type of test in OSDe2e affirms that the operator and dependent objects are installed, running, and configured correctly in a cluster. This level of testing is the simplest to implement but should not be targeted long-term.

**Intermediate Testing**
Flexing the actual purpose of the Operator. For example, if the operator created a database, actually testing functionality by creating a “dbcrd” object and verifying a new database spins up correctly. This should be the standard level of testing for most operators.

**Advanced Testing**
Collecting metrics of the above tests as well as testing recovery of failures. Example: If the pod(s) the operator runs gets deleted, what happens? If the pods created by the operator get deleted does it recover? Testing at this level should be able to capture edge-cases even in the automated CI runs. It involves significant up front development and therefore is not likely the primary target of operator authors.

## Anatomy Of A Test Run
There are several conditional checks (is this an upgrade test, is this a dry-run) that may impact what stages an OSDe2e run may contain, but the most complicated is an upgrade test:

1. Load Config
2. Provision Cluster (If Cluster ID or Kubeconfig not provided)
3. Verify Cluster Integrity
4. Run Tests (pre-upgrade)
5. Capture logs, metrics, and metadata to the `REPORT_DIR`
6. Upgrade Cluster
7. Verify Cluster Integrity
8. Run Tests (post-upgrade)
9. Capture logs, metrics, and metadata to the `REPORT_DIR`

With a dry-run, OSDe2e only performs the “Load Config” step and outputs the parameters the run would have used. With a vanilla-install run (not an upgrade test) steps 6-9 are skipped and the entire upgrade phase does not occur.

A failure at any step taints and fails the run. 

## Reporting / Alerting
Every run of OSDe2e captures as much data as possible. This includes cluster and pod logs, prometheus metrics, and test info. In addition to cluster-specific info, the version of hive and OSDe2e itself is captured to identify potential flakes or environment failures. Every test suite generates a `junit.xml` file that contains test names, pass/fails, and the time the test segment took. It is expected that addon testing will follow this pattern and generate their own `junit.xml` file for their test results. 

The `junit.xml` files are converted to meaningful metrics and stored in DataHub. These metrics are then published via [Grafana dashboards] used by Service Delivery as well as Third Parties to monitor project health and promote confidence in releases. Alerting rules are housed within the DataHub Grafana instance and addon authors can maintain their own individual dashboards.

## Writing tests
To write your own test, see [Writing Tests].

[OpenShift Offline Token]:https://cloud.redhat.com/openshift/token
[configs]:/configs/
[config package]:/pkg/common/config/config.go
[Makefile]:/Makefile
[Operator tests]:/pkg/e2e/operators/
[Addon Testing Guide]:/docs/Addons.md
[Grafana dashboards]:https://grafana.datahub.redhat.com/dashboard/db/osd-health-metrics?orgId=1
[Writing Tests]:/docs/Writing-Tests.md
