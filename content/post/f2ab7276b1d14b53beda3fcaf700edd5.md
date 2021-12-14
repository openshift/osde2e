+++
title = "OSDe2e gcp Weather Report 2021-12-14 12:02:02.340740789 +0000 UTC"
author = "OSDe2e Automation"
date = "2021-12-14 12:02:02.340740789 +0000 UTC"
tags = ["weather-report", "gcp"]
summary = "<table class=\"summary\"><tr><td bgcolor=\"#0bf400\"></td><td>int (Pass rate: 99.60)</td></tr><tr><td bgcolor=\"#0bf400\"></td><td>prod (Pass rate: 99.60)</td></tr><tr><td bgcolor=\"#0cf300\"></td><td>stage (Pass rate: 99.54)</td></tr></table>"
+++
## Summary

| Job Name | Pass Rate | More detail |
|----------|-----------|-------------|
|[osde2e-int-gcp-e2e-next-z](https://prow.ci.openshift.org/?job=osde2e-int-gcp-e2e-next-z)| <span style="color:#0bf400;">99.60%</span>|[More Detail](#osde2e-int-gcp-e2e-next-z)|
|[osde2e-prod-gcp-e2e-default](https://prow.ci.openshift.org/?job=osde2e-prod-gcp-e2e-default)| <span style="color:#01fe00;">100.00%</span>|[More Detail](#osde2e-prod-gcp-e2e-default)|
|[osde2e-prod-gcp-e2e-next](https://prow.ci.openshift.org/?job=osde2e-prod-gcp-e2e-next)| <span style="color:#0bf400;">99.60%</span>|[More Detail](#osde2e-prod-gcp-e2e-next)|
|[osde2e-prod-gcp-e2e-upgrade-to-latest-z](https://prow.ci.openshift.org/?job=osde2e-prod-gcp-e2e-upgrade-to-latest-z)| <span style="color:#1fe000;">98.80%</span>|[More Detail](#osde2e-prod-gcp-e2e-upgrade-to-latest-z)|
|[osde2e-stage-gcp-e2e-default](https://prow.ci.openshift.org/?job=osde2e-stage-gcp-e2e-default)| <span style="color:#01fe00;">100.00%</span>|[More Detail](#osde2e-stage-gcp-e2e-default)|
|[osde2e-stage-gcp-e2e-next-z](https://prow.ci.openshift.org/?job=osde2e-stage-gcp-e2e-next-z)| <span style="color:#01fe00;">100.00%</span>|[More Detail](#osde2e-stage-gcp-e2e-next-z)|
|[osde2e-stage-gcp-e2e-upgrade-to-latest](https://prow.ci.openshift.org/?job=osde2e-stage-gcp-e2e-upgrade-to-latest)| <span style="color:#0bf400;">99.60%</span>|[More Detail](#osde2e-stage-gcp-e2e-upgrade-to-latest)|
|[osde2e-stage-gcp-e2e-upgrade-to-latest-z](https://prow.ci.openshift.org/?job=osde2e-stage-gcp-e2e-upgrade-to-latest-z)| <span style="color:#1fe000;">98.80%</span>|[More Detail](#osde2e-stage-gcp-e2e-upgrade-to-latest-z)|



## osde2e-int-gcp-e2e-next-z

Overall pass rate: <span style="color:#0bf400;">99.60%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1470544170065072128](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-int-gcp-e2e-next-z/1470544170065072128) | 4.9.11-candidate |  | <span style="color:#01fe00;">100.00%</span>|
[1470665022647570432](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-int-gcp-e2e-next-z/1470665022647570432) | 4.9.11-candidate |  | <span style="color:#15ea00;">99.20%</span>|<ul><li>[install] [Suite: e2e] [OSD] RBAC Dedicated Admins SCC permissions scc-test new SCC does not break pods</li></ul>



## osde2e-prod-gcp-e2e-default

Overall pass rate: <span style="color:#01fe00;">100.00%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1470423511955673088](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-default/1470423511955673088) | 4.9.9-candidate |  | <span style="color:#01fe00;">100.00%</span>|
[1470665025986236416](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-default/1470665025986236416) | 4.9.9-candidate |  | <span style="color:#01fe00;">100.00%</span>|



## osde2e-prod-gcp-e2e-next

Overall pass rate: <span style="color:#0bf400;">99.60%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1470544177568681984](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-next/1470544177568681984) | 4.9.11-candidate |  | <span style="color:#15ea00;">99.20%</span>|<ul><li>[install] [Suite: e2e] [OSD] namespace validating webhook namespace validating webhook Non-privileged users can manage all non-privileged namespaces</li></ul>
[1470665026820902912](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-next/1470665026820902912) | 4.9.11-candidate |  | <span style="color:#01fe00;">100.00%</span>|



## osde2e-prod-gcp-e2e-upgrade-to-latest-z

Overall pass rate: <span style="color:#1fe000;">98.80%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1470302621213396992](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-upgrade-to-latest-z/1470302621213396992) | 4.9.9-candidate | 4.9.11 | <span style="color:#1fe000;">98.80%</span>|<ul><li>[install] [Suite: operators] [OSD] Splunk Forwarder Operator Operator Upgrade should upgrade from the replaced version</li><li>[upgrade] [Suite: e2e] Routes should be created for Console</li><li>[upgrade] [Suite: operators] [OSD] Splunk Forwarder Operator Operator Upgrade should upgrade from the replaced version</li></ul>



## osde2e-stage-gcp-e2e-default

Overall pass rate: <span style="color:#01fe00;">100.00%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1470423516158365696](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-default/1470423516158365696) | 4.9.9-candidate |  | <span style="color:#01fe00;">100.00%</span>|
[1470665031031984128](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-default/1470665031031984128) | 4.9.9-candidate |  | <span style="color:#01fe00;">100.00%</span>|



## osde2e-stage-gcp-e2e-next-z

Overall pass rate: <span style="color:#01fe00;">100.00%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1470544194408812544](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-next-z/1470544194408812544) | 4.9.11-candidate |  | <span style="color:#01fe00;">100.00%</span>|



## osde2e-stage-gcp-e2e-upgrade-to-latest

Overall pass rate: <span style="color:#0bf400;">99.60%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1470544195167981568](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-upgrade-to-latest/1470544195167981568) | 4.9.9-candidate | 4.9.11 | <span style="color:#0bf400;">99.60%</span>|<ul><li>[upgrade] [Suite: operators] [OSD] Splunk Forwarder Operator clusterServiceVersion openshift-splunk-forwarder-operator/splunk-forwarder-operator should be present and in succeeded state</li></ul>



## osde2e-stage-gcp-e2e-upgrade-to-latest-z

Overall pass rate: <span style="color:#1fe000;">98.80%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1470423519505420288](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-upgrade-to-latest-z/1470423519505420288) | 4.9.9-candidate | 4.9.11 | <span style="color:#1fe000;">98.80%</span>|<ul><li>[install] [Suite: e2e] Pods should be Running or Succeeded</li><li>[install] [Suite: e2e] Pods should not be Failed</li><li>[upgrade] [Suite: e2e] Cluster state should include Prometheus data</li></ul>
[1470544196006842368](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-upgrade-to-latest-z/1470544196006842368) | 4.9.9-candidate | 4.9.11 | <span style="color:#1fe000;">98.80%</span>|<ul><li>[install] [Suite: operators] [OSD] Splunk Forwarder Operator Operator Upgrade should upgrade from the replaced version</li><li>[upgrade] [Suite: e2e] [OSD] Samesite Cookie Strict Validating samesite cookie should be set for openshift-monitoring OSD managed routes</li><li>[upgrade] [Suite: operators] [OSD] Splunk Forwarder Operator Operator Upgrade should upgrade from the replaced version</li></ul>




