# osde2e Options

## Table of Contents
- [required](#required)
- [tests](#tests)
- [environment](#environment)
- [cluster](#cluster)
- [version](#version)
- [upgrade](#upgrade)
- [testgrid](#testgrid)
- [other](#other)



## required
These options are required to run osde2e.

### `UHC_TOKEN`

- UHCToken is used to authenticate with UHC.

- Type: `string`

## tests


### `CLEAN_RUNS`

- CleanRuns is the number of times the test-version is run before skipping.

- Type: `int`

### `REPORT_DIR`

- ReportDir is the location JUnit XML results are written.

- Type: `string`

### `SUFFIX`

- Suffix is used at the end of test names to identify them.

- Type: `string`

## environment


### `DEBUG_OSD`

- DebugOSD shows debug level messages when enabled.

- Type: `bool`

### `OSD_ENV`

- OSDEnv is the OpenShift Dedicated environment used to provision clusters.

- Type: `string`

## cluster


### `CLUSTER_ID`

- ClusterID identifies the cluster. If set at start, an existing cluster is tested.

- Type: `string`

### `CLUSTER_NAME`

- ClusterName is the name of the cluster being created.

- Type: `string`

### `MULTI_AZ`

- MultiAZ deploys a cluster across multiple availability zones.

- Type: `bool`

### `NO_DESTROY`

- NoDestroy leaves the cluster running after testing.

- Type: `bool`

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

### `TARGET_STREAM`

- TargetStream lets you select a specific release stream from Cincinnati or the Release Controller to install.

- Type: `string`

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

## testgrid
These options configure reporting test results to TestGrid.

### `NO_TESTGRID`

- NoTestGrid disables reporting to TestGrid.

- Type: `bool`

### `TESTGRID_BUCKET`

- TestGridBucket is the Google Cloud Storage bucket where results are reported for TestGrid.

- Type: `string`

### `TESTGRID_PREFIX`

- TestGridPrefix is used to namespace reports.

- Type: `string`

### `TESTGRID_SERVICE_ACCOUNT`

- TestGridServiceAccount is a Base64 encoded Google Cloud Service Account used to access the TestGridBucket.

- Type: `[]byte`

## other
Various additional options for configuring osde2e.

### `USE_PROD`

- UseProd sends requests to production OSD.

Deprecated: Use OSD_ENV=prod instead.

- Type: `bool`