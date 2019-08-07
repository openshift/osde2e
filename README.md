# osde2e

This project checks OSD releases by starting an OSD cluster, verifying operation through testing, and reporting test 
results to TestGrid.

## Running
These steps run the osde2e test suite. All commands should be run from the root of this repo.

A properly setup [Go workspace](https://golang.org/doc/code.html#GOPATH) and [Glide](https://github.com/Masterminds/glide#install) are required.

1. Get token to launch OSD clusters [here](https://cloud.redhat.com/openshift/token).

1. Install dependencies:
    ```bash
    glide install --strip-vendor
    ```
1. Set `UHC_TOKEN` environment variable:
    ```bash
    export UHC_TOKEN=<token from step 1>
    ```
1. Run tests:
    ```bash
    go test -v . -test.timeout 2h
    ```

## Configuring
osde2e is configured using a set of environment variables.
The options available are found [here](./docs/Options.md).

Common ones are:
- [`NO_DESTROY`](./docs/Options.md#no_destroy): don't delete clusters after testing
- [`CLUSTER_ID`](./docs/Options.md#cluster_id): test an existing cluster specified by ID

## Writing tests
Documentation on writing tests can be found [here](./docs/Writing-Tests.md).
