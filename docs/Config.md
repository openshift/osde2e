# **Environment variable configuration**

Some environment variables commonly used for pipelines under osde2e are indicated below. These flags or variables are used by osde2e.


## Common Environment Variables

### Cluster related:

| Environment variable                     | Usage                                                                                                                                            |
| ---------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| TEST_KUBECONFIG                          | This variable represents the filepath of an existing Kubeconfig. Will override fetching credentials from the OCM provider if specified.          |
| OSD_ENV                                  | int, stage, prod                                                                                                                                 |
| CLOUD_PROVIDER_ID                        | aws, gcp                                                                                                                                         |
| CLOUD_PROVIDER_REGION                    | Must be a valid region enabled within OCM. Ex. us-east-1                                                                                         |
| CLUSTER_VERSION                          | Must be a valid clusterimageset enabled within that environment. Ex. openshift-v4.6.0-fc.3-fast                                                  |
| CLUSTER_EXPIRY_IN_MINUTES                | Optional: Set this if you want to ensure the cluster gets cleaned up.                                                                            |
| CLUSTER_ID                               | The value identifying the cluster. If set at start, an existing cluster is tested.                                                               |
| CLUSTER_NAME                             | The name of the cluster to be created.                                                                                                           |
| PROVIDER                                 | Provider is what provider to use to create/delete clusters. Ex. ocm, rosa, mock                                                                  |
| PROVISION_SHARD_ID                       | ProvisionShardID is the shard ID that is set to provision a shard for the cluster.                                                               |
| MULTI_AZ                                 | MultiAZ deploys a cluster across multiple availability zones.                                                                                    |
| DESTROY_CLUSTER                          | Set to true if you want to the cluster to be explicitly deleted after the test.                                                                  |
| AFTER_TEST_CLUSTER_WAIT                  | AfterTestWait is how long to keep a cluster around after tests have run.                                                                         |
| CLUSTER_UP_TIMEOUT                       | InstallTimeout is how long to wait before failing a cluster launch.                                                                              |
| USE_LATEST_VERSION_FOR_INSTALL           | UseLatestVersionForInstall will select the latest cluster image set available for a fresh install.                                               |
| USE_MIDDLE_CLUSTER_IMAGE_SET_FOR_INSTALL | UseMiddleClusterImageSetForInstall will select the cluster image set that is in the middle of the list of ordered cluster versions known to OCM. |
| USE_OLDEST_CLUSTER_IMAGE_SET_FOR_INSTALL | UseOldestClusterImageSetForInstall will select the cluster image set that is in the end of the list of ordered cluster versions known to OCM.    |
| DELTA_RELEASE_FROM_DEFAULT               | DeltaReleaseFromDefault will select the cluster image set that is the given number of releases from the current default in either direction.     |
| NEXT_RELEASE_AFTER_PROD_DEFAULT          | NextReleaseAfterProdDefault will select the cluster image set that the given number of releases away from the the production default.            |
| CLEAN_CHECK_RUNS                         | CleanCheckRuns lets us set the number of osd-verify checks we want to run before deeming a cluster "healthy"                                     |
| INSPECT_NAMESPACES                       | InspectNamespaces is a comma-delimeted list of namespaces to perform an `oc adm inspect` on during E2E cleanup                                   |
| USE_PROXY_FOR_INSTALL                    | UseProxyForInstall will use a cluster-wide proxy for the cluster installation, provided that cluster proxy configuration is also supplied.       |

### ROSA cluster related:-
 
| Environment variable | Usage                                                        |
| -------------------- | ------------------------------------------------------------ |
| ROSA_ENV             | Environment for the e2e testing, default to prod.            |
| ROSA_STS             | Boolean value to indicate the cluster is STS enabled or not. |
| ROSA_REPLICAS        | Compute node count for the rosa cluster, default is 2.       |

### Hypershift cluster related:-
 
| Environment variable | Usage                                                                       |
| -------------------- | --------------------------------------------------------------------------- |
| Hypershift           | Boolean value to indicate the cluster should be created as a HostedCluster. |
 
### OCM related:-
 
| Environment variable           | Usage                                                                                                                                 |
| ------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------- |
| DEBUG_OSD                      | Debug shows debug level messages when enabled.                                                                                        |
| NUM_RETRIES                    | NumRetries is the number of times to retry each OCM call.                                                                             |
| OCM_COMPUTE_MACHINE_TYPE       | ComputeMachineType is the specific cloud machine type to use for compute nodes.                                                       |
| OCM_COMPUTE_MACHINE_TYPE_REGEX | ComputeMachineTypeRegex is the regex for cloud machine type to use for compute nodes.                                                 |
| OCM_USER_OVERRIDE              | UserOverride will hard set the user assigned to the "owner" tag by the OCM provider.                                                  |
| OCM_FLAVOUR                    | Flavour is an OCM cluster descriptor for cluster defaults                                                                             |
| OCM_ADDITIONAL_LABELS          | AdditionalLabels is used to add more specific labels to a cluster in OCM.                                                             |
| OCM_CCS                        | CCS defines whether the cluster should expect cloud credentials or not                                                                |
| OCM_CCS_ADMIN                  | Overwrite Flag that will attempt to cycle osdCcsAdmin credentials for a CCS install when the osdCcsAdmin credentials were not passed. |
| TEST_KUBECONFIG                | Path to a local kubeconfig; will override fetching Kubeconfig credentials from OCM if specified.                                      |
  
### Upgrade variables:-

| Environment variable            | Usage                                                                                                            |
| ------------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| UPGRADE_TYPE                    | UpgradeType will define what managed cluster upgrader to use. Valid values "OSD" (default) or "ARO".             |
| UPGRADE_TO_LATEST               | UpgradeToLatest will upgrade to the latest valid version found.                                                  |
| UPGRADE_TO_LATEST_Z             | UpgradeToLatestZ looks for the newest valid patch-release and selects it.                                        |
| UPGRADE_TO_LATEST_Y             | UpgradeToLatestY looks for the newest valid minor release upgrade path and selects it.                           |
| UPGRADE_RELEASE_NAME            | ReleaseName is the name of the release in a release stream.                                                      |
| UPGRADE_IMAGE                   | Image is the release image a cluster is upgraded to. If set, it overrides the release stream and upgrades.       |
| UPGRADE_MONITOR_ROUTES          | MonitorRoutesDuringUpgrade will monitor the availability of routes whilst an upgrade takes place.                |
| UPGRADE_MANAGED_TEST_PDBS       | Create disruptive Pod Disruption Budget workloads to test the Managed Upgrade Operator's ability to handle them. |
| UPGRADE_MANAGED_TEST_RESCHEDULE | Test the managed upgrade when the upgrade schedule changed.                                                      |

 
 
### Job related:-

| Environment variable | Usage                                                                                                                                                                                                                                                     |
| -------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| JOB_NAME             | JobName lets you name the current e2e job run.                                                                                                                                                                                                            |
| BUILD_NUMBER         | JobID is the ID designated by prow for a specific run.                                                                                                                                                                                                    |
| DRY_RUN              | DryRun is a boolean flag that lets you run osde2e all the way up to the e2e tests then skips them.                                                                                                                                                        |
| REPORT_DIR           | ReportDir is the location JUnit XML results are written.                                                                                                                                                                                                  |
| ARTIFACTS            | An alias for ReportDir but runs on Prow.                                                                                                                                                                                                                  |
| SUFFIX               | Suffix is used at the end of test names to identify them.                                                                                                                                                                                                 |
| MUST_GATHER          | MustGather is a boolean flag that will run a Must-Gather process upon completion of the tests.                                                                                                                                                            |
| BASE_JOB_URL         | BaseJobURL is the root location for all job artifacts. For example, https://storage.googleapis.com/origin-ci-test/logs/osde2e-prod-gcp-e2e-next/61/build-log.txt would be https://storage.googleapis.com/origin-ci-test/logs -- This is also our default. |
| BASE_PROW_URL        | BaseProwURL is the root location of Prow.                                                                                                                                                                                                                 |
 
 
### General test related:-

| Environment variable        | Usage                                                                                                           |
| --------------------------- | --------------------------------------------------------------------------------------------------------------- |
| POLLING_TIMEOUT             | PollingTimeout is how long (in seconds) to wait for an object to be created before failing the test.            |
| GINKGO_SKIP                 | GinkgoSkip is a regex passed to Ginkgo that skips any test suites matching the regex. ex. "Operator"            |
| GINKGO_FOCUS                | GinkgoFocus is a regex passed to Ginkgo that focus on any test suites matching the regex. ex. "Operator"        |
| GINKGO_LOG_LEVEL            | GinkgoLogLevel allows controlling the Ginkgo reporter output                                                    |
| TESTS_TO_RUN                | TestsToRun is a list of files which should be executed as part of a test suite                                  |
| SUPPRESS_SKIP_NOTIFICATIONS | SuppressSkipNotifications suppresses the notifications of skipped tests                                         |
| CLEAN_RUNS                  | CleanRuns is the number of times the test-version is run before skipping.                                       |
| OPERATOR_SKIP               | OperatorSkip is a comma-delimited list of operator names to ignore health checks from. ex. "insights,telemetry" |
| SKIP_CLUSTER_HEALTH_CHECKS  | SkipClusterHealthChecks skips the cluster health checks. Useful when developing against a running cluster.      |
| METRICS_BUCKET              | MetricsBucket is the bucket that metrics data will be uploaded to.                                              |
| SERVICE_ACCOUNT             | ServiceAccount defines what user the tests should run as. By default, osde2e uses system:admin                  |
 

### Addon test related:-

| Environment variable    | Usage                                                                                                                                                                                                                                                                     |
| ----------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ADDON_IDS_AT_CREATION   | Comma separated list of IDs to create at cluster creation time                                                                                                                                                                                                            |
| ADDONS_IDS              | Comma separated list of IDs to install after a cluster is created.                                                                                                                                                                                                        |
| ADDON_TEST_HARNESSES:   | TestHarnesses is a comma separated list of container images that will test the addon                                                                                                                                                                                      |
| ADDON_TEST_USER         | TestUser is the OpenShift user that the tests will run as. If "%s" is detected in the TestUser string, it will evaluate that as the project namespace. Ex. "system:serviceaccount:%s:dedicated-admin" . Evaluated : "system:serviceaccount:osde2e-abc123:dedicated-admin" |
| ADDON_RUN_CLEANUP       | RunCleanup is a boolean to specify whether the testHarnesses should have a separate cleanup phase. This phase would run at the end of all e2e testing                                                                                                                     |
| ADDON_CLEANUP_HARNESSES | CleanupHarnesses is a comma separated list of container images that will clean up any artifacts created after test harnesses have run                                                                                                                                     |
| ADDON_POLLING_TIMEOUT   | PollingTimeout defines in seconds the amount of time to wait for an add-on test job to finish before timing it out                                                                                                                                                        |
 
### Prometheus related:-

| Environment variable    | Usage                                             |
| ----------------------- | ------------------------------------------------- |
| PROMETHEUS_ADDRESS      | Address of the Prometheus instance to connect to. |
| PROMETHEUS_BEARER_TOKEN | Token needed for communicating with Prometheus.   |
 
### Proxy related:-

| Environment variable | Usage                                                                                                              |
| -------------------- | ------------------------------------------------------------------------------------------------------------------ |
| TEST_HTTP_PROXY      | Address of the HTTP Proxy to be added to a cluster.                                                                |
| TEST_HTTPS_PROXY     | Address of the HTTPS Proxy to be added to a cluster.                                                               |
| USER_CA_BUNDLE       | A file contains a PEM-encoded X.509 certificate bundle that will be added to the nodes' trusted certificate store. |


## Command Line Flags for osde2e

CLI flags that are commonly used include:
 
### For the test sub-command:
```
--cluster-id: Existing OCM cluster ID to run tests against.
--configs:  A comma separated list of built in configs to use. (stage, prod, e2e-suite, etc.)
--custom-config: Custom config file for osde2e.
--destroy-cluster: A flag to trigger cluster deletion after test completion.
--environment: Cluster provider environment to use. (ocm, rosa, etc.).
--kube-config: Path to local Kube config for running tests against.
--skip-health-check:  a flag to skip cluster health checks.
--skip-tests: Skip any Ginkgo tests whose names match the regular expression.
``` 

### For the query sub-command:
```
--output-format:  Output format for query results (json|prom). Defaults to json. (default "-")
```
 
## Common config flag values

The following are the values that can be plugged in for the --configs flag when running osde2e. The values correspond to existing YAML files in the /configs folder:-

### OSD environment values:


| Config Value | Usage                                                                                             |
| ------------ | ------------------------------------------------------------------------------------------------- |
| int          | To run osde2e in the integration environment.                                                     |
| stage        | To run osde2e on stage.                                                                           |
| prod         | To run osde2e in the production environment. (This is the default value if nothing is specified.) |
| scale        | To set scale testing configurations for a cluster.                                                |


### Cloud Provider values:

| Config Value | Usage                                 |
| ------------ | ------------------------------------- |
| aws          | To specify aws as the cloud provider. |
| gcp          | To specify gcp as the cloud provider. |


### AWS specific values:
| Environment variable  | Usage                                                           |
| --------------------- | --------------------------------------------------------------- |
| AWS_ACCOUNT           | AWS account to use for testing.                                 |
| AWS_ACCESS_KEY        | AWSAccessKeyID for provisioning clusters.                       |
| AWS_SECRET_ACCESS_KEY | AWSSecretAccessKey for provisioning clusters.                   |
| AWS_REGION            | AWSRegion for provisioning clusters.                            |
| AWS_VPC_SUBNET_IDS    | AWSVPCSubnetIDs for provisioning clusters for BYO-VPC clusters. |

### Cluster Provider values:

| Config Value | Usage                                    |
| ------------ | ---------------------------------------- |
| ocm          | To specify ocm as the cluster provider.  |
| rosa         | To specify rosa as the cluster provider. |


### Test suite values:

| Config Value               | Usage                                                                                                  |
| -------------------------- | ------------------------------------------------------------------------------------------------------ |
| e2e-suite                  | To test osde2e using the e2e tets suite (Includes operates, service-definition and app-builds suites). |
| informing-suite            | To run informing tests on osde2e clusters.                                                             |
| openshift-suite            | To run the openshift test suite on osde2e clusters.                                                    |
| addon-suite                | To include addon testing for osde2e clusters.                                                          |
| conformance-suite          | To run conformance tests on osde2e clusters.                                                           |
| scale-mastervertical-suite | To run the scale:master vertical suite on osde2e clusters.                                             |
| scale-nodes-and-pods-suite | To run the scale:nodes and pods suite on osde2e clusters.                                              |
| scale-performance-suite    | To run the scale:performance suite on osde2e clusters.                                                 |


### Other test flags:

| Config Value       | Usage                                                                              |
| ------------------ | ---------------------------------------------------------------------------------- |
| dry-run            | To run osde2e all the way up to the e2e tests then skips them.                     |
| skip-health-checks | To run osde2e while skipping all the preliminary cluster health checks.            |
| long-timeout       | To extend cluster expiry to about 8 hours.                                         |
| log-metrics        | To set log metric configs for a couple of errors that pop up while running osde2e. |
| region-random      | To set a random region for a cluster being created by the cloud provider.          |



### Install/Upgrade flags:

| Config Value                     | Usage                                                                                                                    |
| -------------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| use-middle-version               | To use the middle version for cluster install.                                                                           |
| use-oldest-version               | To use the oldest version for cluster install.                                                                           |
| one-release-from-prod-default    | To select the cluster image set that the given number of releases away from the the production default (1 in this case). |
| two-releases-from-prod-default   | To select the cluster image set that the given number of releases away from the the production default (2 in this case). |
| nightly-release-for-prod-default | To select the cluster image set that the given number of releases away from the the production default (0 in this case). |
| upgrade-to-latest                | To select the newest valid version to upgrade to                                                                         |
| upgrade-to-latest-z              | To select the newest valid patch version to upgrade to                                                                   |
| upgrade-to-next-y                | To select the newest valid minor release to upgrade to                                                                   |
| upgrade-rescheduled              | To test the upgrade being rescheduled.                                                                                   |
