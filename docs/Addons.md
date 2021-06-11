# **Add-On Testing**

This document describes the requirements to configure E2E testing of an Addon within `osde2e`. This is only one part of the overall process of onboarding an addon to OSD. The full process is outlined in the documentation [available here](https://gitlab.cee.redhat.com/service/managed-tenants/-/tree/master).

## **The Structure of an Addon Test**

How an add-on is tested can vary between groups and projects. In light of this, we've tried to create a very flexible and unopinionated framework for your testing. Your test harness should take the form of an OCI (docker) container that does the following:

*   Assume it is executing in a pod within an OpenShift cluster.
*   Assume the pod will inherit `cluster-admin` rights. 
*	Block until your addon is ready to be tested (we will launch your container after requesting installation of the addon, but we can't control when the addon is finished installing).
*   Output a valid `junit.xml` file to the `/test-run-results` directory before the container exits.
*   Output metadata to `addon-metadata.json` in the `/test-run-results` directory.

The [Prow Operator Test] is a good example of a [Basic operator test]. It verifies that the Prow operator and all the necessary CRDs are installed in the cluster. 

## **Test Environments**

We have three test environments: integration (int), staging (stage), and production (prod). Your job will probably want to be configured for all of them once you have gained confidence in your test harness. Each environment requires a separate prow job configuration. The next section covers prow configuration in detail.

## SKUs and Quota

In order to provision OSD and install your addon, our OCM token will need to have a quota of OSD clusters and installations of your addon available. In order to allocate quota for your addon, it must be assigned a SKU. You can request a SKU [by following these instructions](https://gitlab.cee.redhat.com/service/managed-tenants/-/tree/master).

Once you have a SKU, you'll need to also allocate quota to test within [`app-interface`](https://gitlab.cee.redhat.com/service/app-interface/#manage-openshift-resourcequotas-via-app-interface-openshiftquota-1yml). Quota is allocated independently in each of `int`, `stage`, and `prod` (different instances of OCM), so you'll need to allocate quota three times.

[Here](https://gitlab.cee.redhat.com/service/ocm-resources/-/blob/master/data/uhc-production/orgs/13215750.yaml#L13) is an example of SD-CICD's quota for production.

You need to open an MR to update the `SDCICD` org's quota so that it can provision your addon (as well as bumping the number of CCS clusters by 2 or so). You'll need to modify the following three files:

- [Our production quota](https://gitlab.cee.redhat.com/service/ocm-resources/-/blob/master/data/uhc-production/orgs/13215750.yaml)
- [Our stage quota](https://gitlab.cee.redhat.com/service/ocm-resources/-/blob/master/data/uhc-stage/orgs/13215750.yaml)
- [Our integration quota](https://gitlab.cee.redhat.com/service/ocm-resources/-/blob/master/data/uhc-integration/orgs/13215750.yaml)

Please bump the quota for SKU `MW00530` by 2 so that we can provision additional CCS clusters for you!

### Providing Secrets to Your Build

If you are not a part of the public GitHub Organization `OpenShift`, join it by following [these instructions](https://source.redhat.com/groups/public/atomicopenshift/atomicopenshift_wiki/setting_up_your_accounts_openshift).

Follow the documentation [here](https://docs.ci.openshift.org/docs/how-tos/adding-a-new-secret-to-ci/) to create secrets and configure them to be mirrored into the `ci` namespace [like ours](https://github.com/openshift/release/blob/master/core-services/secret-mirroring/_mapping.yaml#L62).

You'll need to provide some additional details about your AWS account in a secret. In particular, you'll need to provide these values in your credentials secret:

```
ocm-aws-account
ocm-aws-access-key
ocm-aws-secret-access-key
```

## **Configuring OSDe2e**

Once a test harness has been written, an OSDe2e test needs to be configured to install the desired add-on, then run the test harness against it. This is done by creating a PR ([example PR]) against the [openshift/release] repo. 

Your PR will instruct Prow (our CI service) to run a pod on some schedule. You will then specify that the job should run `osde2e` with some specific environment variables and flags that describe your test harness.

For addon testing, `osde2e` uses four primary environment variables: `ADDON_IDS`, `ADDON_TEST_HARNESSES`, `ADDON_TEST_USER`, and `ADDON_PARAMETERS`.

The first two of these are comma-delimited lists when supplied by environment variables. `ADDON_IDS` informs OSDe2e which addons to install once a cluster is healthy. `ADDON_TEST_HARNESSES` is a list of addon test containers to run as pods within the test cluster. 

`ADDON_TEST_USER` will specify the in-cluster user that the test harness containers will run as. It allows for a single wildcard (`%s`) which will automatically be evaluated as the OpenShift Project (namespace) the test harness is created under.

`ADDON_PARAMETERS` allows you to configure parameters that will be passed to OCM for the installation of your addon. The format is a two-level JSON object. The outer object's keys are the IDs of addons, and the inner objects are key-value pairs that will be passed to the associated addon.

An example prow job that configures the "prow" operator in the stage environment:

```yaml
- agent: kubernetes
  cluster: build02
  cron: 0 */12 * * *
  decorate: true
  extra_refs:
  - base_ref: main
    org: openshift
    repo: osde2e
  labels:
    pj-rehearse.openshift.io/can-be-rehearsed: "false"
  name: osde2e-stage-aws-addon-prow-operator
  spec:
    containers:
    - args:
      - test
      - --secret-locations
      - $(SECRET_LOCATIONS)
      - --configs
      - $(CONFIGS)
      command:
      - /osde2e
      env:
      - name: ADDON_IDS
        value: prow-operator
      - name: OCM_CCS
        value: "true"
      - name: ADDON_TEST_HARNESSES
        value: quay.io/miwilson/prow-operator-test-harness
      - name: CONFIGS
        value: aws,stage,addon-suite
      - name: SECRET_LOCATIONS
        value: /usr/local/osde2e-common,/usr/local/osde2e-credentials,/usr/local/prow-operator-credentials
      image: quay.io/app-sre/osde2e
      imagePullPolicy: Always
      name: ""
      resources:
        requests:
          cpu: 10m
      volumeMounts:
      - mountPath: /usr/local/osde2e-common
        name: osde2e-common
        readOnly: true
      - mountPath: /usr/local/osde2e-credentials
        name: osde2e-credentials
        readOnly: true
      - mountPath: /usr/local/prow-operator-credentials
        name: prow-operator-credentials
        readOnly: true
    serviceAccountName: ci-operator
    volumes:
    - name: osde2e-common
      secret:
        secretName: osde2e-common
    - name: osde2e-credentials
      secret:
        secretName: osde2e-credentials
    - name: prow-operator-credentials
      secret:
        secretName: prow-operator-credentials
```

To adapt this to your job, you would redefine the `ADDON_IDS` and `ADDON_TEST_HARNESSES`, as well as potentially adding some of the other variables discussed above.

You will *also* need to provide your own secrets by swapping the `prow-operator-credentials` above with your job's secrets. Note that we load osde2e's credentials, followed by the ones you supply. This allows your credentials to override any duplicate credentials supplied in our config.

> *NOTE*: If you want your job to run in a different environment, such as `int` or `prod`, you need to both change its name to include the proper environment *and* redefine the `CONFIGS` environment variable by replacing `stage` with the name of the appropriate environment.

You can change the cron scheduling of the job as well.

### ROSA Testing

If you would like to test against ROSA instead of plain OSD, you will need to change your configs from using `aws` to `rosa`, and you will need to supply differently-named secrets.

```
rosa-aws-account
rosa-aws-access-key
rosa-aws-secret-access-key
```

These secrets need the same values as the ones prefixed with `ocm-` described in [Providing Secrets to Your Build](#providing-secrets-to-your-build) above.

You should also provide these environment variables:

- `ROSA_AWS_REGION` (we recommend setting this to `random`)
- `ROSA_ENV` (set this to `integration`, `stage`, or `production` based on the environment the test is executing within).

For an example build that tests an Addon on ROSA, search [this file](https://github.com/openshift/release/blob/master/ci-operator/jobs/openshift/osde2e/openshift-osde2e-main-periodics.yaml) for the configuration of `osde2e-stage-rosa-addon-prow-operator`.

### Addon Cleanup

If your addon test creates or affects anything outside of the OSD cluster lifecycle, a separate cleanup action is required. If `ADDON_RUN_CLEANUP` is set to `true`, OSDe2e will run your test harness container a **second time** passing the argument `cleanup` to the container (as the first command line argument).

There may be a case where a separate cleanup container/harness is required. That may be configured using the `ADDON_CLEANUP_HARNESSES` config option. It is formatted in the same way as `ADDON_TEST_HARNESSES`. This however, may cause some confusion as to what is run when:

`ADDON_RUN_CLEANUP` is true, and `ADDON_CLEANUP_HARNESSES` is not set, OSDe2e will only run `ADDON_TEST_HARNESSES` again, passing the `cleanup` argument.

`ADDON_RUN_CLEANUP` is true, and `ADDON_CLEANUP_HARNESSES` is set, OSDe2e will only run the `ADDON_CLEANUP_HARNESSES`, passing no arguments.

> *NOTE*: Your OSD clusters will automatically back themselves up to S3 in your AWS account. You can find these backups by running `aws s3 ls --profile osd`. You should probably clean them up as part of the cleanup phase of your build.
 
### Slack Notifications

If you want to be notified of the results of your builds in slack, you can take advantage of [this feature](https://docs.ci.openshift.org/docs/how-tos/notification/). [Here](https://github.com/openshift/release/pull/16674/files#diff-d214756a87b37f0ad838abce8ddfa8993c7cd6a7614fc15384f5f3e4307f079aR1983) is an example PR of someone configuring slack alerts for an Addon.

## **Querying results from Datahub**

Once your job has been running in prow, you will be able to programmatically query Thanos/Prometheus for job results. All OSDe2e data points stored within Thanos/Prometheus are prefixed with `cicd_`. Currently there are three primary metrics stored:

```
cicd_event{environment="int",event="InstallSuccessful",install_version="openshift-v4.2.0-0.nightly-2020-01-15-224532",job="periodic-ci-openshift-osde2e-master-e2e-int-4.2-4.2",monitor="datahub",upgrade_version="openshift-v4.2.0-0.nightly-2020-01-15-231532"}

cicd_jUnitResult{environment="int",install_version="openshift-v4.2.0-0.nightly-2020-01-15-224532",job="periodic-ci-openshift-osde2e-master-e2e-int-4.2-4.2",monitor="datahub",phase="install",result="failed",suite="OSD e2e suite",testname="[OSD] Managed Velero Operator deployment should have all desired replicas ready",upgrade_version="openshift-v4.2.0-0.nightly-2020-01-15-231532"}

cicd_metadata{cluster_id="1a2bc3",environment="int",install_version="openshift-v4.2.0-0.nightly-2020-01-15-224532",job="periodic-ci-openshift-osde2e-master-e2e-int-4.2-4.2",job_id="123",metadata_name="time-to-cluster-ready",monitor="datahub",phase="",upgrade_version="openshift-v4.2.0-0.nightly-2020-01-15-231532"}

cicd_addon_metadata{cluster_id="1a2bc3",environment="int",install_version="openshift-v4.2.0-0.nightly-2020-01-15-224532",job="periodic-ci-openshift-osde2e-master-e2e-int-4.2-4.2",job_id="123",metadata_name="time-to-cluster-ready",monitor="datahub",phase="",upgrade_version="openshift-v4.2.0-0.nightly-2020-01-15-231532"}
```

In addition to programmatically gating your addon releases, you can also use the [Grafana instance] hosted by DataHub to build out a dashboard and alerting to monitor the health of the addon as versions change.

[Prow Operator Test]:https://github.com/meowfaceman/prow-operator-test-harness
[Basic operator test]:https://github.com/openshift/osde2e#operator-testing
[example PR]:https://github.com/openshift/release/pull/6721/files
[openshift/release]:https://github.com/openshift/release
[Managing Organization Quota]:https://gitlab.cee.redhat.com/service/ocm-resources/blob/master/docs/quota.md
[https://cloud.redhat.com/openshift/token]:https://cloud.redhat.com/openshift/token
[these instructions]:https://github.com/openshift/release/blob/master/core-services/secret-mirroring/README.md
[Grafana instance]:https://grafana.datahub.redhat.com/
