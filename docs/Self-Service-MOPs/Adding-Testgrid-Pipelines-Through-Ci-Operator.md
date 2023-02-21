# **Adding Testgrid Pipelines Through Ci-Operator(release repo)**

OSDe2e offers an integration with ci-operator to add testgrid pipelines.  This document will walk through the steps to add a new testgrid pipeline to the release repo.
This functionality was added in this [PR](https://github.com/openshift/ci-tools/pull/3244).

## **Steps**
* The new testgrid pipeline using OSDe2e needs to be a prowgen job.
* This job does not have to live inside of the /release/master/ci-operator/config/openshift/osde2e directory. It can live in any directory and it is encouraged to live in the directory that is most relevant to the test.
* The name of the job created in Prow will need to be added to [_allow-list.yaml](https://github.com/openshift/release/blob/master/core-services/testgrid-config-generator/_allow-list.yaml) in the release repo.  This will allow the job to be added to the testgrid dashboard.
* OSDe2e has a custom tag 'osde2e' that adds this pipeline to [redhat-openshift-osd](https://testgrid.k8s.io/redhat-openshift-osd) dashboard.
* The other tags in this file are inherent to the release repo and can be found [Openshift CI Docs](https://docs.ci.openshift.org/docs/how-tos/add-jobs-to-testgrid/)

## **Results**
Once the PR has been merged, the testgrid pipeline will be added once the testgrid-config-generator runs.  This will happen every 24 hours.  The testgrid pipeline will be added to the [redhat-openshift-osd](https://testgrid.k8s.io/redhat-openshift-osd) dashboard.
An example PR can be found [here](https://github.com/openshift/release/pull/36578).