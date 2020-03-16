# **Add-On Testing**


## **Test Requirements**

How an add-on is tested can vary between groups and projects. In light of this, there are a few requirements for an add-on test harness to be integrated into OSDe2e. The test should:



*   Assume it is executing in a pod within an OpenShift cluster. This means once the test code is written, it needs to be packaged into a container image.
*   Output a valid `junit.xml` file to the `/test-run-results` directory.
*   Output metadata to `addon-metadata.json` in the `/test-run-results` directory.

The [Prow Operator Test](https://github.com/meowfaceman/prow-operator-test-harness) is a good example of a [Basic operator test](https://github.com/openshift/osde2e#operator-testing). It verifies that the Prow operator and all the necessary CRDs are installed in the cluster. 


## **Configuring OSDe2e**

Once a test harness has been written, an OSDe2e test needs to be configured to install the desired add-on, then run the test harness against it. This is done by creating a PR ([example](https://github.com/openshift/release/pull/6721/files)) against the [openshift/release](https://github.com/openshift/release) repo. 

Regarding addon testing, OSDe2e has two primary config options: `ADDON_IDS` and `ADDON_TEST_HARNESSES`. Both of these are comma-delimited lists when supplied by environment variables, or YAML arrays when using the YAML config. `ADDON_IDS` informs OSDe2e which addons to install once a cluster is healthy. `ADDON_TEST_HARNESSES` is a list of addon test containers to run as pods within the test cluster. 

```
env:
- name: ADDON_IDS
  value: prow-operator
- name: ADDON_TEST_HARNESSES
  value: quay.io/miwilson/prow-operator-test-harness
```

### **Getting an OCM refresh token for your tests**

You will need to request an OCM refresh token in order to run your tests. The easiest way to do this is to visit [https://cloud.redhat.com/openshift/token](https://cloud.redhat.com/openshift/token) and copy the OFFLINE_REFRESH_TOKEN. 

Your account will need the following permissions:

*   Credentials API access
*   ...

### **Configuring your job to use your OCM refresh token**

In order to run addon tests in osde2e, you will need to create a secret in Origin CI with your OCM refresh token. Please follow [these instructions](https://github.com/openshift/release/blob/e877df16a32be22b60a62b6313ef3e0fe2e9256b/core-services/secret-mirroring/README.md) to both create a secret and a secret mapping into the ci namespace.

## **Querying results from Datahub**

Once your job has been running in prow, you will be able to programmatically query Thanos/Prometheus for job results. All OSDe2e data points stored within Thanos/Prometheus are prefixed with `cicd_`. Currently there are three primary metrics stored:

```
cicd_event{environment="int",event="InstallSuccessful",install_version="openshift-v4.2.0-0.nightly-2020-01-15-224532",job="periodic-ci-openshift-osde2e-master-e2e-int-4.2-4.2",monitor="datahub",upgrade_version="openshift-v4.2.0-0.nightly-2020-01-15-231532"}

cicd_jUnitResult{environment="int",install_version="openshift-v4.2.0-0.nightly-2020-01-15-224532",job="periodic-ci-openshift-osde2e-master-e2e-int-4.2-4.2",monitor="datahub",phase="install",result="failed",suite="OSD e2e suite",testname="[OSD] Managed Velero Operator deployment should have all desired replicas ready",upgrade_version="openshift-v4.2.0-0.nightly-2020-01-15-231532"}

cicd_metadata{cluster_id="1a2bc3",environment="int",install_version="openshift-v4.2.0-0.nightly-2020-01-15-224532",job="periodic-ci-openshift-osde2e-master-e2e-int-4.2-4.2",job_id="123",metadata_name="time-to-cluster-ready",monitor="datahub",phase="",upgrade_version="openshift-v4.2.0-0.nightly-2020-01-15-231532"}
```

In addition to programmatically gating your addon releases, you can also use the [Grafana instance](https://grafana.datahub.redhat.com/) hosted by DataHub to build out a dashboard and alerting to monitor the health of the addon as versions change.
