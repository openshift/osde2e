## Description

- Operators in OSD currently use the older operator-sdk versions (v0.15.x, v0.17.x, v0.18.x).
- With the release of operator-sdk v1.x, operator-sdk now has a new file directory structure which is closely aligned to [Kubebuilder](https://book.kubebuilder.io/).
- [Boilerplate](https://github.com/openshift/boilerplate) with v2.0 build image has stopped enforcing operator-sdk versions on repositories and solely relies on controller-gen for generating manifests.
- This serves as an opportunity to upgrade the existing OSD operators to the latest operator-sdk version so they can leverage the latest API versions as part of the upgraded SDK.
- This document is a guide to migrate/upgrade an OSD operator to the latest version.

## Prerequisites

- An existing operator repository, working with operator-sdk in some version.
- A local fork of the operator repository.
- Latest version of operator-sdk installed (v1.19.1 at the time of writing this). Install it from here: https://sdk.operatorframework.io/docs/installation.

## Directory Structure

Drawing out a directory structure comparison between v0.x & v1.x operator-sdk versions and outlining the changes.

| **v0.x**                                                     | **v1.x**                                                            | **Purpose**                                                                                                        |
| ------------------------------------------------------------ | ------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| cmd/manager/main.go                                          | main.go                                                             | This instantiates a new manager which registers all custom resource definitions.                                   |
| pkg/apis                                                     | /api                                                                | Contains the directory tree that defines the APIs of the Custom Resource Definitions(CRD).                         |
| pkg/controller                                               | /controllers                                                        | This pkg contains the controller implementations.                                                                  |
| build/Dockerfile                                             | Dockerfile                                                          | Dockerfile and build scripts used to build the operator.                                                           |
| deploy<br>deploy/crds<br>deploy/operator.yaml<br>deploy/rbac | config<br>config/crds<br>config/manager/manager.yaml<br>config/rbac | Contains various YAML manifests for registering CRDs, setting up RBAC, and deploying the operator as a deployment. |

## Migration

- Initialize a new project with the same domain.

```bash
$ mkdir <OPERATOR_NAME>
$ cd <OPERATOR_NAME>
$ operator-sdk init --domain <DOMAIN> --repo <UPSTREAM_OPERATOR_REPOSITORY>
```

**Note:-**

1. We will refer to the forked operator repo as _Fork_ going forward. Copy the Fork's `.git` to the new project folder.

2. The domain can be determined by looking at the `spec.group` field in one of the CRDs in the Fork’s deploy/crds directory.

- Create new APIs with resource and controllers.

```bash
$ operator-sdk create api \
    --group=<group> \
    --version=<version> \
    --kind=<kind> \
    --resource \
    --controller
```

- Migrate APIs. Copy and edit the API definition from Fork `pkg/apis/../<kind>_types.go` to `api/../<kind>_types.go`.

**Note:-**

1. This involves copying the code from the `Spec` and `Status` fields.

2. The `+k8s:deepcopy-gen:interfaces=...` marker was replaced with `+kubebuilder:object:root=true`

- Migrate the controller code from Fork `pkg/controller/<kind>/<kind>_controller.go` to `controllers/<kind>_controller.go`. This involves copying and moving around the code from Fork's controllers.

  - Copy over any struct fields from the existing project into the new `<Kind>Reconciler` struct.

    **Note:** The Reconciler struct has been renamed from `Reconcile<Kind>` to `<Kind>Reconciler`.

  - Replace the `// your logic here` in the new layout with your reconcile logic.
  - Copy the code under `func add(mgr manager.Manager, r reconcile.Reconciler)` to `func SetupWithManager`:

- Migrate tests. Edit `go:generate mocken -destination=..` to point to the relevant directories and run `make go-generate` to generate mocks. Copy over tests from Fork `pkg/api/controller/<kind>_test.go` to `controllers/<kind>_test`.

- Edit `main.go` as per Fork's `cmd/manager/main.go`.
  **Note:-** The SDK's `leader.Become` was replaced by the controller-runtime’s `leader` with lease mechanism.

## Making it Boilerplate compatible

- Move `Dockerfile` to `build/Dockerfile`. If not boilerplate bootstrap would fail.

```bash
    $ mkdir build && mv Dockerfile build
```

- Make sure to commit all your changes till here as the next steps would require a clean checkout of the repository.

- Bootstrap boilerplate and commit the changes.

```bash
curl --output boilerplate/update --create-dirs https://raw.githubusercontent.com/openshift/boilerplate/master/boilerplate/update
chmod +x boilerplate/update
echo "openshift/golang-osd-operator" > boilerplate/update.cfg
printf "\n.PHONY: boilerplate-update\nboilerplate-update:\n\t@boilerplate/update\n" >> Makefile
make boilerplate-update
sed -i '1s,^,include boilerplate/generated-includes.mk\n\n,' Makefile
```

- Export your Fork's directory as `OLD_SDK_REPO_DIR`.

```bash
  $ export OLD_SDK_REPO_DIR=<Forked repository directory>
```

- Run `make migrate-to-osdk1`. This will:

  - Create a directory named _deploy_ in the root of the current project and copy all files/folders from _${OLD_SDK_REPO_DIR}/deploy_ to _/deploy_ except the _crds_ folder.
  - Copy _${OLD_SDK_REPO_DIR}/Makefile_ and _${OLD_SDK_REPO_DIR}/.gitignore_ to the current project.
  - Copy _${OLD_SDK_REPO_DIR}/pkg_ to _./pkg_. This will copy everything except _pkg/apis_ and _pkg/controlles_ folders.
  - Copy rest of files and folders from Fork's root to the current project's root except the existing ones.
    <br>**Note:** _${OLD_SDK_REPO_DIR}/cmd_, _{OLD_SDK_REPO_DIR}/version_ and existing files/folders are excluded.

- The latest Operator SDK requires `controller-gen v0.8.0` and `openapi-gen v0.23.0`. To ensure better portability we will use boilerplate's `container-make`.

- Run `./boilerplate/_lib/container-make generate`. This will:
  - Generate CRDs under _deploy/crds_ via controller-gen-v0.8.0.
  - Generate openAPI definition file through openapi-gen v0.23.0 to be used in open API spec generation on API servers under the _api_ directory.
    <br>**Note**: openapi-gen (which relies on GOPATH) produces absolute paths.
  - Run `go generate`.

## Post Migration Checks

- You can try testing the operator locally and make sure it is functioning as expected.

- Make sure all the unit tests are passing - `./boilerplate/_lib/container-make test`

- Coverage succeeds - `./boilerplate/_lib/container-make coverage`

- There are no linting errors - `./boilerplate/_lib/container-make lint`

- Validation passes - `./boilerplate/_lib/container-make validate`

- Test App SRE scripts locally - [app-sre.md](https://github.com/openshift/boilerplate/blob/master/boilerplate/openshift/golang-osd-operator/app-sre.md)

## References

- https://sdk.operatorframework.io/docs/building-operators/golang/migration
- https://sdk.operatorframework.io/docs/upgrading-sdk-version
