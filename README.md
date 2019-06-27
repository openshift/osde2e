# osde2e

This project verifies OSD releases by:
- Starting an OSD cluster
- Verifying operation through testing
- Reporting test results to TestGrid

## Build
These steps will build the osde2e test suite.

### Local
A properly setup [Go workspace](https://golang.org/doc/code.html#GOPATH), Make, and [Glide](https://github.com/Masterminds/glide#install) are required to build osde2e.

```bash
glide install --strip-vendor
make out/osde2e
```

### Docker
The built image contains the test binary which is executed by default on run.

```
make build-image
```

## Test
Configuration must be defined before running osde2e. The complete set of options can be found in [`pkg/config`](./pkg/config/config.go).

The following environment variables are required:
- `UHC_TOKEN`: The token used to authenticate with the OSD environment. This token can be retrieved [here](https://cloud.redhat.com/openshift/token).

The following environment variables are required for TestGrid reporting:
- `TESTGRID_BUCKET`: The Google Storage bucket storing TestGrid builds.
- `TESTGRID_PREFIX`: The prefix for builds in the bucket.
- `TESTGRID_SERVICE_ACCOUNT`: The Base64 encoded JSON Service Account used to access the bucket.

The following environment variables disable certain functionality:
- `NO_DESTROY`: Cluster is not destroyed after testing.
- `NO_TESTGRID`: Results are not uploaded to TestGrid after build

The following environment variable allows an existing cluster to be test:
- `CLUSTER_ID`: The ID of the cluster provided by OSD.

### Local
```
make test
```

### Docker
```
make docker-test 
```

## Writing tests
Documentation on writing tests can be found [here](./docs/Writing-Tests.md).
