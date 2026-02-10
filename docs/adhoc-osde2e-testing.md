# 1 Introduction

Ad Hoc OSDe2e test job runs may be useful for e2e testing a new cloud region, a new instance type, gap analysis or validating a particular issue.
# 2 Procedure

Prerequisite: Connect to Red Hat VPN

1. Navigate to https://ci.int.devshift.net/blue/organizations/jenkins/osde2e-parameterized-job/activity
2. Click Login if necessary (button in upper-right corner)
3. Click "Run" to create a new osde2e build
4. Review and update parameters as needed. See "Test Scenarios" below.
5. Click "Run" to initiate the with parameters build

Inspect build success. See Troubleshooting below

**Test Scenarios**

Use the following parameters guide for various test scenarios:

***AWS***

* Cloud Provider Region: Select appropriate AWS region ID, e.g. `me-central-1`
* Instance Type/Machine Type: For instance type testing: update the instance type, e.g. `x1e.xlarge`, otherwise leave empty to select default instance type.
* Osde2e Configs: For OSDe2e config examples (see [OSDe2e config documentation for values](https://github.com/openshift/osde2e/blob/main/docs/Config.md#common-config-flag-values)):
    * run ROSA sanity test suite in stage env: `rosa,stage,sanity`
    * run full OSDe2e test suite in ROSA production env: `rosa,prod,e2e-suite`

***GCP***
* OCM_CCS: Uncheck for GCP non-CCS, check for GCP CCS. For CCS, the gcp project used for CCS is `osde2e-ccs`.
* Cloud Provider Region: Select appropriate GCP based region ID, e.g. `southamerica-west1`
* Instance Type/Machine Type: Leave empty unless specific machine type is needed.  When unselected, a random machine type from a list of machine types supported by GCP OSD will be used.  For instance type enablement testing: update to specify machine type e.g. `custom-8-32768`.
* Osde2e Configs: For OSDe2e config examples (see [OSDe2e config documentation for values](https://github.com/openshift/osde2e/blob/main/docs/Config.md#common-config-flag-values)):
    * run sanity test suite in stage env: `gcp,stage,sanity`
    * run sanity test suite in prod env: `gcp,prod,sanity`
    * run full OSDe2e test suite in GCP production env: `gcp,prod,e2e-suite`


# 3 Troubleshooting

* Review Jenkins build logs
* Ping [@sd-cicd-team](https://redhat-internal.slack.com/admin/user_groups) in [#sd-cicd](https://redhat-internal.slack.com/archives/CMK13BP4J)

# 4 References

* Jenkins OSDe2e parameterized job: https://ci.int.devshift.net/blue/organizations/jenkins/osde2e-parameterized-job
* OSDe2e CI jobs: https://github.com/openshift/osde2e/blob/main/docs/CI-Jobs.md
* OSDe2e config values: https://github.com/openshift/osde2e/blob/main/docs/Config.md#common-config-flag-values
