+++
title = "OSDe2e aws Weather Report 2021-07-06 12:00:32.171778588 +0000 UTC"
author = "OSDe2e Automation"
date = "2021-07-06 12:00:32.171778588 +0000 UTC"
tags = ["weather-report", "aws"]
summary = "<table class=\"summary\"><tr><td bgcolor=\"#0ef100\"></td><td>prod (Pass rate: 99.46)</td></tr></table>"
+++
## Summary

| Job Name | Pass Rate | More detail |
|----------|-----------|-------------|
|[osde2e-prod-aws-e2e-default](https://prow.ci.openshift.org/?job=osde2e-prod-aws-e2e-default)| <span style="color:#01fe00;">100.00%</span>|[More Detail](#osde2e-prod-aws-e2e-default)|
|[osde2e-prod-aws-e2e-middle-imageset](https://prow.ci.openshift.org/?job=osde2e-prod-aws-e2e-middle-imageset)| <span style="color:#10ef00;">99.38%</span>|[More Detail](#osde2e-prod-aws-e2e-middle-imageset)|
|[osde2e-prod-aws-e2e-next](https://prow.ci.openshift.org/?job=osde2e-prod-aws-e2e-next)| <span style="color:#21de00;">98.73%</span>|[More Detail](#osde2e-prod-aws-e2e-next)|
|[osde2e-prod-aws-e2e-upgrade-prod-minus-one-to-next](https://prow.ci.openshift.org/?job=osde2e-prod-aws-e2e-upgrade-prod-minus-one-to-next)| <span style="color:#10ef00;">99.38%</span>|[More Detail](#osde2e-prod-aws-e2e-upgrade-prod-minus-one-to-next)|
|[osde2e-prod-aws-e2e-upgrade-prod-minus-three-to-next](https://prow.ci.openshift.org/?job=osde2e-prod-aws-e2e-upgrade-prod-minus-three-to-next)| <span style="color:#10ef00;">99.37%</span>|[More Detail](#osde2e-prod-aws-e2e-upgrade-prod-minus-three-to-next)|
|[osde2e-prod-aws-e2e-upgrade-prod-minus-two-to-next](https://prow.ci.openshift.org/?job=osde2e-prod-aws-e2e-upgrade-prod-minus-two-to-next)| <span style="color:#10ef00;">99.38%</span>|[More Detail](#osde2e-prod-aws-e2e-upgrade-prod-minus-two-to-next)|
|[osde2e-prod-aws-e2e-upgrade-rescheduled](https://prow.ci.openshift.org/?job=osde2e-prod-aws-e2e-upgrade-rescheduled)| <span style="color:#08f700;">99.69%</span>|[More Detail](#osde2e-prod-aws-e2e-upgrade-rescheduled)|
|[osde2e-prod-aws-e2e-upgrade-to-latest](https://prow.ci.openshift.org/?job=osde2e-prod-aws-e2e-upgrade-to-latest)| <span style="color:#0ef100;">99.48%</span>|[More Detail](#osde2e-prod-aws-e2e-upgrade-to-latest)|



## osde2e-prod-aws-e2e-default

Overall pass rate: <span style="color:#01fe00;">100.00%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1412320612520562688](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-aws-e2e-default/1412320612520562688) | 4.7.18-candidate |  | <span style="color:#01fe00;">100.00%</span>|
[1412018661387931648](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-aws-e2e-default/1412018661387931648) | 4.7.18-candidate |  | <span style="color:#01fe00;">100.00%</span>|
[1412199880654327808](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-aws-e2e-default/1412199880654327808) | 4.7.18-candidate |  | <span style="color:#01fe00;">100.00%</span>|



## osde2e-prod-aws-e2e-middle-imageset

Overall pass rate: <span style="color:#10ef00;">99.38%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1412018662184849408](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-aws-e2e-middle-imageset/1412018662184849408) | 4.6.34-candidate |  | <span style="color:#10ef00;">99.38%</span>|<ul><li>[install] [Suite: operators] [OSD] Configure AlertManager Operator roles with prefix should exist</li></ul>



## osde2e-prod-aws-e2e-next

Overall pass rate: <span style="color:#21de00;">98.73%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1412199881476411392](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-aws-e2e-next/1412199881476411392) | 4.8.0-rc.3-candidate |  | <span style="color:#11ee00;">99.37%</span>|<ul><li>[install] [Suite: operators] [OSD] Splunk Forwarder Operator Operator Upgrade should upgrade from the replaced version</li></ul>
[1412320613367812096](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-aws-e2e-next/1412320613367812096) | 4.8.0-rc.3-candidate |  | <span style="color:#31ce00;">98.10%</span>|<ul><li>[install] [Suite: e2e] [OSD] namespace validating webhook namespace validating webhook Members of SRE groups can manage all namespaces</li><li>[install] [Suite: operators] [OSD] Custom Domains Operator Should allow dedicated-admins to create domains Should be resolvable by external services</li><li>[install] [Suite: operators] [OSD] Splunk Forwarder Operator Operator Upgrade should upgrade from the replaced version</li></ul>



## osde2e-prod-aws-e2e-upgrade-prod-minus-one-to-next

Overall pass rate: <span style="color:#10ef00;">99.38%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1412229994087714816](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-aws-e2e-upgrade-prod-minus-one-to-next/1412229994087714816) | 4.7.18-candidate | 4.7.19 | <span style="color:#10ef00;">99.38%</span>|<ul><li>[install] [Suite: operators] [OSD] Custom Domains Operator Should allow dedicated-admins to create domains Should be resolvable by external services</li><li>[upgrade] [Suite: service-definition] [OSD] NodeLabels Modifying nodeLabels is not allowed node-label cannot be added</li></ul>



## osde2e-prod-aws-e2e-upgrade-prod-minus-three-to-next

Overall pass rate: <span style="color:#10ef00;">99.37%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1411988448704729088](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-aws-e2e-upgrade-prod-minus-three-to-next/1411988448704729088) | 4.7.16-candidate | 4.8.0-rc.0 | <span style="color:#10ef00;">99.37%</span>|<ul><li>[upgrade] [Suite: operators] CloudIngressOperator apischeme apischemes CR instance must be present on cluster</li><li>[upgrade] [Suite: operators] [OSD] Splunk Forwarder Operator Operator Upgrade should upgrade from the replaced version</li></ul>



## osde2e-prod-aws-e2e-upgrade-prod-minus-two-to-next

Overall pass rate: <span style="color:#10ef00;">99.38%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1411927851225059328](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-aws-e2e-upgrade-prod-minus-two-to-next/1411927851225059328) | 4.7.17-candidate | 4.7.19 | <span style="color:#18e700;">99.07%</span>|<ul><li>[install] [Suite: e2e] [OSD] namespace validating webhook namespace validating webhook Non-privileged users can manage all non-privileged namespaces</li><li>[upgrade] [Suite: informing] CloudIngressOperator rh-ssh-test cidr block changes should updated the service</li><li>[upgrade] [Suite: operators] [OSD] Custom Domains Operator Should allow dedicated-admins to create domains Should be resolvable by external services</li></ul>
[1412290396620328960](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-aws-e2e-upgrade-prod-minus-two-to-next/1412290396620328960) | 4.7.17-candidate | 4.7.19 | <span style="color:#08f700;">99.69%</span>|<ul><li>[upgrade] [Suite: informing] [OSD] hive ownership validating webhook hiveownership validating webhook dedicated admins cannot delete managed CRQs</li></ul>



## osde2e-prod-aws-e2e-upgrade-rescheduled

Overall pass rate: <span style="color:#08f700;">99.69%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1412229994125463552](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-aws-e2e-upgrade-rescheduled/1412229994125463552) | 4.7.18-candidate | 4.7.19 | <span style="color:#08f700;">99.69%</span>|<ul><li>[upgrade] [Suite: service-definition] [OSD] NodeLabels Modifying nodeLabels is not allowed node-label cannot be added</li></ul>



## osde2e-prod-aws-e2e-upgrade-to-latest

Overall pass rate: <span style="color:#0ef100;">99.48%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1412260193382699008](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-aws-e2e-upgrade-to-latest/1412260193382699008) | 4.7.18-candidate | 4.7.19 | <span style="color:#08f700;">99.69%</span>|<ul><li>[upgrade] [Suite: openshift][disruptive] should run until completion</li></ul>
[1411958070690451456](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-aws-e2e-upgrade-to-latest/1411958070690451456) | 4.7.18-candidate | 4.7.19 | <span style="color:#10ef00;">99.38%</span>|<ul><li>[install] [Suite: operators] CloudIngressOperator rh-api-test cidr block changes should updated the service</li><li>[upgrade] [Suite: e2e] Workload (guestbook) should get created in the cluster</li></ul>
[1412018664705626112](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-aws-e2e-upgrade-to-latest/1412018664705626112) | 4.7.18-candidate | 4.7.19 | <span style="color:#10ef00;">99.38%</span>|<ul><li>[upgrade] [Suite: operators] [OSD] Custom Domains Operator Should allow dedicated-admins to create domains Should be resolvable by external services</li><li>[upgrade] [Suite: operators] [OSD] Prune jobs pruner jobs should works builds-pruner should run successfully</li></ul>




