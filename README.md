# osde2e

This project checks OSD releases by starting an OSD cluster and verifying operation through testing

## Running
These steps run the osde2e test suite. All commands should be run from the root of this repo.

A properly setup [Go workspace](https://golang.org/doc/code.html#GOPATH) using Go 1.13+ is required.

1. Get token to launch OSD clusters [here](https://cloud.redhat.com/openshift/token).

2. Install dependencies:
    ```bash
    # Install dependencies
    $ go mod tidy

    # Copy them to a vendor dir
    $ go mod vendor
    ```
1. Set `OCM_TOKEN` environment variable:
    ```bash
    $ export OCM_TOKEN=<token from step 1>
    ```
4. Run tests:
    ```bash
    $ go test -v . -test.timeout 4h
    ```

## Configuring
osde2e is configured using a set of environment variables.
The options available are found [here](./docs/Options.md).

Common ones are:
- [`NO_DESTROY`](./docs/Options.md#no_destroy): don't delete clusters after testing
- [`CLUSTER_ID`](./docs/Options.md#cluster_id): test an existing cluster specified by ID

## Writing tests
Documentation on writing tests can be found [here](./docs/Writing-Tests.md).
