+++
title = "OSDe2e gcp Weather Report 2021-11-08 12:01:24.512970229 +0000 UTC"
author = "OSDe2e Automation"
date = "2021-11-08 12:01:24.512970229 +0000 UTC"
tags = ["weather-report", "gcp"]
summary = "<table class=\"summary\"><tr><td bgcolor=\"#04fb00\"></td><td>prod (Pass rate: 99.85)</td></tr><tr><td bgcolor=\"#29d600\"></td><td>stage (Pass rate: 98.40)</td></tr></table>"
+++
## Summary

| Job Name | Pass Rate | More detail |
|----------|-----------|-------------|
|[osde2e-prod-gcp-e2e-default](https://prow.ci.openshift.org/?job=osde2e-prod-gcp-e2e-default)| <span style="color:#06f900;">99.80%</span>|[More Detail](#osde2e-prod-gcp-e2e-default)|
|[osde2e-prod-gcp-e2e-next](https://prow.ci.openshift.org/?job=osde2e-prod-gcp-e2e-next)| <span style="color:#01fe00;">100.00%</span>|[More Detail](#osde2e-prod-gcp-e2e-next)|
|[osde2e-prod-gcp-e2e-upgrade-to-latest-z](https://prow.ci.openshift.org/?job=osde2e-prod-gcp-e2e-upgrade-to-latest-z)| <span style="color:#0bf400;">99.60%</span>|[More Detail](#osde2e-prod-gcp-e2e-upgrade-to-latest-z)|
|[osde2e-stage-gcp-e2e-upgrade-rescheduled](https://prow.ci.openshift.org/?job=osde2e-stage-gcp-e2e-upgrade-rescheduled)| <span style="color:#29d600;">98.40%</span>|[More Detail](#osde2e-stage-gcp-e2e-upgrade-rescheduled)|
|[osde2e-stage-gcp-e2e-upgrade-to-latest](https://prow.ci.openshift.org/?job=osde2e-stage-gcp-e2e-upgrade-to-latest)| <span style="color:#29d600;">98.40%</span>|[More Detail](#osde2e-stage-gcp-e2e-upgrade-to-latest)|



## osde2e-prod-gcp-e2e-default

Overall pass rate: <span style="color:#06f900;">99.80%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1457256598501068800](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-default/1457256598501068800) | 4.9.4-candidate |  | <span style="color:#01fe00;">100.00%</span>|
[1457377395576147968](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-default/1457377395576147968) | 4.9.4-candidate |  | <span style="color:#15ea00;">99.20%</span>|<ul><li>[install] [Suite: e2e] [OSD] RBAC Dedicated Admins SCC permissions scc-test new SCC does not break pods</li></ul>
[1457498271663525888](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-default/1457498271663525888) | 4.9.4-candidate |  | <span style="color:#01fe00;">100.00%</span>|
[1457618989713723392](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-default/1457618989713723392) | 4.9.4-candidate |  | <span style="color:#01fe00;">100.00%</span>|



## osde2e-prod-gcp-e2e-next

Overall pass rate: <span style="color:#01fe00;">100.00%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1457256599323152384](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-next/1457256599323152384) | 4.9.6-candidate |  | <span style="color:#01fe00;">100.00%</span>|
[1457498272493998080](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-next/1457498272493998080) | 4.9.6-candidate |  | <span style="color:#01fe00;">100.00%</span>|
[1457618990389006336](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-next/1457618990389006336) | 4.9.6-candidate |  | <span style="color:#01fe00;">100.00%</span>|



## osde2e-prod-gcp-e2e-upgrade-to-latest-z

Overall pass rate: <span style="color:#0bf400;">99.60%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1457256600157818880](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-upgrade-to-latest-z/1457256600157818880) | 4.9.4-candidate | 4.9.6 | <span style="color:#0bf400;">99.60%</span>|<ul><li>[upgrade] [Suite: e2e] Storage storage create PVCs</li></ul>



## osde2e-stage-gcp-e2e-upgrade-rescheduled

Overall pass rate: <span style="color:#29d600;">98.40%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1457317019501203456](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-upgrade-rescheduled/1457317019501203456) | 4.9.4-candidate | 4.9.6 | <span style="color:#29d600;">98.40%</span>|<ul><li>[upgrade] [Suite: operators] CloudIngressOperator deployment should have all desired replicas ready</li><li>[upgrade] [Suite: operators] CloudIngressOperator rh-api-test cidr block changes should updated the service</li></ul>



## osde2e-stage-gcp-e2e-upgrade-to-latest

Overall pass rate: <span style="color:#29d600;">98.40%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1457256606872899584](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-upgrade-to-latest/1457256606872899584) | 4.9.4-candidate | 4.9.6 | <span style="color:#29d600;">98.40%</span>|<ul><li>[upgrade] [Suite: operators] CloudIngressOperator deployment should have all desired replicas ready</li><li>[upgrade] [Suite: operators] CloudIngressOperator rh-api-test cidr block changes should updated the service</li></ul>




