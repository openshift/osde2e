# OSDe2e
 
[![GoDoc](https://godoc.org/github.com/openshift/osde2e?status.svg)](https://godoc.org/github.com/openshift/osde2e)
 
Comprehensive testing solution for Service Delivery
 
## Purpose
Provide a standard for testing every aspect of the Openshift Dedicated product. Use data derived from tests to inform release and product decisions.
 
## Setup
 
Log into OCM, then go here to obtain an [OpenShift Offline Token].
 
A properly setup Go workspace using **Go 1.18+ is required**.
 
Install dependencies:
```
# Install dependencies
$ go mod tidy
```
 
Set OCM_TOKEN environment variable:
```
$ export OCM_TOKEN=<token from step 1>
```
 
## The `osde2e` command
 
The `osde2e` command is the root command that executes all functionality within the osde2e repo through a number of subcommands.
 
### Running from source
 
To run osde2e locally, first build the binary (do this after all changes) by running `make build`. The resulting binaries will be in `./out/`.
 
Once built, you can invoke osde2e by running `./out/osde2e`.
 
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
 
Another example:
```bash
#!/usr/bin/env bash
make build
 
OSD_ENV="prod" \
CLOUD_PROVIDER_ID="aws" \
CLOUD_PROVIDER_REGION="us-east-1" \
CLUSTER_VERSION="openshift-v4.6.0-fc.3-fast" \
CLUSTER_EXPIRY_IN_MINUTES="120" \
OCM_TOKEN="[OCM token here]" \
./out/osde2e test --configs "e2e-suite"
```
 
Please note: Do not commit or push any local scripts into osde2e.
 
### Running the latest docker image
 
The following command would help in running the latest docker image.
 
To run the latest docker image:
 
```
#!/usr/bin/env bash
 
docker run -e
-e OSD_ENV="prod" \
-e CLOUD_PROVIDER_ID="aws" \
-e CLOUD_PROVIDER_REGION="us-east-1" \
-e CLUSTER_VERSION="openshift-v4.6.0-fc.3-fast" \
-e CLUSTER_EXPIRY_IN_MINUTES="120" \
-e OCM_TOKEN="[OCM token here]" \
quay.io/app-sre/osde2e test --configs "e2e-suite"
```

### Running via a local Kubeconfig

By default, osde2e will try to obtain Kubeconfig admin credentials for the cluster by calling OCM's [credentials](https://api.openshift.com/#/default/get_api_clusters_mgmt_v1_clusters__cluster_id__credentials) API.

Permission to use that API is dependent upon a user's role in OCM. This will be noticable if you encounter the following error:

```
could not get kubeconfig for cluster: couldn't retrieve credentials for cluster '$CLUSTERID'
```

In this situation, you can override the credentials fetch by using a locally-sourced Kubeconfig:

- Log in to the cluster you wish to test against, to update your kubeconfig. 
- Many tests require elevated permissions. Elevate to be a member of a cluster-admin group.
- Set the `TEST_KUBECONFIG` environment variable to the path of your kubeconfig.
- Run osde2e as usual.

A full example of this process is presented below:

```bash
$ ocm cluster login <cluster>
$ oc adm groups add-users osd-sre-cluster-admins $(oc whoami)
$ export TEST_KUBECONFIG=$HOME/.kube/config
$ export OCM_TOKEN=${YOUR_OCM_TOKEN_HERE}
$ osde2e test --configs e2e-suite,stage --skip-health-check
```

## Configuration
 
There are many options to drive an osde2e run. Please refer to the [config package] for the most up to date config options. While golang, each option is well documented and includes the environment variable name for the option (where applicable.)
 
### Composable configs
 
OSDe2e comes with a number of [configs] that can be passed to the `osde2e test` command using the --configs argument. These can be strung together in a comma separated list to create a more complex scenario for testing.
 
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

A list of commonly used environment variables are included in [Config variables].


#### Using a custom YAML config
 
The composable configs consist of a number of small YAML files that can all be loaded together. Rather than use these built in configs, you can also elect to build your own custom YAML file and provide that using the `--custom-config` parameter.
 
```
osde2e test --custom-config ./osde2e.yaml
```

The custom config below is a basic example for deploying a ROSA STS cluster and running
all of the OSD operators tests that do not have the informing label associated to them.

```
dryRun: false
provider: rosa
cloudProvider:
  providerId: aws
  region: us-east-1
rosa:
  env: stage
  STS: true
cluster:
  name: osde2e
tests:
  ginkgoLabelFilter: Operators && !Informing
```

A list of existing config files that can be used are included in [Config variables].
 
#### Via the command-line
 
Some configuration settings are also exposed as command-line parameters. A full list can be displayed by providing `--help` after the command.
 
An example is included below:
 
```
osde2e test --cluster-id 1ddkj9cr9j908gdlb1q5v6ld4b7ina5m \
   --provider stage \
   --skip-health-check \
   --focus-tests "RBAC Operator"
```

Optionally, you may skip cluster health check, must gather, as follows. (Using ./out/osde2e binary created from `make build`)

```
POLLING_TIMEOUT=1     ./out/osde2e test --cluster-id=$CLUSTER_ID  --configs stage --must-gather=False --skip-health-check  --focus-tests="rh-api-lb-test"
```
 
A list of commonly used CLI flags are included in [Config variables].

#### Order of precedence
 
Config options are currently parsed by loading defaults, attempting to load environment variables, attempting to load composable configs, and finally attempting to load config data from the custom YAML file. There are instances where you may want to have most of your config in a custom YAML file while keeping one or two sensitive config options as environment variables (OCM Token)

### Testing against non OSD clusters
 
It is possible to test against non-OSD clusters by specifying a kubeconfig to test against.
 
```
PROVIDER=mock \
TEST_KUBECONFIG=~/.kube/config \
osde2e test --configs prod --custom-config .osde2e.yaml
```
*Note: You must skip certain Operator tests that only exist in a hosted OSD instance. This can be skipped by skipping the operators test suite.*
 
## Tests

### Selecting Tests To Run

OSDe2e supports a couple different ways you can select which tests you would like to run. Below presents
the commonly used methods for this:

1. Using the label filter. Labels are ginkgos way to tag test cases. The examples below
   will tell osde2e to run all tests that have the `E2E` label applied.

```
# Command line option
osde2e test --label-filter E2E

# Passed in using a custom config file
tests:
  ginkgoLabelFilter: E2E
```

2. Using focus strings. Focus strings are ginkos way to select test cases based on string regex.

```
# Command line option
osde2e test --focus-tests "OCM Agent Operator"

# Custom config file
tests:
  focus: "OCM Agent Operator"
```

3. Using a combination of labels and focus strings to fine tune your test selection.
   The examples below tell osde2e to run all ocm agent operator tests and avoid running
   the upgrade test case.

```
# Command line options
osde2e test --label-filter "Operators && !Upgrade" --focus-tests "OCM Agent Operator"

# Custom config file
tests:
  ginkgoLabelFilter: "Operators && !Upgrade"
  focus: "OCM Agent Operator"
```

### Test Types

OSDe2e currently holds all core and operator specific tests and are maintained by the CICD team.
Test types range from core OSD verification, OSD operators to scale/conformance.

### Writing Tests

Refer to the [Writing Tests] document for guidelines and standards.
 
Third-party (Addon) tests are built as containers that spin up and report back results to OSDe2e. These containers are built and maintained by external groups looking to get CI signal for their product within OSD. The definition of a third-party test is maintained within the `managed-tenants` repo and is returned via the Add-Ons API.
 
For more information please see the [Addon Testing Guide]
 
### Operator Testing

Much like the different phases of operators laid out on OperatorHub, Operator tests using OSDe2e falls under one of a few categories:
 
**Basic Testing**
This type of test in OSDe2e affirms that the operator and dependent objects are installed, running, and configured correctly in a cluster. This level of testing is the simplest to implement but should not be targeted long-term.
 
**Intermediate Testing**
Flexing the actual purpose of the Operator. For example, if the operator created a database, actually testing functionality by creating a “dbcrd” object and verifying a new database spins up correctly. This should be the standard level of testing for most operators.
 
**Advanced Testing**
Collecting metrics of the above tests as well as testing recovery of failures. Example: If the pod(s) the operator runs gets deleted, what happens? If the pods created by the operator get deleted does it recover? Testing at this level should be able to capture edge-cases even in the automated CI runs. It involves significant up front development and therefore is not likely the primary target of operator authors.
 
### Anatomy Of A Test Run

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

### CI/CD Job Results Database

We have provisioned an AWS RDS Postgres database to store information about our CI jobs and the tests that they execute. We used to store our data only within prometheus, but prometheus's timeseries paradigm prevented us from being able to express certain queries (even simple ones like "when was the last time this test failed").

The test results database (at time of writing) stores data about each job and its configuration, as well as about each test case reported by the Junit XML output of the job.

This data allows us to answer questions about frequency of job/test failure, relationships between failures, and more. The code responsible for managing the database can be found in the [`./pkg/db/`](https://github.com/openshift/osde2e/tree/cfd38c75532274d619840ad505c1232881eb417a/pkg/db) directory, along with a README describing how to develop against it.

#### Database usage from OSDe2e

Because `osde2e` runs a a cluster of parallel, ephemeral prow jobs, our usage of the database is unconventional. We have to write all of our database interaction logic with the understanding that any number of other prow jobs could be modifying the data at the same time that we are.

We use the database to generate alerts for the CI Watcher to use, and we follow this algorithm to generate those alerts safely in our highly-concurrent usecase (at time of writing, implemented [here](https://github.com/openshift/osde2e/blob/cfd38c75532274d619840ad505c1232881eb417a/pkg/e2e/e2e.go#L1029)):

1. At the end of each job, list all testcases that failed during the current job. Implemented by [`ListAlertableFailuresForJob`](https://github.com/openshift/osde2e/blob/cfd38c75532274d619840ad505c1232881eb417a/pkg/db/queries/queries.sql#L66).
1. Generate a list of testcases (in any job) that have failed more than once during the last 48 hours. Implemented by [`ListProblematicTests`](https://github.com/openshift/osde2e/blob/cfd38c75532274d619840ad505c1232881eb417a/pkg/db/queries/queries.sql#L105).
1. For each failing testcase in the current job, create a PD alert if the testcase is one of those that have failed more than once in the last 48 hours.
1. After generating all alerts as above, merge all pagerduty alerts that indicate failures for the same testcase (this merge uses the title of the alert, which is the testcase name, to group the alerts).
1. Finally, close any PD incident for a testcase that does not appear in the list of testcases failing during the last 48 hours.

> Why does each job only report its own failures? The database is global, and a single job could report for all of them.

If each job reported for the failures of all recent jobs, we'd create an enormous number of redundant alerts for no benefit. Having each job only report its own failures keeps the level of noise down _without_ requiring us to build some kind of concensus mechanism between the jobs executing in parallel.

> Why close the PD incidents for test cases that haven't failed in the last 48 hours?

This is a heuristic designed to automatically close incidents when the underlying test problem has been dealt with. If we stop seeing failures for a testcase, it probably means that the testcase has stopped failing. This can backfire, and a more intelligent heuristic is certainly possible.
 
[OpenShift Offline Token]:https://cloud.redhat.com/openshift/token
[configs]:/configs/
[config package]:/pkg/common/config/config.go
[Makefile]:/Makefile
[Addon Testing Guide]:/docs/Addons.md
[Grafana dashboards]:https://grafana.datahub.redhat.com/dashboard/db/osd-health-metrics?orgId=1
[Writing Tests]:/docs/Writing-Tests.md
[Config variables]:/docs/Config.md
