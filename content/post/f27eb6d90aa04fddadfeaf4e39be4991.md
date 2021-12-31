+++
title = "OSDe2e gcp Weather Report 2021-12-31 12:00:46.999686953 +0000 UTC"
author = "OSDe2e Automation"
date = "2021-12-31 12:00:46.999686953 +0000 UTC"
tags = ["weather-report", "gcp"]
summary = "<table class=\"summary\"><tr><td bgcolor=\"#15ea00\"></td><td>int (Pass rate: 99.20)</td></tr><tr><td bgcolor=\"#0df200\"></td><td>prod (Pass rate: 99.51)</td></tr><tr><td bgcolor=\"#0bf400\"></td><td>stage (Pass rate: 99.60)</td></tr></table>"
+++
## Summary

| Job Name | Pass Rate | More detail |
|----------|-----------|-------------|
|[osde2e-int-gcp-e2e-next-z](https://prow.ci.openshift.org/?job=osde2e-int-gcp-e2e-next-z)| <span style="color:#15ea00;">99.20%</span>|[More Detail](#osde2e-int-gcp-e2e-next-z)|
|[osde2e-prod-gcp-e2e-default](https://prow.ci.openshift.org/?job=osde2e-prod-gcp-e2e-default)| <span style="color:#06f900;">99.80%</span>|[More Detail](#osde2e-prod-gcp-e2e-default)|
|[osde2e-prod-gcp-e2e-next](https://prow.ci.openshift.org/?job=osde2e-prod-gcp-e2e-next)| <span style="color:#10ef00;">99.40%</span>|[More Detail](#osde2e-prod-gcp-e2e-next)|
|[osde2e-prod-gcp-e2e-upgrade-to-latest-z](https://prow.ci.openshift.org/?job=osde2e-prod-gcp-e2e-upgrade-to-latest-z)| <span style="color:#1fe000;">98.80%</span>|[More Detail](#osde2e-prod-gcp-e2e-upgrade-to-latest-z)|
|[osde2e-stage-gcp-e2e-default](https://prow.ci.openshift.org/?job=osde2e-stage-gcp-e2e-default)| <span style="color:#0ef100;">99.47%</span>|[More Detail](#osde2e-stage-gcp-e2e-default)|
|[osde2e-stage-gcp-e2e-next-z](https://prow.ci.openshift.org/?job=osde2e-stage-gcp-e2e-next-z)| <span style="color:#07f800;">99.73%</span>|[More Detail](#osde2e-stage-gcp-e2e-next-z)|



## osde2e-int-gcp-e2e-next-z

Overall pass rate: <span style="color:#15ea00;">99.20%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1476825598142713856](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-int-gcp-e2e-next-z/1476825598142713856) | 4.9.12-candidate |  | <span style="color:#15ea00;">99.20%</span>|<ul><li>[install] [Suite: operators] [OSD] OSD Metrics Exporter Basic Test Operator Upgrade should upgrade from the replaced version</li></ul>
[1476584006077124608](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-int-gcp-e2e-next-z/1476584006077124608) | 4.9.12-candidate |  | <span style="color:#15ea00;">99.20%</span>|<ul><li>[install] [Suite: e2e] Workload (guestbook) should get created in the cluster</li></ul>



## osde2e-prod-gcp-e2e-default

Overall pass rate: <span style="color:#06f900;">99.80%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1476463214043598848](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-default/1476463214043598848) | 4.9.11-candidate |  | <span style="color:#01fe00;">100.00%</span>|
[1476584010695053312](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-default/1476584010695053312) | 4.9.11-candidate |  | <span style="color:#15ea00;">99.20%</span>|<ul><li>[install] [Suite: e2e] Workload (guestbook) should get created in the cluster</li></ul>
[1476704891685572608](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-default/1476704891685572608) | 4.9.11-candidate |  | <span style="color:#01fe00;">100.00%</span>|
[1476825604442558464](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-default/1476825604442558464) | 4.9.11-candidate |  | <span style="color:#01fe00;">100.00%</span>|



## osde2e-prod-gcp-e2e-next

Overall pass rate: <span style="color:#10ef00;">99.40%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1476463215721320448](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-next/1476463215721320448) | 4.9.12-candidate |  | <span style="color:#01fe00;">100.00%</span>|
[1476584011525525504](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-next/1476584011525525504) | 4.9.12-candidate |  | <span style="color:#29d600;">98.40%</span>|<ul><li>[install] [Suite: e2e] Workload (guestbook) should get created in the cluster</li><li>[install] [Suite: operators] [OSD] RBAC Operator Operator Upgrade should upgrade from the replaced version</li></ul>
[1476704892503461888](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-next/1476704892503461888) | 4.9.12-candidate |  | <span style="color:#15ea00;">99.20%</span>|<ul><li>[install] [Suite: e2e] Workload (guestbook) should get created in the cluster</li></ul>
[1476825605273030656](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-next/1476825605273030656) | 4.9.12-candidate |  | <span style="color:#01fe00;">100.00%</span>|



## osde2e-prod-gcp-e2e-upgrade-to-latest-z

Overall pass rate: <span style="color:#1fe000;">98.80%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1476584012372774912](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-upgrade-to-latest-z/1476584012372774912) | 4.9.11-candidate | 4.9.12 | <span style="color:#1fe000;">98.80%</span>|<ul><li>[install] [Suite: e2e] Workload (guestbook) should get created in the cluster</li><li>[upgrade] [Suite: e2e] Workload (guestbook) should get created in the cluster</li><li>[upgrade] [Suite: service-definition] [OSD] DaemonSets DaemonSets are not allowed empty node-label daemonset should get created</li></ul>



## osde2e-stage-gcp-e2e-default

Overall pass rate: <span style="color:#0ef100;">99.47%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1476704908450205696](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-default/1476704908450205696) | 4.9.11-candidate |  | <span style="color:#01fe00;">100.00%</span>|
[1476825609882570752](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-default/1476825609882570752) | 4.9.11-candidate |  | <span style="color:#15ea00;">99.20%</span>|<ul><li>[install] [Suite: e2e] Workload (guestbook) should get created in the cluster</li></ul>
[1476584014457344000](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-default/1476584014457344000) | 4.9.11-candidate |  | <span style="color:#15ea00;">99.20%</span>|<ul><li>[install] [Suite: e2e] Workload (guestbook) should get created in the cluster</li></ul>



## osde2e-stage-gcp-e2e-next-z

Overall pass rate: <span style="color:#07f800;">99.73%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1476584016135065600](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-next-z/1476584016135065600) | 4.9.12-candidate |  | <span style="color:#01fe00;">100.00%</span>|
[1476704910127927296](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-next-z/1476704910127927296) | 4.9.12-candidate |  | <span style="color:#15ea00;">99.20%</span>|<ul><li>[install] [Suite: e2e] Workload (guestbook) should get created in the cluster</li></ul>
[1476825612399153152](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-next-z/1476825612399153152) | 4.9.12-candidate |  | <span style="color:#01fe00;">100.00%</span>|




