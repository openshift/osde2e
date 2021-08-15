+++
title = "OSDe2e gcp Weather Report 2021-08-15 12:00:33.389840101 +0000 UTC"
author = "OSDe2e Automation"
date = "2021-08-15 12:00:33.389840101 +0000 UTC"
tags = ["weather-report", "gcp"]
summary = "<table class=\"summary\"><tr><td bgcolor=\"#01fe00\"></td><td>int (Pass rate: 100.00)</td></tr><tr><td bgcolor=\"#0ef100\"></td><td>prod (Pass rate: 99.48)</td></tr><tr><td bgcolor=\"#08f700\"></td><td>stage (Pass rate: 99.69)</td></tr></table>"
+++
## Summary

| Job Name | Pass Rate | More detail |
|----------|-----------|-------------|
|[osde2e-int-gcp-e2e-next-z](https://prow.ci.openshift.org/?job=osde2e-int-gcp-e2e-next-z)| <span style="color:#01fe00;">100.00%</span>|[More Detail](#osde2e-int-gcp-e2e-next-z)|
|[osde2e-prod-gcp-e2e-default](https://prow.ci.openshift.org/?job=osde2e-prod-gcp-e2e-default)| <span style="color:#10ef00;">99.38%</span>|[More Detail](#osde2e-prod-gcp-e2e-default)|
|[osde2e-prod-gcp-e2e-next](https://prow.ci.openshift.org/?job=osde2e-prod-gcp-e2e-next)| <span style="color:#0bf400;">99.59%</span>|[More Detail](#osde2e-prod-gcp-e2e-next)|
|[osde2e-prod-gcp-e2e-upgrade-to-latest-z](https://prow.ci.openshift.org/?job=osde2e-prod-gcp-e2e-upgrade-to-latest-z)| <span style="color:#10ef00;">99.38%</span>|[More Detail](#osde2e-prod-gcp-e2e-upgrade-to-latest-z)|
|[osde2e-stage-gcp-e2e-default](https://prow.ci.openshift.org/?job=osde2e-stage-gcp-e2e-default)| <span style="color:#01fe00;">100.00%</span>|[More Detail](#osde2e-stage-gcp-e2e-default)|
|[osde2e-stage-gcp-e2e-upgrade-rescheduled](https://prow.ci.openshift.org/?job=osde2e-stage-gcp-e2e-upgrade-rescheduled)| <span style="color:#10ef00;">99.38%</span>|[More Detail](#osde2e-stage-gcp-e2e-upgrade-rescheduled)|



## osde2e-int-gcp-e2e-next-z

Overall pass rate: <span style="color:#01fe00;">100.00%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1426695356128694272](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-int-gcp-e2e-next-z/1426695356128694272) | 4.8.3-candidate |  | <span style="color:#01fe00;">100.00%</span>|



## osde2e-prod-gcp-e2e-default

Overall pass rate: <span style="color:#10ef00;">99.38%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1426574507807608832](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-default/1426574507807608832) | 4.8.3-candidate |  | <span style="color:#10ef00;">99.38%</span>|<ul><li>[install] [Suite: operators] [OSD] RBAC Operator Operator Upgrade should upgrade from the replaced version</li></ul>
[1426816103492882432](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-default/1426816103492882432) | 4.8.3-candidate |  | <span style="color:#10ef00;">99.38%</span>|<ul><li>[install] [Suite: operators] [OSD] RBAC Operator Operator Upgrade should upgrade from the replaced version</li></ul>



## osde2e-prod-gcp-e2e-next

Overall pass rate: <span style="color:#0bf400;">99.59%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1426574508646469632](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-next/1426574508646469632) | 4.8.5-candidate |  | <span style="color:#01fe00;">100.00%</span>|
[1426695362822803456](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-next/1426695362822803456) | 4.8.4-candidate |  | <span style="color:#10ef00;">99.38%</span>|<ul><li>[install] [Suite: operators] [OSD] RBAC Operator Operator Upgrade should upgrade from the replaced version</li></ul>
[1426453811123195904](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-next/1426453811123195904) | 4.8.4-candidate |  | <span style="color:#10ef00;">99.38%</span>|<ul><li>[install] [Suite: operators] [OSD] RBAC Operator Operator Upgrade should upgrade from the replaced version</li></ul>



## osde2e-prod-gcp-e2e-upgrade-to-latest-z

Overall pass rate: <span style="color:#10ef00;">99.38%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1426695363678441472](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-upgrade-to-latest-z/1426695363678441472) | 4.8.3-candidate | 4.8.4 | <span style="color:#10ef00;">99.38%</span>|<ul><li>[upgrade] [Suite: e2e] Encrypted Storage in GCP clusters can be created by dedicated admins</li><li>[upgrade] [Suite: operators] [OSD] Certman Operator certificate secret should be applied when cluster installed certificate secret exist under openshift-config namespace</li></ul>



## osde2e-stage-gcp-e2e-default

Overall pass rate: <span style="color:#01fe00;">100.00%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1426453817829888000](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-default/1426453817829888000) | 4.8.3-candidate |  | <span style="color:#01fe00;">100.00%</span>|



## osde2e-stage-gcp-e2e-upgrade-rescheduled

Overall pass rate: <span style="color:#10ef00;">99.38%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1426514114850590720](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-upgrade-rescheduled/1426514114850590720) | 4.8.3-candidate | 4.8.5 | <span style="color:#10ef00;">99.38%</span>|<ul><li>[upgrade] [Suite: e2e] Encrypted Storage in GCP clusters can be created by dedicated admins</li><li>[upgrade] [Suite: e2e] [OSD] OCM Metrics do exist and are not empty</li></ul>




