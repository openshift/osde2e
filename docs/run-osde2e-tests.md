# How To: test osde2e on an existing cluster on staging
[osde2e](https://github.com/openshift/osde2e) runs automated tests on staging and production clusters, we will see a way of running them on-demand

> :warning: This is not meant to be open-sourced as some of the steps are only available to select members (SRE-P is one of them)

## Create a staging cluster / obtain the clusterid

if you don't have a stage cluster create it, if you do, take the clusterid with:

``` console
$ CLUSTER= # enter your cluster identification here
$ CLUSTERID=$(ocm describe cluster ${CLUSTER} --json | jq --raw-output .id)
$ echo ${CLUSTERID}
abcdefghijklmnopqrstuvwxyzabcdefgh
```

## Pull the kubeconfig

use the endpoint to pull the kubeconfig ( :warning: only works on stage :warning: )

we need the kubeconfig this was as osde2e requires kubeadmin, and `ocm cluster login -t $CLUSTERID` would need to be elevated to run the commands (which is not ideal)

``` console
$ KUBECONFIG_LOCATION= # enter a path
$ KUBECONFIG_LOCATION=$(mktemp) # in case you just want to have a random place that will be deleted at poweroff
$ CLUSTERID= # from prev step
$ export OCM_CLIENT_ID= # your OCM service account client ID
$ export OCM_CLIENT_SECRET= # your OCM service account client secret
$ ocm login --url stg --client-id ${OCM_CLIENT_ID} --client-secret ${OCM_CLIENT_SECRET}
$ ocm get /api/clusters_mgmt/v1/clusters/${CLUSTERID}/credentials | jq --raw-output .kubeconfig > ${KUBECONFIG_LOCATION}

```

## Set Testing environment


Test execution is controlled via environment variables, to see list of handled variables run `egrep "Env:\W+" pkg/common/config/config.go` at the root of the cloned e2e tests repo.

> :warning: By default e2e suite will hibernate your cluster after completing the tests

```
$ export POLLING_TIMEOUT=30             # wait for an object to be created before failing the test.
$ export MUST_GATHER=false              # don't run a Must-Gather process upon completion of the tests.
$ export HIBERNATE_AFTER_USE=false      # don't hibernate the cluster
```

## Build and run the osde2e tests

As specified in the osde2e's README, build osde2e using `make build`, which will save a binary in `out/osde2e` that will contain and execute all defined tests.

Run tests:

``` console
$ ./out/osde2e test \
   --cluster-id ${CLUSTERID} \
   --environment stage \
   --configs informing-suite,stage \
   --kube-config=${KUBECONFIG_LOCATION} \
```

- cluster-id: staging cluster id
- environment: set test execution environment
- configs
  - informing-suite: runs the tests that are not in the e2e, only the stage
  - stage: the staging ocm url, so we see the right cluster
- kube-config: the kubeconfig we pulled in a previous step

# How to: Run selected tests

Run tests:

``` console
$ ./out/osde2e test \
   --cluster-id ${CLUSTERID} \
   --environment stage \
   --configs stage \
   --kube-config=${KUBECONFIG_LOCATION} \
   --focus-tests='.*Exporter'
```

- cluster-id: staging cluster id
- environment: set test execution environment
- configs
  - stage: the staging ocm url, so we see the right cluster
- kube-config: the kubeconfig we pulled in a previous step
- focus-tests:
    This is on making running only running the tests I want, it's using ginkgo's `--focus` feature


# How To: test cluster upgrades with osde2e

## On an existing cluster, using OCM and managed-upgrade-operator

Follow the same steps in the preceding section about creating the cluster and setting your OCM credentials.

* Set the OCM service account credentials environment variables.

```bash
export OCM_CLIENT_ID=your-client-id
export OCM_CLIENT_SECRET=your-client-secret
```

Performing an OCM-driven upgrade requires that you _not_ use the Kubeconfig-driven method of access. Set the `CLUSTER_ID` environment variable to be your cluster's internal CLUSTER ID.

```bash
export CLUSTER_ID=xxxxxxxxx
```

* Set the `UPGRADE_MANAGED` environment variable to be `true` to indicate OSDE2E should upgrade using the `managed-upgrade-operator`.

```bash
export UPGRADE_MANAGED=true
```

Now you must decide on which version of OpenShift you wish OSDE2E to upgrade you to.

* First, decide which upgrade channel group your cluster needs to be on and ensure your cluster is in that group.

You can determine your current channel via OCM:

```bash
$ ocm get cluster $CLUSTER_ID | jq -r ".version.channel_group"
stable
```

You can change the upgrade channel similarly via OCM. You can choose `stable`, `fast` or `candidate`:

```bash
$ echo '{"version":{"channel_group":"candidate"}}' | ocm patch cluster $CLUSTER_ID
```

* If you wish to upgrade to a specific version, first [check that it is a valid version to upgrade to](https://access.redhat.com/labs/ocpupgradegraph/update_path). Then set the `UPGRADE_RELEASE_NAME` environment variable to be the cluster version which is your target *To* version (Note: at present, a `openshift-v` prefix is required to be added)

```bash
export UPGRADE_RELEASE_NAME=openshift-v4.7.17
```

* Otherwise, you can set either of the `UPGRADE_TO_LATEST`, `UPGRADE_TO_LATEST_Z` or `UPGRADE_TO_LATEST_Y` environment variables to `true` to upgrade to the latest version / latest Z-Stream version / latest Y-Stream version possible, given your cluster's current version and channel group.

```bash
export UPGRADE_TO_LATEST_Z=true
```

* Launch OSDE2E to install a cluster and upgrade it to the version specified:

```bash
$ ocm login --url stg --client-id ${OCM_CLIENT_ID} --client-secret ${OCM_CLIENT_SECRET}
$ ./out/osde2e test \
   -i ${CLUSTER_ID} \
   --configs e2e-suite,stage \
   --focus-tests='Only.*|run.*|these.*|tests.*'
```

This will:

- verify the cluster is healthy
- run the tests specified
- upgrade the cluster
- run a health check to verify the cluster is healthy post-upgrade

## On a new cluster

Follow the same steps as above, but remove the `"-i ${CLUSTER_ID}"` portion of the `osde2e` call:

```bash
$ ocm login --url stg --client-id ${OCM_CLIENT_ID} --client-secret ${OCM_CLIENT_SECRET}
$ ./out/osde2e test \
   --configs e2e-suite,stage \
   --focus-tests='Only.*|run.*|these.*|tests.*'
```

This will:

- create a new cluster
- run the tests specified
- upgrade the cluster
- run a health check to verify the cluster is healthy post-upgrade
