# OSDe2e
Comprehensive testing solution for Service Delivery

## Purpose
Provide a standard for testing every aspect of the Openshift Dedicated product. Use data derived from tests to inform release and product decisions.

## Execution 
These steps run the OSDe2e test suite. All commands should be run from the root of this repo.

A properly setup Go workspace using **Go 1.13+ is required**.

Get token to launch OSD clusters here.


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

Run tests:
```
make test
```


## Test Examples
Below are some common examples when running OSDe2e. There is a large number of config options that may change or help, so please view the config package for more information.

Using ENV Config
Testing against an existing cluster (with your kubeconfig pointing at it)
```
TEST_KUBECONFIG=~/.kube/config \
make test
```
*Note: You must skip certain Operator tests that only exist in a hosted OSD instance:*
```
 -ginkgo.skip="Managed Velero Operator|Dedicated Admin Operator|Configure AlertManager Operator"
```

Spinning up a hosted-OSD instance and testing against it
```
OCM_TOKEN=$(cat ~/.ocm-token) \
OSD_ENV=prod \
CLUSTER_NAME=my-name-osd-test \
MAJOR_TARGET=4 \
MINOR_TARGET=2 \
make test
``` 

Using YAML Config
```
E2ECONFIG=./osde2e.yaml \
make test
```

Dry Run example
```
dryRun: true
kubeconfig:
 path: /path/to/your/.kube/config
```

Full Test example
```
dryRun: false
cluster:
 name: jsica-test
upgrades:
 majorTarget: 4
 minorTarget: 2
ocm:
 debug: false
 token: [Redacted]
 env: stage
tests:
 ginkgoSkip: OpenShift E2E|Cluster state|Managed Velero Operator|Dedicated Admin Operator|Configure AlertManager Operator
```

*This is a recent change and is not widely used yet. Please refer to pkg/config for more info on YAML config options.*

Config options are currently parsed by loading defaults, attempting to load environment variables, and finally attempting to load config data from a YAML file. There are instances where you may want to have most of your config in a YAML file while keeping one or two sensitive config options as environment variables (OCM Token)

## Different Test Types
Core tests and Operator tests reside within the OSDe2e repo and are maintained by the CICD team. The tests are written and compiled as part of the OSDe2e project. 
* Core Tests
* OpenShift Conformance
* OC Must Gather
* Verify 
  * All pods are healthy or successful
  * ImageStreams exist
  * Project creation possible
  * Ingress to console possible
* Operator tests
  * ConfigureAlertManager
  * DedicatedAdmin
  * ManagedVelero

Third-party (Addon) tests are built as containers that spin up and report back results to OSDe2e. These containers are built and maintained by external groups looking to get CI signal for their product within OSD. The definition of a third-party test is maintained within the `managed-tenants` repo and is returned via the Add-Ons API.

Third-party testing, unlike the previous test types, occurs within the `ci-int` Jenkins instance so non-SD groups have a single pane of glass for managing things within Service Delivery.
 * Third Party Tests
   * ClusterLogging

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
2. Provision Cluster
3. Verify Cluster Integrity
4. Run Tests (pre-upgrade)
5. Capture / Upload logs to GCS
6. Upgrade Cluster
7. Verify Cluster Integrity
8. Run Tests (post-upgrade)
9. Capture / Upload logs to GCS

With a dry-run, OSDe2e only performs the “Load Config” step and outputs the parameters the run would have used. With a vanilla-install run, the run is complete after the first “Capture/Upload” step.

A failure at any step taints and fails the run. 

## Reporting / Alerting
Every run of OSDe2e captures as much data as possible. This includes cluster and pod logs, prometheus metrics, and test info. In addition to cluster-specific info, the version of hive and OSDe2e itself is captured to identify potential flakes or environment failures. Every test suite generates a `junit.xml` file that contains test names, pass/fails, and the time the test segment took. It is expected that addon testing will follow this pattern and generate their own `junit.xml` file for their test results. 

The `junit.xml` files are converted to meaningful metrics and stored in DataHub. These metrics are then published via Grafana dashboards used by Service Delivery as well as Third Parties to monitor project health and promote confidence in releases. Alerting rules are housed within the DataHub Grafana instance and addon authors can maintain their own individual dashboards.

## Writing tests
Documentation on writing tests can be found [here](./docs/Writing-Tests.md).
