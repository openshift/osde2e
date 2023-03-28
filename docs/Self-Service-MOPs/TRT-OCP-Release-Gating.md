# TRT Release Gating Jobs

## Introduction

This page explains how SREP periodic CI jobs are included as part of the
OpenShift release gates provided by the [Technical Release Team "TRT"][TRT].
These jobs run on new OCP TRT nightly builds and verify the payload for a
Managed OpenShift Platform.

To learn more about OpenShift release gates, refer to this
[page][OpenShift Release Gates].

*Currently all SREP jobs are included as informative jobs. These jobs will
not block the release but provide a valuable signal into the payload.
Eventually these jobs can be moved into blocking at a later date.*

## Jobs

TRT jobs live [here][Release Periodic Folder] within the
[release][Release Repo] repository. Example test grid
[dashboard][OpenShift 4.13 Informing Dashboard] for OpenShift 4.13
informing jobs.

### Add Job

To include a periodic job as part of the informing signal, refer to this
[page][Add Informative Job] for TRT official documentation. 2 PR's will need
to be opened to officially add the job to the informing signal:

1. PR to the [release][Release Repo] repository to add the periodic jobs.
2. PR to the [continuous-release-jobs][Continuous Release Jobs Repo] repository
   to add the signal notification. Refer to this [PR][Add Periodic Signal] as
   an example.

### Remove Job

Once a periodic job has been added as part of the release gate jobs. It will
continue to run once the OCP version has GA'd. After that, the cadence will be
shortened to run less often. When a new OCP version is under development,
the previous periodic job will be carried through.

To remove a periodic job (when OCP version goes EOL), 2 PR's need to be opened:

1. PR to the [release][Release Repo] repository to remove the periodic job.
   Refer to this [PR][Remove Periodic Informing Job] as an example.
2. PR to the [continuous-release-jobs][Continuous Release Jobs Repo] repository
   to remove the signal notification. Refer to this
   [PR][Remove Periodic Signal] as an example.

Once both PR's are merged, the existing periodic job will be removed/no
longer run.

[Add Informative Job]: https://docs.ci.openshift.org/docs/architecture/release-gating/#add-a-periodic-informative-job
[Add Periodic Signal]: https://github.com/openshift/continuous-release-jobs/pull/1286
[Continuous Release Jobs Repo]: https://github.com/openshift/continuous-release-jobs
[OpenShift 4.13 Informing Dashboard]: https://testgrid.k8s.io/redhat-openshift-ocp-release-4.13-informing
[OpenShift Release Gates]: https://docs.ci.openshift.org/docs/architecture/release-gating/
[Release Periodic Folder]: https://github.com/openshift/release/tree/master/ci-operator/jobs/openshift/release
[Release Repo]: https://github.com/openshift/release
[Remove Periodic Informing Job]: https://github.com/openshift/release/pull/37716
[Remove Periodic Signal]: https://github.com/openshift/continuous-release-jobs/pull/1287
[TRT]: https://docs.ci.openshift.org/docs/release-oversight/the-technical-release-team/
