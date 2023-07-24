# TRT OpenShift Release Gating Jobs

## Introduction

This page explains how SREP CI jobs are included as part of the
OpenShift release gating provided by the [Technical Release Team "TRT"][TRT].
These jobs run on new OpenShift nightly builds and verify the build for the
supported Managed OpenShift Platforms (e.g. OSD, ROSA Classic, ROSA HCP).

To learn more about OpenShift release gates, refer to this
[page][OpenShift Release Gates].

*Currently all SREP jobs are included as informing jobs. These jobs will
not block the release but provide a valuable signal into the nightly build.
These jobs will eventually be moved into blocking signal at a later date.
Requires [SDCICD-1058][SDCICD-1058] to be completed first.*

## Jobs

The jobs verifying Managed OpenShift Platforms for new OpenShift nightly builds
reside within [ci-operator/config/openshift/osde2e][OSDE2E Prowgen Job Configs].
They are the multi step jobs and the actual jobs reside within
[ci-operator/jobs/openshift/osde2e][OSDE2E Prowgen Jobs].

### Add Job

To include a new job to be part of the release signal (informing or blocking),
you can refer to this [page][Add Job] for TRT official documentation. 2 PR's
will need to be opened to officially add the job:

1. PR to the [release][Release Repo] repository to add the periodic job. An
   example can be seen [here][Add SDCICD Job].
   1. Add a new file under [ci-operator/config/openshift/osde2e][OSDE2E Prowgen Job Configs]
   if the OpenShift version does not exist or add the new prowgen job to the
   existing files.
   2. Generate the actual job from the job configuration `make jobs`.
2. PR to the [continuous-release-jobs][Continuous Release Jobs Repo] repository
   to add the signal notification. Refer to this [PR][Add Periodic Signal] as
   an example.

### Remove Job

In the event a job needs to be removed (e.g. OCP version is EOL),
a PR will need to be opened to the [release][Release Repo] repository to
remove the job.

To remove a periodic job (when OCP version goes EOL), 2 PR's need to be opened:

1. PR to the [release][Release Repo] repository to remove the periodic job.
   1. Remove the file under [ci-operator/config/openshift/osde2e][OSDE2E Prowgen Job Configs]
   if need to remove all jobs for the given OpenShift version or remove the prowgen
   job from one of the existing files.
2. PR to the [continuous-release-jobs][Continuous Release Jobs Repo] repository
   to remove the signal notification. Refer to this
   [PR][Remove Periodic Signal] as an example.

Once both PR's are merged, the previous periodic job will be removed/no
longer run.

[Add Job]: https://docs.ci.openshift.org/docs/architecture/release-gating/
[Add SDCICD Job]: https://github.com/openshift/release/pull/41245
[Add Periodic Signal]: https://github.com/openshift/continuous-release-jobs/pull/1286
[Continuous Release Jobs Repo]: https://github.com/openshift/continuous-release-jobs
[OpenShift Release Gates]: https://docs.ci.openshift.org/docs/architecture/release-gating/
[OSDE2E Prowgen Job Configs]: https://github.com/openshift/release/tree/master/ci-operator/config/openshift/osde2e
[OSDE2E Prowgen Jobs]: https://github.com/openshift/release/tree/master/ci-operator/jobs/openshift/osde2e
[Release Repo]: https://github.com/openshift/release
[Remove Periodic Signal]: https://github.com/openshift/continuous-release-jobs/pull/1287
[SDCICD-1058]: https://issues.redhat.com//browse/SDCICD-1058
[TRT]: https://docs.ci.openshift.org/docs/release-oversight/the-technical-release-team/
