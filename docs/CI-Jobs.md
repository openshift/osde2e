# CI Jobs

The SD CICD team has a variety of periodic jobs testing different Managed
OpenShift flavors. Tests consist of verifying the cluster from the SREP
point of view which will ensure the cluster gets deployed, hive clusters
apply the necessary selector sync sets to the cluster, SREP operators get
deployed, verify SREP operators are functioning based on their functional
requirements and destroy the cluster.

This page provides you with direct links to all the active jobs. The
jobs will reside in both the *SD CICD osde2e space* along with jobs being
part of the *[OCP TRT release informing signal][OpenShift Release Gates]*.

*All periodic jobs are using OCM stage unless noted otherwise.*

*All jobs send alerts to the following slack channel: [#sd-cicd-alerts].*

## SD CICD Self Service Jenkins Job

The SD CICD team has a self service Jenkins job that allows users to run
OSDe2e tests using parameters. The job can be found [here](https://ci.int.devshift.net/view/osde2e/job/osde2e-parameterized-job/)
The job is using an AWS account for deployments. *******5153

## SD CICD Periodic Jobs

The table below represents the periodic jobs the SD CICD team manages as part
of the osde2e test framework. Overall dashboard can be found
[here][SD CICD Test Grid Dashboard].

| Job Name                                                     | OCP Version                         | Results                                             | AWS Account |
| ------------------------------------------------------------ | ----------------------------------- | --------------------------------------------------- | ----------- |
| ROSA BYOVPC Proxy Install                                    | Latest Pre-GA                       | [Test Grid][SD CICD ROSA BYOVPC Proxy Install]      | *******3696 |
| ROSA BYOVPC Proxy Post Install                               | Latest Pre-GA                       | [Test Grid][SD CICD ROSA BYOVPC Proxy Post Install] | *******3696 |
| OSD AWS Upgrade Latest Default Y Minus 1 To Latest Default Y | Latest Default Y-1 to Latest Y      | [Test Grid][SD CICD OSD AWS Upgrade Y-1 To Y]       | *******3696 |
| OSD AWS Upgrade Latest Default Z Minus 1 To Latest Default Z | Latest Default Z-1 to Latest Z      | [Test Grid][SD CICD OSD AWS Upgrade Z-1 To Z]       | *******3696 |
| OSD AWS Upgrade Latest Default Y To Latest Default Y Plus 1  | Latest Default Y to Latest Y Plus 1 | [Test Grid][SD CICD OSD AWS Upgrade Y To Y+1]       | *******3696 |
| OSD AWS Upgrade Latest Default Y Plus 1 To Latest Y          | Latest Default Y+1 to Latest Y      | [Test Grid][SD CICD OSD AWS Upgrade Y+1 To Y]       | *******3696 |
| OSD AWS SREP Operator Informing Suite                        | Latest Pre-GA                       | [Test Grid][SD CICD OSD AWS Informing Suite]        | *******3696 |

### OCP Versions

The [SD CICD Periodic Jobs](#sd-cicd-periodic-jobs) above may not always be
testing the same OpenShift version. One of the main reason for this is in
regards to ROSA STS jobs. STS policies may not exist yet for the latest pre-GA
version. Which means the job needs to install the latest GA'd version,
until the policies are enabled for the latest pre-GA version.

| OCP Version   | Definition                                                                           |
| ------------- | ------------------------------------------------------------------------------------ |
| Latest GA     | Installs the latest GA X.Y.Z version from the channel group specified by the job     |
| Latest Pre-GA | Installs the latest Pre-GA X.Y.Z version from the channel group specified by the job |

## TRT Nightly Periodic Jobs

The table below represents the periodic jobs that run to validate Managed
OpenShift for new nightly OCP TRT builds and provide a *informing* signal
to the release.

### OSD

| OCP Version | AWS                           | GCP                           | AWS Account |
| ----------- | ----------------------------- | ----------------------------- | ----------- |
| 4.14        | [Test Grid][4.14 TRT OSD AWS] | [Test Grid][4.14 TRT OSD GCP] | *******0241 |
| 4.13        | [Test Grid][4.13 TRT OSD AWS] | [Test Grid][4.13 TRT OSD GCP] | *******0241 |
| 4.12        | [Test Grid][4.12 TRT OSD AWS] | [Test Grid][4.12 TRT OSD GCP] | *******0241 |
| 4.11        | [Test Grid][4.11 TRT OSD AWS] | [Test Grid][4.11 TRT OSD GCP] | *******0241 |
| 4.10        | [Test Grid][4.10 TRT OSD AWS] | [Test Grid][4.10 TRT OSD GCP] | *******0241 |

### ROSA

| OCP Version | ROSA Classic STS                       | AWS Account | ROSA HCP                       | AWS Account |
| ----------- | -------------------------------------- | ----------- | ------------------------------ | ----------- |
| 4.14        | [Test Grid][4.14 TRT ROSA CLASSIC STS] | *******0241 | N/A                            | -           |
| 4.13        | [Test Grid][4.13 TRT ROSA CLASSIC STS] | *******0241 | [Test Grid][4.13 TRT ROSA HCP] | *******4366 |
| 4.12        | [Test Grid][4.12 TRT ROSA CLASSIC STS] | *******0241 | [Test Grid][4.12 TRT ROSA HCP] | *******4366 |
| 4.11        | [Test Grid][4.11 TRT ROSA CLASSIC STS] | *******0241 | N/A                            | -           |
| 4.10        | [Test Grid][4.10 TRT ROSA CLASSIC STS] | *******0241 | N/A                            | -           |

[SD CICD Test Grid Dashboard]: https://testgrid.k8s.io/redhat-openshift-osd
[SD CICD ROSA BYOVPC Proxy Install]: https://testgrid.k8s.io/redhat-openshift-osd#periodic-ci-openshift-osde2e-main-rosa-stage-e2e-byo-vpc-proxy-install&width=90
[SD CICD ROSA BYOVPC Proxy Post Install]: https://testgrid.k8s.io/redhat-openshift-osd#periodic-ci-openshift-osde2e-main-rosa-stage-e2e-byo-vpc-proxy-postinstall&width=90
[SD CICD OSD AWS Informing Suite]: https://testgrid.k8s.io/redhat-openshift-osd#periodic-ci-openshift-osde2e-main-aws-stage-informing-default&width=90
[SD CICD OSD AWS Upgrade Y-1 To Y]: https://testgrid.k8s.io/redhat-openshift-osd#periodic-ci-openshift-osde2e-main-osd-aws-upgrade-latest-default-y-minus-1-to-latest-default-y&width=90
[SD CICD OSD AWS Upgrade Z-1 To Z]: https://testgrid.k8s.io/redhat-openshift-osd#periodic-ci-openshift-osde2e-main-osd-aws-upgrade-latest-default-z-minus-1-to-latest-default-z&width=90
[SD CICD OSD AWS Upgrade Y To Y+1]: https://testgrid.k8s.io/redhat-openshift-osd#periodic-ci-openshift-osde2e-main-osd-aws-upgrade-latest-default-y-to-latest-y-plus-1&width=90
[SD CICD OSD AWS Upgrade Y+1 To Y]: https://testgrid.k8s.io/redhat-openshift-osd#periodic-ci-openshift-osde2e-main-osd-aws-upgrade-latest-default-y-plus-1-to-latest-y&width=90

[4.14 TRT OSD AWS]: https://testgrid.k8s.io/redhat-openshift-ocp-release-4.14-informing#release-openshift-ocp-osd-aws-nightly-4.14&width=90
[4.14 TRT OSD GCP]: https://testgrid.k8s.io/redhat-openshift-ocp-release-4.14-informing#release-openshift-ocp-osd-gcp-nightly-4.14&width=90
[4.13 TRT OSD AWS]: https://testgrid.k8s.io/redhat-openshift-ocp-release-4.13-informing#release-openshift-ocp-osd-aws-nightly-4.13&width=90
[4.13 TRT OSD GCP]: https://testgrid.k8s.io/redhat-openshift-ocp-release-4.13-informing#release-openshift-ocp-osd-gcp-nightly-4.13&width=90
[4.12 TRT OSD AWS]: https://testgrid.k8s.io/redhat-openshift-ocp-release-4.12-informing#release-openshift-ocp-osd-aws-nightly-4.12&width=90
[4.12 TRT OSD GCP]: https://testgrid.k8s.io/redhat-openshift-ocp-release-4.12-informing#release-openshift-ocp-osd-gcp-nightly-4.12&width=90
[4.11 TRT OSD AWS]: https://testgrid.k8s.io/redhat-openshift-ocp-release-4.11-informing#release-openshift-ocp-osd-aws-nightly-4.11&width=90
[4.11 TRT OSD GCP]: https://testgrid.k8s.io/redhat-openshift-ocp-release-4.11-informing#release-openshift-ocp-osd-gcp-nightly-4.11&width=90
[4.10 TRT OSD AWS]: https://testgrid.k8s.io/redhat-openshift-ocp-release-4.10-informing#release-openshift-ocp-osd-aws-nightly-4.10&width=90
[4.10 TRT OSD GCP]: https://testgrid.k8s.io/redhat-openshift-ocp-release-4.10-informing#release-openshift-ocp-osd-gcp-nightly-4.10&width=90

[4.14 TRT ROSA CLASSIC STS]: https://testgrid.k8s.io/redhat-openshift-ocp-release-4.14-informing#release-openshift-ocp-rosa-classic-sts-nightly-4.14&width=90
[4.13 TRT ROSA CLASSIC STS]: https://testgrid.k8s.io/redhat-openshift-ocp-release-4.13-informing#release-openshift-ocp-rosa-classic-sts-nightly-4.13&width=90
[4.12 TRT ROSA CLASSIC STS]: https://testgrid.k8s.io/redhat-openshift-ocp-release-4.12-informing#release-openshift-ocp-rosa-classic-sts-nightly-4.12&width=90
[4.11 TRT ROSA CLASSIC STS]: https://testgrid.k8s.io/redhat-openshift-ocp-release-4.11-informing#release-openshift-ocp-rosa-classic-sts-nightly-4.11&width=90
[4.10 TRT ROSA CLASSIC STS]: https://testgrid.k8s.io/redhat-openshift-ocp-release-4.10-informing#release-openshift-ocp-rosa-classic-sts-nightly-4.10&width=90

[4.13 TRT ROSA HCP]: https://testgrid.k8s.io/redhat-openshift-ocp-release-4.13-informing#release-openshift-ocp-rosa-hcp-nightly-4.13&width=90
[4.12 TRT ROSA HCP]: https://testgrid.k8s.io/redhat-openshift-ocp-release-4.12-informing#release-openshift-ocp-rosa-hcp-nightly-4.12&width=90

[#sd-cicd-alerts]: https://app.slack.com/client/T027F3GAJ/CNYM6PB6X

[OpenShift Release Gates]: https://docs.ci.openshift.org/docs/architecture/release-gating/
