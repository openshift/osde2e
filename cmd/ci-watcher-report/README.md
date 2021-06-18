# ci-watcher-report

This CLI prints a report of active issues that fall within the scope of the CI Watcher.

It will print the pagerduty incidents, their alert counts, their status, the names of pipelines that generated them, a link to the PD incident, and any notes attached to the incident.

To use it, you must supply the `PAGERDUTY_TOKEN` environment variable set to your own PD token.

A sample usage:

```sh
$ PAGERDUTY_TOKEN=(pass tokens/pagerduty/cwaldon) go run ./cmd/ci-watcher-report/
[Suite: operators] [OSD] Splunk Forwarder Operator Operator Upgrade should upgrade from the replaced version failed
23  acknowledged
osde2e-prod-aws-e2e-next
osde2e-prod-gcp-e2e-upgrade-to-next-y
osde2e-prod-aws-e2e-upgrade-prod-minus-two-to-next
osde2e-prod-aws-e2e-upgrade-to-latest
osde2e-stage-gcp-e2e-next-z
osde2e-prod-aws-e2e-upgrade-prod-plus-one-to-latest
https://redhat.pagerduty.com/incidents/P1A7Z4U
https://issues.redhat.com/browse/OSD-7532

[Suite: operators] [OSD] Custom Domains Operator Should allow dedicated-admins to create domains Should be resolvable by external services failed
11  acknowledged
osde2e-stage-aws-e2e-default
osde2e-prod-aws-e2e-upgrade-prod-minus-two-to-next
osde2e-prod-aws-e2e-default
osde2e-stage-aws-e2e-next-z
osde2e-stage-aws-nightly-4.8
https://redhat.pagerduty.com/incidents/P5SPF5U
https://issues.redhat.com/browse/OSD-7533

[Suite: e2e] Pods should be Running or Succeeded failed
6   triggered
osde2e-stage-rosa-e2e-next-y
osde2e-prod-aws-e2e-next
osde2e-stage-gcp-e2e-next-z
https://redhat.pagerduty.com/incidents/PB6ZLO0

```
