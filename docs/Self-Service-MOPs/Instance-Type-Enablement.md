# **Testing New Cloud Instance Types with OSDE2E**    

As Openshift expands support for new cloud instance types, it is important to ensure that the new instance types are tested and validated. This document outlines the process for testing new instance types with OSDE2E.

## **Prerequisites**
After SREP has enabled the new instance type in the cloud provider, the following prerequisites must be met before testing can begin:
* The new instance type must be enabled in Stage and available to query via OCM/ROSA clis. 
To verify that the instance type is available, run the following command(s):
For OSD:
``` 
ocm get /api/clusters_mgmt/v1/machine_types --parameter search="id LIKE '$INSTACE_TYPE.%'" | jq -r '.items[].id'
```
Note: When using jq your shell type might require slightly different syntax. The author is using zsh.
For ROSA:
```
rosa list instance-types | grep "$INSTANCE_TYPE"
```
The given output(s) should match the enabled instances types defined in the ticket.  If the instance type is not available, please contact SREP to ensure that the instance type is enabled in Stage.

* Depending on the instance type, the following may be required:
    * The new instance type must be enabled in the cloud provider's quota.  If the instance type is not enabled in quota, please contact SREP to ensure that the instance type is enabled in quota.
    * The new instance type must be enabled in the cloud provider's pricing.  If the instance type is not enabled in pricing, please contact SREP to ensure that the instance type is enabled in pricing.

## **Testing**
Once the prerequisites are met, the following steps should be taken to test the new instance type:
* We will be using the [release repo](https://github.com/openshift/release) to test the new instance types and leverage prowgen.
* Create a new PR for the [release repo osde2e jobs](https://github.com/openshift/release/blob/master/ci-operator/config/openshift/osde2e/openshift-osde2e-main.yaml)
* Copy and modify the following yaml snippet to create a new job for the new instance type:
```
- as: rosa-stage-e2e-machine-type-enablement-$INSTANCE_TYPE
  commands: |
    export REPORT_DIR="$ARTIFACT_DIR"
    export CONFIGS="rosa,e2e-suite"
    export ROSA_ENV="stage"
    export ROSA_STS="true"
    export ROSA_COMPUTE_MACHINE_TYPE="$INSTANCE_TYPE"
    export SECRET_LOCATIONS="/usr/local/osde2e-common,/usr/local/osde2e-credentials,/usr/local/osde2e-rosa-stage"
    /osde2e test --secret-locations ${SECRET_LOCATIONS} --configs ${CONFIGS}
  container:
    clone: true
    from: osde2e
  cron: "$CRON_TIME"
  secrets:
  - mount_path: /usr/local/osde2e-common
    name: osde2e-common
  - mount_path: /usr/local/osde2e-credentials
    name: osde2e-credentials
  - mount_path: /usr/local/osde2e-rosa-stage
    name: osde2e-rosa-stage
```
* The following variables will need to be set:
    * `as`: This is the name of the job.  It should be in the format of `rosa-stage-e2e-machine-type-enablement-$INSTANCE_TYPE`
    * `cron`: This is the cron for the job.  It should be set to twice a day, if setting multiple instances in the same PR please set the cron to different times, preferably 1 hours apart.
    * `ROSA_COMPUTE_MACHINE_TYPE`: This is the instance type that will be tested.  It should be set to the instance type that is being tested.
* The mounted secrets are the used to run these test using the SD-CICD account. These secrets are managed by us and will allow the test to run successfully.  If you are running the test locally, you will need to create your own secrets and mount them to the job.

* You'll have to run the release repos `make jobs` command to generate the prowgen jobs.
```
make jobs
```
* Once this is completed you will see changes to the following files:
    * `ci-operator/config/openshift/osde2e/openshift-osde2e-main.yaml`
    * `ci-operator/jobs/openshift/osde2e/openshift-osde2e-main-periodics.yaml`

* Commit and push the changes to the release repo and open a PR. You can reach out to the SD-CICD team for review in the #sd-cicd channel in slack and tag @sd-cicd-team for review.
Once the PR is merged, the jobs will be created and run automatically.

## **Results**
Once the jobs have run, you can view the results in the [prow](https://prow.ci.openshift.org/?job=*osde2e*machine*)
The test should pass three times consecutively before the instance type is considered validated.  If the test fails, please reach out to the SD-CICD team in the #sd-cicd channel in slack and tag @sd-cicd-team for review.