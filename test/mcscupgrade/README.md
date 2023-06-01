# HyperShift Management & Service Cluster Upgrade Test Suite

This folder contains the test suite to validate cluster upgrades for
Management (MC) & Service clusters (SC) can be upgraded successfully for
either Y/Z stream when hosted control plane workloads (HCP) are running
on them.

The test suite can be easily run by invoking `ginkgo run` directly,
from a `compiled binary` or from a `container image`. The container image
option makes this image consumable by other CI frameworks for easy running.

## Test Suite Model

Below outlines what the test suite is capable of:

* Before Suite
  * Input data validation
  * Identify MC/SC install and upgrade versions
  * Get MC/SC kubeconfigs
  * Deploy ROSA hosted control plane cluster targeting a specific provision shard id
* After Suite
  * Destroy the deployed ROSA hosted control plane cluster
* Test Cases:
  * Perform SC pre upgrade health checks
  * Perform SC upgrade
  * Perform SC post upgrade health checks
    * Checks prometheus alerts
    * Runs osd-cluster-ready health check job
  * Perform hosted control plane cluster post sc upgrade health checks
  * Perform MC pre upgrade health checks
  * Perform MC upgrade
  * Perform MC post upgrade health checks
    * Checks prometheus alerts
    * Runs osd-cluster-ready health check job
  * Perform hosted control plane cluster post mc upgrade health checks

The test suite is ordered and has labels which allow you to customize what is
performed. For example:

* `ginkgo run`
  * Will run through the entire test suite mentioned above
* `ginkgo run --label-filter="ApplyHCPWorkloads || RemoveHCPWorkloads || SCUpgrade || SCUpgradeHealthChecks"`
  * Will do everything except upgrade the management cluster
* `ginkgo run --label-filter="ApplyHCPWorkloads || RemoveHCPWorkloads || MCUpgrade || MCUpgradeHealthChecks"`
  * Will do everything except upgrade the service cluster
* `ginkgo run --label-filter="ApplyHCPWorkloads || SCUpgrade || MCUpgrade || SCUpgradeHealthChecks || MCUpgradeHealthChecks"`
  * Will do everything except remove the hcp clusters deployed

## How To Run

### Prerequisites

* Your RH OCM account must have all the necessary permissions for:
  * OSD fleet manager
  * Access OSD fleet manager clusters via oc client (kubeconfig)
  * Ability to deploy hosted control plane clusters
  * Ability to pin hosted control plane clusters
* You have run the command below to [create a service cluster](#create-service-cluster)
  and [retrieved the cluster id](#get-service-cluster-id)
* You have run the command below to [create a management cluster](#create-management-cluster)
  and [retrieved the cluster id](#get-management-cluster-id)
* You have run the command below to [get the provision shard id](#get-provision-shard-id)

### Run Ginkgo

```shell
AWS_REGION=<AWS_REGION> \
AWS_PROFILE=<AWS_PROFILE> \
OCM_TOKEN=<OCM_TOKEN> \
OSD_FLEET_MGMT_MANAGEMENT_CLUSTER_ID=<OSD_FLEET_MGMT_MANAGEMENT_CLUSTER_ID> \
OSD_FLEET_MGMT_SERVICE_CLUSTER_ID=<OSD_FLEET_MGMT_SERVICE_CLUSTER_ID> \
PROVISION_SHARD_ID=<PROVISION_SHARD_ID> \
UPGRADE_TYPE=<UPGRADE_TYPE> \
ginkgo run --timeout 10h
```

### Run Container Image

```shell
# Build image
make build-image

# Run container
export CONTAINER_ENGINE=<podman|docker>
$CONTAINER_ENGINE run -e OCM_TOKEN \
-e AWS_REGION=<AWS_REGION> \
-e AWS_PROFILE=<AWS_PROFILE> \
-e OSD_FLEET_MGMT_MANAGEMENT_CLUSTER_ID=<OSD_FLEET_MGMT_MANAGEMENT_CLUSTER_ID> \
-e OSD_FLEET_MGMT_SERVICE_CLUSTER_ID=<OSD_FLEET_MGMT_SERVICE_CLUSTER_ID> \
-e PROVISION_SHARD_ID=<PROVISION_SHARD_ID> \
-e UPGRADE_TYPE=<UPGRADE_TYPE> \
validate-mcscupgrade:latest
```

### Run Binary

```shell
make build
AWS_REGION=<AWS_REGION> \
AWS_PROFILE=<AWS_PROFILE> \
OCM_TOKEN=<OCM_TOKEN> \
OSD_FLEET_MGMT_MANAGEMENT_CLUSTER_ID=<OSD_FLEET_MGMT_MANAGEMENT_CLUSTER_ID> \
OSD_FLEET_MGMT_SERVICE_CLUSTER_ID=<OSD_FLEET_MGMT_SERVICE_CLUSTER_ID> \
PROVISION_SHARD_ID=<PROVISION_SHARD_ID> \
UPGRADE_TYPE=<UPGRADE_TYPE> \
mcscupgrade.test
```

## Commands

Each of the commands below has assumed you are authenticated with ocm:

```shell
ocm login --token $OCM_TOKEN --url integration
```

### Create Service Cluster

```shell
echo '{"region":"eu-central-1", "cloud_provider":"aws"}' | ocm post /api/osd_fleet_mgmt/v1/service_clusters
```

### Get Service Cluster ID

```shell
ocm get /api/osd_fleet_mgmt/v1/service_clusters -p search="sector is 'upgradetesting' and status is 'maintenance'" | jq -r .items[].id
```

### Delete Service Cluster

```shell
export SC_CLUSTER_ID=""
ocm delete /api/osd_fleet_mgmt/v1/service_clusters/$SC_CLUSTER_ID
ocm get /api/osd_fleet_mgmt/v1/service_clusters/$SC_CLUSTER_ID | jq .status
ocm delete /api/osd_fleet_mgmt/v1/service_clusters/$SC_CLUSTER_ID/ack
```

### Create Management Cluster

```shell
export SERVICE_CLUSTER_ID=""
echo '{"service_cluster_id":'\"$SERVICE_CLUSTER_ID\"'}' | ocm post /api/osd_fleet_mgmt/v1/management_clusters
```

### Get Management Cluster ID

```shell
ocm get /api/osd_fleet_mgmt/v1/management_clusters -p search="sector='upgradetesting'" | jq -r .items[].id
```

### Delete Management Cluster

```shell
export MC_CLUSTER_ID=""
ocm delete /api/osd_fleet_mgmt/v1/management_clusters/$MC_CLUSTER_ID
ocm get /api/osd_fleet_mgmt/v1/management_clusters/$MC_CLUSTER_ID | jq .status
ocm delete /api/osd_fleet_mgmt/v1/management_clusters/$MC_CLUSTER_ID/ack
```

### Get Management Cluster Kubeconfig

```shell
export MC_CLUSTER_ID=""
cluster_href=`ocm get /api/osd_fleet_mgmt/v1/management_clusters/$MC_CLUSTER_ID | jq -r .cluster_management_reference.href`
ocm get $cluster_href/credentials | jq -r .kubeconfig > $MC_CLUSTER_ID-kubeconfig
```

### Get Service Cluster Kubeconfig

```shell
export SC_CLUSTER_ID=""
cluster_href=`ocm get /api/osd_fleet_mgmt/v1/service_clusters/$SC_CLUSTER_ID | jq -r .cluster_management_reference.href`
ocm get $cluster_href/credentials | jq -r .kubeconfig > $SC_CLUSTER_ID-kubeconfig
```

### Get provision Shard ID

```shell
ocm get /api/osd_fleet_mgmt/v1/service_clusters -p search="sector is 'upgradetesting' and status is 'maintenance'" | jq -r .items[].provision_shard_reference.id
```
