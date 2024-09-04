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

* `ginkgo run --label-filter="DefaultBuild"`
    * Will run through the entire test suite mentioned above
* `ginkgo run --label-filter="DefaultBuildWithProxy"`
    * Will run through the entire test suite against a cluster with a cluster wide proxy
* `ginkgo run --label-filter="SdnToOvn"`
    * Will only do the sdn to ovn migration

## How To Run

### Prerequisites

* RH OCM account with long-lived token
* The proxy needs to be created manually if end-to-end tests are going to be run against a cluster with a cluster-wide proxy

### Run Ginkgo
*NOTE:CLUSTER_ID is optional (internal ID), and AWS_HTTP_PROXY, AWS_HTTP_PROXYS, CA_BUNDLE, and 
SUBNETS are also optional unless end-to-end tests need to be run against a cluster with a cluster-wide proxy
```shell
AWS_REGION=<AWS_REGION> \
AWS_SECRET_ACCESS_KEY=<AWS_SECRET_ACCESS_KEY> \
AWS_ACCESS_KEY_ID=<AWS_ACCESS_KEY_ID> \
OCM_TOKEN=<OCM_TOKEN> \
CLUSTER_ID=<CLUSTER_ID> \
CLUSTER_NAME = <CLUSTER_NAME> \
REPLICAS=<REPLICAS>\
AWS_HTTP_PROXY=<AWS_HTTP_PROXY>\
AWS_HTTPs_PROXY=<AWS_HTTP_PROXY> \
CA_BUNDLE=<CA_BUNDLE> \
SUBNETS=<SUBNETS> \
ginkgo run 
```
