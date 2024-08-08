# Rosa SDN to OVN Migration Test Suite


This folder contains the test suite to validate rosa cluster upgrades and network
migration from SDN to OVN for 14.14.x cluster.


The test suite can be easily run by invoking `ginkgo run` directly,
from a `compiled binary`
## Test Suite Model

Below outlines what the test suite is capable of:

* Before Suite
    * Input data validation
    * Deploy ROSA 14.14 cluster with SDN network type
* After Suite
    * Destroy the deployed ROSA cluster
* Test Cases:
    * Perform Pre Upgrade Check
    * Perform upgrade
    * Perform post upgrade health checks
        * Checks prometheus alerts
        * Runs osd-cluster-ready health check job
    * Perform sdn to ovn migration 
    * Perform post migration health checks
        * Checks prometheus alerts
        * Runs osd-cluster-ready health check job

The test suite is ordered and has labels which allow you to customize what is
performed. For example:

* `ginkgo run`
    * Will run through the entire test suite mentioned above
* `ginkgo run --label-filter="PostMigrationCheck || RosaUpgrade || PostUpgradeCheck || SdnToOvn || RemoveRosaCluster ""`
    * Will do everything except create a cluster
* `ginkgo run --label-filter="SdnToOvn"`
    * Will only do the sdn to ovn migration

## How To Run

### Prerequisites

* RH OCM account with long-lived token

### Run Ginkgo
*NOTE: CLUSTER_ID is optional (internal ID)
```shell
AWS_REGION=<AWS_REGION> \
AWS_SECRET_ACCESS_KEY=<AWS_SECRET_ACCESS_KEY> \
AWS_ACCESS_KEY_ID=<AWS_ACCESS_KEY_ID> \
OCM_TOKEN=<OCM_TOKEN> \
CLUSTER_ID=<CLUSTER_ID> \
UPGRADE_TYPE=<UPGRADE_TYPE> \
ginkgo run 
```
