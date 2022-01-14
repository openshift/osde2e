+++
title = "OSDe2e gcp Weather Report 2022-01-14 12:00:55.498664386 +0000 UTC"
author = "OSDe2e Automation"
date = "2022-01-14 12:00:55.498664386 +0000 UTC"
tags = ["weather-report", "gcp"]
summary = "<table class=\"summary\"><tr><td bgcolor=\"#29d600\"></td><td>int (Pass rate: 98.40)</td></tr><tr><td bgcolor=\"#13ec00\"></td><td>prod (Pass rate: 99.27)</td></tr><tr><td bgcolor=\"#20df00\"></td><td>stage (Pass rate: 98.75)</td></tr></table>"
+++
## Summary

| Job Name | Pass Rate | More detail |
|----------|-----------|-------------|
|[osde2e-int-gcp-e2e-next-z](https://prow.ci.openshift.org/?job=osde2e-int-gcp-e2e-next-z)| <span style="color:#29d600;">98.40%</span>|[More Detail](#osde2e-int-gcp-e2e-next-z)|
|[osde2e-prod-gcp-e2e-default](https://prow.ci.openshift.org/?job=osde2e-prod-gcp-e2e-default)| <span style="color:#0bf400;">99.60%</span>|[More Detail](#osde2e-prod-gcp-e2e-default)|
|[osde2e-prod-gcp-e2e-next](https://prow.ci.openshift.org/?job=osde2e-prod-gcp-e2e-next)| <span style="color:#15ea00;">99.20%</span>|[More Detail](#osde2e-prod-gcp-e2e-next)|
|[osde2e-prod-gcp-e2e-upgrade-to-latest-z](https://prow.ci.openshift.org/?job=osde2e-prod-gcp-e2e-upgrade-to-latest-z)| <span style="color:#1fe000;">98.80%</span>|[More Detail](#osde2e-prod-gcp-e2e-upgrade-to-latest-z)|
|[osde2e-stage-gcp-e2e-default](https://prow.ci.openshift.org/?job=osde2e-stage-gcp-e2e-default)| <span style="color:#24db00;">98.60%</span>|[More Detail](#osde2e-stage-gcp-e2e-default)|
|[osde2e-stage-gcp-e2e-next-z](https://prow.ci.openshift.org/?job=osde2e-stage-gcp-e2e-next-z)| <span style="color:#15ea00;">99.20%</span>|[More Detail](#osde2e-stage-gcp-e2e-next-z)|
|[osde2e-stage-gcp-e2e-upgrade-to-latest-z](https://prow.ci.openshift.org/?job=osde2e-stage-gcp-e2e-upgrade-to-latest-z)| <span style="color:#33cc00;">98.01%</span>|[More Detail](#osde2e-stage-gcp-e2e-upgrade-to-latest-z)|



## osde2e-int-gcp-e2e-next-z

Overall pass rate: <span style="color:#29d600;">98.40%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1481778451697373184](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-int-gcp-e2e-next-z/1481778451697373184) | 4.9.15-candidate |  | <span style="color:#29d600;">98.40%</span>|<ul><li>[install] [Suite: e2e] Cluster state should have no alerts</li><li>[install] [Suite: e2e] Workload (guestbook) should get created in the cluster</li></ul>



## osde2e-prod-gcp-e2e-default

Overall pass rate: <span style="color:#0bf400;">99.60%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1481898962523787264](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-default/1481898962523787264) | 4.9.12-candidate |  | <span style="color:#01fe00;">100.00%</span>|
[1481778458416648192](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-default/1481778458416648192) | 4.9.12-candidate |  | <span style="color:#15ea00;">99.20%</span>|<ul><li>[install] [Suite: e2e] Workload (guestbook) should get created in the cluster</li></ul>



## osde2e-prod-gcp-e2e-next

Overall pass rate: <span style="color:#15ea00;">99.20%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1481536787464589312](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-next/1481536787464589312) | 4.9.15-candidate |  | <span style="color:#15ea00;">99.20%</span>|<ul><li>[install] [Suite: e2e] Workload (guestbook) should get created in the cluster</li></ul>
[1481778459242926080](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-next/1481778459242926080) | 4.9.15-candidate |  | <span style="color:#15ea00;">99.20%</span>|<ul><li>[install] [Suite: e2e] Workload (guestbook) should get created in the cluster</li></ul>
[1481898963463311360](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-next/1481898963463311360) | 4.9.15-candidate |  | <span style="color:#15ea00;">99.20%</span>|<ul><li>[install] [Suite: operators] [OSD] RBAC Operator Operator Upgrade should upgrade from the replaced version</li></ul>



## osde2e-prod-gcp-e2e-upgrade-to-latest-z

Overall pass rate: <span style="color:#1fe000;">98.80%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1481657625635459072](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-upgrade-to-latest-z/1481657625635459072) | 4.9.12-candidate | 4.9.15 | <span style="color:#1fe000;">98.80%</span>|<ul><li>[install] [Suite: e2e] Workload (guestbook) should get created in the cluster</li><li>[upgrade] [Suite: e2e] Workload (guestbook) should get created in the cluster</li><li>[upgrade] [Suite: service-definition] [OSD] DaemonSets DaemonSets are not allowed empty node-label daemonset should get created</li></ul>



## osde2e-stage-gcp-e2e-default

Overall pass rate: <span style="color:#24db00;">98.60%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1481898969633132544](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-default/1481898969633132544) | 4.9.12-candidate |  | <span style="color:#15ea00;">99.20%</span>|<ul><li>[install] [Suite: e2e] Cluster state should have no alerts</li></ul>
[1481536794603294720](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-default/1481536794603294720) | 4.9.12-candidate |  | <span style="color:#29d600;">98.40%</span>|<ul><li>[install] [Suite: e2e] Workload (guestbook) should get created in the cluster</li><li>[install] [Suite: e2e] [OSD] namespace validating webhook namespace validating webhook Non-privileged users can manage all non-privileged namespaces</li></ul>
[1481657629011873792](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-default/1481657629011873792) | 4.9.12-candidate |  | <span style="color:#29d600;">98.40%</span>|<ul><li>[install] [Suite: e2e] Cluster state should have no alerts</li><li>[install] [Suite: e2e] Workload (guestbook) should get created in the cluster</li></ul>
[1481778474346614784](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-default/1481778474346614784) | 4.9.12-candidate |  | <span style="color:#29d600;">98.40%</span>|<ul><li>[install] [Suite: e2e] Cluster state should have no alerts</li><li>[install] [Suite: e2e] Workload (guestbook) should get created in the cluster</li></ul>



## osde2e-stage-gcp-e2e-next-z

Overall pass rate: <span style="color:#15ea00;">99.20%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1481536797107294208](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-next-z/1481536797107294208) | 4.9.15-candidate |  | <span style="color:#15ea00;">99.20%</span>|<ul><li>[install] [Suite: e2e] Workload (guestbook) should get created in the cluster</li></ul>
[1481657631054499840](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-next-z/1481657631054499840) | 4.9.15-candidate |  | <span style="color:#01fe00;">100.00%</span>|
[1481778476083056640](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-next-z/1481778476083056640) | 4.9.15-candidate |  | <span style="color:#29d600;">98.40%</span>|<ul><li>[install] [Suite: e2e] Cluster state should have no alerts</li><li>[install] [Suite: e2e] Workload (guestbook) should get created in the cluster</li></ul>



## osde2e-stage-gcp-e2e-upgrade-to-latest-z

Overall pass rate: <span style="color:#33cc00;">98.01%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1481536798424305664](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-upgrade-to-latest-z/1481536798424305664) | 4.9.12-candidate | 4.9.15 | <span style="color:#33cc00;">98.01%</span>|<ul><li>[install] [Suite: e2e] Workload (guestbook) should get created in the cluster</li><li>[install] [Suite: operators] [OSD] Splunk Forwarder Operator Operator Upgrade should upgrade from the replaced version</li><li>[upgrade] [Suite: e2e] ImageStreams should exist in the cluster</li><li>[upgrade] [Suite: e2e] Workload (guestbook) should get created in the cluster</li><li>[upgrade] [Suite: operators] [OSD] Splunk Forwarder Operator Operator Upgrade should upgrade from the replaced version</li></ul>




