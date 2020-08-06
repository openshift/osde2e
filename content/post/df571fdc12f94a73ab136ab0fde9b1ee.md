+++
title = "OSDe2e moa Weather Report 2020-08-06 12:00:54.528780808 +0000 UTC"
author = "OSDe2e Automation"
date = "2020-08-06 12:00:54.528780808 +0000 UTC"
tags = ["weather-report", "moa"]
summary = "<table class=\"summary\"><tr><td bgcolor=\"#ff0000\"></td><td>int (Pass rate: 33.85)</td></tr><tr><td bgcolor=\"#0cf300\"></td><td>prod (Pass rate: 99.55)</td></tr><tr><td bgcolor=\"#ff0000\"></td><td>stage (Pass rate: 61.82)</td></tr></table>"
+++
## Summary

| Job Name | Pass Rate | More detail |
|----------|-----------|-------------|
|[osde2e-int-moa-e2e-osd-default-nightly](https://prow.svc.ci.openshift.org/?job=osde2e-int-moa-e2e-osd-default-nightly)| <span style="color:#2fd000;">98.18%</span>|[More Detail](#osde2e-int-moa-e2e-osd-default-nightly)|
|[osde2e-int-moa-e2e-osd-default-plus-one-nightly](https://prow.svc.ci.openshift.org/?job=osde2e-int-moa-e2e-osd-default-plus-one-nightly)| <span style="color:#ff0000;">48.18%</span>|[More Detail](#osde2e-int-moa-e2e-osd-default-plus-one-nightly)|
|[osde2e-int-moa-e2e-osd-default-plus-two-nightly](https://prow.svc.ci.openshift.org/?job=osde2e-int-moa-e2e-osd-default-plus-two-nightly)| <span style="color:#ff0000;">0.00%</span>|[More Detail](#osde2e-int-moa-e2e-osd-default-plus-two-nightly)|
|[osde2e-int-moa-e2e-upgrade-to-osd-default-nightly](https://prow.svc.ci.openshift.org/?job=osde2e-int-moa-e2e-upgrade-to-osd-default-nightly)| <span style="color:#ff0000;">32.73%</span>|[More Detail](#osde2e-int-moa-e2e-upgrade-to-osd-default-nightly)|
|[osde2e-int-moa-e2e-upgrade-to-osd-default-plus-one-nightly](https://prow.svc.ci.openshift.org/?job=osde2e-int-moa-e2e-upgrade-to-osd-default-plus-one-nightly)| <span style="color:#ff0000;">16.36%</span>|[More Detail](#osde2e-int-moa-e2e-upgrade-to-osd-default-plus-one-nightly)|
|[osde2e-prod-moa-e2e-default](https://prow.svc.ci.openshift.org/?job=osde2e-prod-moa-e2e-default)| <span style="color:#18e700;">99.09%</span>|[More Detail](#osde2e-prod-moa-e2e-default)|
|[osde2e-prod-moa-e2e-next](https://prow.svc.ci.openshift.org/?job=osde2e-prod-moa-e2e-next)| <span style="color:#01fe00;">100.00%</span>|[More Detail](#osde2e-prod-moa-e2e-next)|
|[osde2e-stage-moa-e2e-default](https://prow.svc.ci.openshift.org/?job=osde2e-stage-moa-e2e-default)| <span style="color:#01fe00;">100.00%</span>|[More Detail](#osde2e-stage-moa-e2e-default)|
|[osde2e-stage-moa-e2e-next](https://prow.svc.ci.openshift.org/?job=osde2e-stage-moa-e2e-next)| <span style="color:#ff0000;">32.12%</span>|[More Detail](#osde2e-stage-moa-e2e-next)|
|[osde2e-stage-moa-e2e-upgrade-default-next](https://prow.svc.ci.openshift.org/?job=osde2e-stage-moa-e2e-upgrade-default-next)| <span style="color:#ff0000;">66.06%</span>|[More Detail](#osde2e-stage-moa-e2e-upgrade-default-next)|



## osde2e-int-moa-e2e-osd-default-nightly

Overall pass rate: <span style="color:#2fd000;">98.18%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1291162139393789952](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-int-moa-e2e-osd-default-nightly/1291162139393789952) | 4.4.0-0.nightly-2020-08-05-165415 |  | <span style="color:#2fd000;">98.18%</span>|<ul><li>[install] [Suite: e2e] Cluster state should have no alerts</li></ul>
[1291282927174291456](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-int-moa-e2e-osd-default-nightly/1291282927174291456) | 4.4.0-0.nightly-2020-08-05-165415 |  | <span style="color:#2fd000;">98.18%</span>|<ul><li>[install] [Suite: e2e] Cluster state should have no alerts</li></ul>



## osde2e-int-moa-e2e-osd-default-plus-one-nightly

Overall pass rate: <span style="color:#ff0000;">48.18%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1291282928868790272](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-int-moa-e2e-osd-default-plus-one-nightly/1291282928868790272) | 4.5.0-0.nightly-2020-08-05-164433 |  | <span style="color:#ff0000;">0.00%</span>|<ul><li>[install] BeforeSuite</li></ul>
[1291162141075705856](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-int-moa-e2e-osd-default-plus-one-nightly/1291162141075705856) | 4.5.0-0.nightly-2020-08-05-164433 |  | <span style="color:#5da200;">96.36%</span>|<ul><li>[install] [Suite: e2e] Cluster state should have no alerts</li><li>[install] [Suite: service-definition] [OSD] DaemonSets dedicated-admin group permissions cannot add members to cluster-admin</li></ul>



## osde2e-int-moa-e2e-osd-default-plus-two-nightly

Overall pass rate: <span style="color:#ff0000;">0.00%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1291282930542317568](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-int-moa-e2e-osd-default-plus-two-nightly/1291282930542317568) | 4.6.0-0.nightly-2020-08-05-174122 |  | <span style="color:#ff0000;">0.00%</span>|<ul><li>[install] BeforeSuite</li></ul>
[1290799899289325568](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-int-moa-e2e-osd-default-plus-two-nightly/1290799899289325568) | 4.6.0-0.nightly-2020-08-04-210224 |  | <span style="color:#ff0000;">0.00%</span>|<ul><li>[install] BeforeSuite</li></ul>
[1291162142757621760](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-int-moa-e2e-osd-default-plus-two-nightly/1291162142757621760) | 4.6.0-0.nightly-2020-08-05-174122 |  | <span style="color:#ff0000;">0.00%</span>|<ul><li>[install] BeforeSuite</li></ul>



## osde2e-int-moa-e2e-upgrade-to-osd-default-nightly

Overall pass rate: <span style="color:#ff0000;">32.73%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1291162140257816576](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-int-moa-e2e-upgrade-to-osd-default-nightly/1291162140257816576) | 4.4.11 | 4.4.0-0.nightly-2020-08-05-165415 | <span style="color:#ff0000;">49.09%</span>|<ul><li>[install] [Suite: e2e] Cluster state should have no alerts</li><li>[upgrade] BeforeSuite</li></ul>
[1291282928013152256](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-int-moa-e2e-upgrade-to-osd-default-nightly/1291282928013152256) | 4.4.11 | 4.4.0-0.nightly-2020-08-05-220634 | <span style="color:#ff0000;">49.09%</span>|<ul><li>[install] [Suite: e2e] Cluster state should have no alerts</li><li>[upgrade] BeforeSuite</li></ul>
[1290920682028273664](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-int-moa-e2e-upgrade-to-osd-default-nightly/1290920682028273664) | 4.4.11 | 4.4.0-0.nightly-2020-08-03-123644 | <span style="color:#ff0000;">0.00%</span>|<ul><li>[install] BeforeSuite</li></ul>



## osde2e-int-moa-e2e-upgrade-to-osd-default-plus-one-nightly

Overall pass rate: <span style="color:#ff0000;">16.36%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1291162141922955264](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-int-moa-e2e-upgrade-to-osd-default-plus-one-nightly/1291162141922955264) | 4.4.11 | 4.5.0-0.nightly-2020-08-05-164433 | <span style="color:#ff0000;">49.09%</span>|<ul><li>[install] [Suite: e2e] Cluster state should have no alerts</li><li>[upgrade] BeforeSuite</li></ul>
[1291282929695068160](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-int-moa-e2e-upgrade-to-osd-default-plus-one-nightly/1291282929695068160) | 4.4.11 | 4.5.0-0.nightly-2020-08-06-050650 | <span style="color:#ff0000;">0.00%</span>|<ul><li>[install] BeforeSuite</li></ul>
[1290799898433687552](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-int-moa-e2e-upgrade-to-osd-default-plus-one-nightly/1290799898433687552) | 4.4.11 | 4.5.0-0.nightly-2020-08-03-123303 | <span style="color:#ff0000;">0.00%</span>|<ul><li>[install] BeforeSuite</li></ul>



## osde2e-prod-moa-e2e-default

Overall pass rate: <span style="color:#18e700;">99.09%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1291041344482971648](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-moa-e2e-default/1291041344482971648) | 4.4.11 |  | <span style="color:#2fd000;">98.18%</span>|<ul><li>[Log Metrics] cluster-mgmt-500</li><li>[install] [Suite: e2e] Cluster state should have no alerts</li></ul>
[1291282940575092736](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-moa-e2e-default/1291282940575092736) | 4.4.11 |  | <span style="color:#01fe00;">100.00%</span>|



## osde2e-prod-moa-e2e-next

Overall pass rate: <span style="color:#01fe00;">100.00%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1291162152849117184](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-moa-e2e-next/1291162152849117184) | 4.4.11 |  | <span style="color:#01fe00;">100.00%</span>|
[1291282942248620032](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-moa-e2e-next/1291282942248620032) | 4.4.11 |  | <span style="color:#01fe00;">100.00%</span>|



## osde2e-stage-moa-e2e-default

Overall pass rate: <span style="color:#01fe00;">100.00%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1290920688684634112](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-moa-e2e-default/1290920688684634112) | 4.4.11 |  | <span style="color:#01fe00;">100.00%</span>|
[1291282934698872832](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-moa-e2e-default/1291282934698872832) | 4.4.11 |  | <span style="color:#01fe00;">100.00%</span>|



## osde2e-stage-moa-e2e-next

Overall pass rate: <span style="color:#ff0000;">32.12%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1291162147786592256](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-moa-e2e-next/1291162147786592256) | 4.5.5 |  | <span style="color:#5da200;">96.36%</span>|<ul><li>[install] [Suite: e2e] Cluster state should have no alerts</li><li>[install] [Suite: service-definition] [OSD] DaemonSets dedicated-admin group permissions cannot add members to cluster-admin</li></ul>
[1291282936410148864](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-moa-e2e-next/1291282936410148864) | 4.5.5 |  | <span style="color:#ff0000;">0.00%</span>|<ul><li>[install] BeforeSuite</li></ul>
[1290799904309907456](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-moa-e2e-next/1290799904309907456) | 4.5.4 |  | <span style="color:#ff0000;">0.00%</span>|<ul><li>[install] BeforeSuite</li></ul>



## osde2e-stage-moa-e2e-upgrade-default-next

Overall pass rate: <span style="color:#ff0000;">66.06%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1290799903462658048](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-moa-e2e-upgrade-default-next/1290799903462658048) | 4.4.11 | 4.4.15 | <span style="color:#01fe00;">100.00%</span>|
[1291162146964508672](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-moa-e2e-upgrade-default-next/1291162146964508672) | 4.4.11 | 4.4.15 | <span style="color:#ff0000;">49.09%</span>|<ul><li>[install] [Suite: e2e] Cluster state should have no alerts</li><li>[upgrade] BeforeSuite</li></ul>
[1291282935541927936](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-moa-e2e-upgrade-default-next/1291282935541927936) | 4.4.11 | 4.4.15 | <span style="color:#ff0000;">49.09%</span>|<ul><li>[install] [Suite: e2e] Cluster state should have no alerts</li><li>[upgrade] BeforeSuite</li></ul>



