# osde2e Options

## Table of Contents
- [required](#required)
- [tests](#tests)
- [environment](#environment)
- [cluster](#cluster)
- [version](#version)
- [upgrade](#upgrade)



## required
These options are required to run osde2e.

### `OCM_TOKEN`

- OCMToken is used to authenticate with OCM.

- Type: `string`

## tests


### `CLEAN_RUNS`

- CleanRuns is the number of times the test-version is run before skipping.

- Type: `int`

### `DRY_RUN`

- DryRun lets you run osde2e all the way up to the e2e tests then skips them.

- Type: `bool`

### `GINKGO_FOCUS`

- GinkgoFocus is a regex passed to Ginkgo that focus on any test suites matching the regex. ex. "Operator"

- Type: `string`

### `GINKGO_SKIP`

- GinkgoSkip is a regex passed to Ginkgo that skips any test suites matching the regex. ex. "Operator"

- Type: `string`

### `OPERATOR_SKIP`

- OperatorSkip is a comma-delimited list of operator names to ignore health checks from. ex. "insights,telemetry"

- Type: `string`
- Default: `insights`

### `POLLING_TIMEOUT`

- PollingTimeout is how long (in mimutes) to wait for an object to be created
before failing the test.

- Type: `int64`
- Default: `30`

### `REPORT_DIR`

- ReportDir is the location JUnit XML results are written.

- Type: `string`

### `SUFFIX`

- Suffix is used at the end of test names to identify them.

- Type: `string`

## environment


### `AFTER_TEST_CLUSTER_WAIT`

- AfterTestClusterWait is how long to keep a cluster around after tests have run.

- Type: `int64`
- Default: `60`

### `CLUSTER_UP_TIMEOUT`

- ClusterUpTimeout is how long to wait before failing a cluster launch.

- Type: `int64`
- Default: `135`

### `DEBUG_OSD`

- DebugOSD shows debug level messages when enabled.

- Type: `bool`
- Default: `false`

### `OSD_ENV`

- OSDEnv is the OpenShift Dedicated environment used to provision clusters.

- Type: `string`
- Default: `prod`

## cluster


### `CLUSTER_EXPIRY_IN_MINUTES`

- ClusterExpiryInMinutes is how long before a cluster expires and is deleted by OSD.

- Type: `int64`
- Default: `210`

### `CLUSTER_ID`

- ClusterID identifies the cluster. If set at start, an existing cluster is tested.

- Type: `string`

### `CLUSTER_NAME`

- ClusterName is the name of the cluster being created.

- Type: `string`

### `DESTROY_CLUSTER`

- DestroyClusterAfterTest set to false if you want OCM to clean up the cluster itself after the test completes.

- Type: `bool`
- Default: `true`

### `MULTI_AZ`

- MultiAZ deploys a cluster across multiple availability zones.

- Type: `bool`
- Default: `false`

### `TEST_KUBECONFIG`

- Kubeconfig is used to access a cluster.

- Type: `[]byte`

## version


### `CLUSTER_VERSION`

- ClusterVersion is the version of the cluster being deployed.

- Type: `string`

### `MAJOR_TARGET`

- MajorTarget is the major version to target. If specified, it is used in version selection.

- Type: `int64`

### `MINOR_TARGET`

- MinorTarget is the minor version to target. If specified, it is used in version selection.

- Type: `int64`

## upgrade


### `UPGRADE_IMAGE`

- UpgradeImage is the release image a cluster is upgraded to. If set, it overrides the release stream and upgrades.

- Type: `string`

### `UPGRADE_RELEASE_NAME`

- UpgradeReleaseName is the name of the release in a release stream. UpgradeReleaseStream must be set.

- Type: `string`

### `UPGRADE_RELEASE_STREAM`

- UpgradeReleaseStream used to retrieve latest release images. If set, it will be used to perform an upgrade.

- Type: `string`
