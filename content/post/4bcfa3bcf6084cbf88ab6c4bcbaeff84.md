+++
title = "OSDe2e gcp Weather Report 2021-12-25 12:00:44.22618426 +0000 UTC"
author = "OSDe2e Automation"
date = "2021-12-25 12:00:44.22618426 +0000 UTC"
tags = ["weather-report", "gcp"]
summary = "<table class=\"summary\"><tr><td bgcolor=\"#01fe00\"></td><td>int (Pass rate: 100.00)</td></tr><tr><td bgcolor=\"#1fe000\"></td><td>prod (Pass rate: 98.80)</td></tr><tr><td bgcolor=\"#1ce300\"></td><td>stage (Pass rate: 98.94)</td></tr></table>"
+++
## Summary

| Job Name | Pass Rate | More detail |
|----------|-----------|-------------|
|[osde2e-int-gcp-e2e-next-z](https://prow.ci.openshift.org/?job=osde2e-int-gcp-e2e-next-z)| <span style="color:#01fe00;">100.00%</span>|[More Detail](#osde2e-int-gcp-e2e-next-z)|
|[osde2e-prod-gcp-e2e-upgrade-to-latest-z](https://prow.ci.openshift.org/?job=osde2e-prod-gcp-e2e-upgrade-to-latest-z)| <span style="color:#1fe000;">98.80%</span>|[More Detail](#osde2e-prod-gcp-e2e-upgrade-to-latest-z)|
|[osde2e-stage-gcp-e2e-default](https://prow.ci.openshift.org/?job=osde2e-stage-gcp-e2e-default)| <span style="color:#3ec100;">97.60%</span>|[More Detail](#osde2e-stage-gcp-e2e-default)|
|[osde2e-stage-gcp-e2e-next-z](https://prow.ci.openshift.org/?job=osde2e-stage-gcp-e2e-next-z)| <span style="color:#01fe00;">100.00%</span>|[More Detail](#osde2e-stage-gcp-e2e-next-z)|
|[osde2e-stage-gcp-e2e-upgrade-to-latest](https://prow.ci.openshift.org/?job=osde2e-stage-gcp-e2e-upgrade-to-latest)| <span style="color:#1fe000;">98.80%</span>|[More Detail](#osde2e-stage-gcp-e2e-upgrade-to-latest)|
|[osde2e-stage-gcp-e2e-upgrade-to-latest-z](https://prow.ci.openshift.org/?job=osde2e-stage-gcp-e2e-upgrade-to-latest-z)| <span style="color:#1fe000;">98.80%</span>|[More Detail](#osde2e-stage-gcp-e2e-upgrade-to-latest-z)|



## osde2e-int-gcp-e2e-next-z

Overall pass rate: <span style="color:#01fe00;">100.00%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1474288853782106112](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-int-gcp-e2e-next-z/1474288853782106112) | 4.9.12-candidate |  | <span style="color:#01fe00;">100.00%</span>|



## osde2e-prod-gcp-e2e-upgrade-to-latest-z

Overall pass rate: <span style="color:#1fe000;">98.80%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1474530517515767808](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-upgrade-to-latest-z/1474530517515767808) | 4.9.11-candidate | 4.9.12 | <span style="color:#1fe000;">98.80%</span>|<ul><li>[install] [Suite: e2e] Workload (guestbook) should get created in the cluster</li><li>[upgrade] [Suite: e2e] Workload (guestbook) should get created in the cluster</li><li>[upgrade] [Suite: e2e] [OSD] namespace validating webhook namespace validating webhook dedicated admins cannot manage privileged namespaces</li></ul>



## osde2e-stage-gcp-e2e-default

Overall pass rate: <span style="color:#3ec100;">97.60%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1474409660411809792](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-default/1474409660411809792) | 4.9.11-candidate |  | <span style="color:#3ec100;">97.60%</span>|<ul><li>[install] [Suite: e2e] Pods should be Running or Succeeded</li><li>[install] [Suite: e2e] Pods should not be Failed</li><li>[install] [Suite: e2e] Workload (guestbook) should get created in the cluster</li></ul>



## osde2e-stage-gcp-e2e-next-z

Overall pass rate: <span style="color:#01fe00;">100.00%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1474288866780254208](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-next-z/1474288866780254208) | 4.9.12-candidate |  | <span style="color:#01fe00;">100.00%</span>|
[1474409662081142784](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-next-z/1474409662081142784) | 4.9.12-candidate |  | <span style="color:#01fe00;">100.00%</span>|



## osde2e-stage-gcp-e2e-upgrade-to-latest

Overall pass rate: <span style="color:#1fe000;">98.80%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1474288867614920704](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-upgrade-to-latest/1474288867614920704) | 4.9.11-candidate | 4.9.12 | <span style="color:#1fe000;">98.80%</span>|<ul><li>[install] [Suite: e2e] Workload (guestbook) should get created in the cluster</li><li>[upgrade] [Suite: e2e] Workload (guestbook) should get created in the cluster</li><li>[upgrade] [Suite: operators] [OSD] RBAC Dedicated Admins SubjectPermission SubjectPermission should have the expected ClusterRoles, ClusterRoleBindings and RoleBindinsg</li></ul>
[1474409662534127616](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-upgrade-to-latest/1474409662534127616) | 4.9.11-candidate | 4.9.12 | <span style="color:#1fe000;">98.80%</span>|<ul><li>[install] [Suite: e2e] Workload (guestbook) should get created in the cluster</li><li>[upgrade] [Suite: e2e] Workload (guestbook) should get created in the cluster</li><li>[upgrade] [Suite: operators] CloudIngressOperator apischeme apischemes CR instance must be present on cluster</li></ul>
[1474530535127650304](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-upgrade-to-latest/1474530535127650304) | 4.9.11-candidate | 4.9.12 | <span style="color:#1fe000;">98.80%</span>|<ul><li>[install] [Suite: e2e] Workload (guestbook) should get created in the cluster</li><li>[upgrade] [Suite: e2e] Workload (guestbook) should get created in the cluster</li><li>[upgrade] [Suite: operators] CloudIngressOperator rh-api-test hostname should resolve</li></ul>



## osde2e-stage-gcp-e2e-upgrade-to-latest-z

Overall pass rate: <span style="color:#1fe000;">98.80%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1474288868453781504](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-upgrade-to-latest-z/1474288868453781504) | 4.9.11-candidate | 4.9.12 | <span style="color:#1fe000;">98.80%</span>|<ul><li>[install] [Suite: e2e] Workload (guestbook) should get created in the cluster</li><li>[upgrade] [Suite: e2e] Storage storage create PVCs</li><li>[upgrade] [Suite: e2e] Workload (guestbook) should get created in the cluster</li></ul>
[1474409663335239680](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-upgrade-to-latest-z/1474409663335239680) | 4.9.11-candidate | 4.9.12 | <span style="color:#1fe000;">98.80%</span>|<ul><li>[install] [Suite: e2e] Workload (guestbook) should get created in the cluster</li><li>[upgrade] [Suite: e2e] MachineHealthChecks infra MHC should exist</li><li>[upgrade] [Suite: e2e] Workload (guestbook) should get created in the cluster</li></ul>
[1474530535974899712](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-upgrade-to-latest-z/1474530535974899712) | 4.9.11-candidate | 4.9.12 | <span style="color:#1fe000;">98.80%</span>|<ul><li>[install] [Suite: e2e] Workload (guestbook) should get created in the cluster</li><li>[upgrade] [Suite: e2e] Routes should be created for Console</li><li>[upgrade] [Suite: e2e] Workload (guestbook) should get created in the cluster</li></ul>




