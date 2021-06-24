+++
title = "OSDe2e aws Weather Report 2021-06-24 12:00:33.237734968 +0000 UTC"
author = "OSDe2e Automation"
date = "2021-06-24 12:00:33.237734968 +0000 UTC"
tags = ["weather-report", "aws"]
summary = "<table class=\"summary\"><tr><td bgcolor=\"#0bf400\"></td><td>prod (Pass rate: 99.58)</td></tr><tr><td bgcolor=\"#04fb00\"></td><td>stage (Pass rate: 99.84)</td></tr></table>"
+++
## Summary

| Job Name | Pass Rate | More detail |
|----------|-----------|-------------|
|[osde2e-prod-aws-e2e-default](https://prow.ci.openshift.org/?job=osde2e-prod-aws-e2e-default)| <span style="color:#04fb00;">99.88%</span>|[More Detail](#osde2e-prod-aws-e2e-default)|
|[osde2e-prod-aws-e2e-next](https://prow.ci.openshift.org/?job=osde2e-prod-aws-e2e-next)| <span style="color:#09f600;">99.68%</span>|[More Detail](#osde2e-prod-aws-e2e-next)|
|[osde2e-prod-aws-e2e-upgrade-prod-minus-one-to-next](https://prow.ci.openshift.org/?job=osde2e-prod-aws-e2e-upgrade-prod-minus-one-to-next)| <span style="color:#10ef00;">99.38%</span>|[More Detail](#osde2e-prod-aws-e2e-upgrade-prod-minus-one-to-next)|
|[osde2e-prod-aws-e2e-upgrade-prod-minus-three-to-next](https://prow.ci.openshift.org/?job=osde2e-prod-aws-e2e-upgrade-prod-minus-three-to-next)| <span style="color:#30cf00;">98.14%</span>|[More Detail](#osde2e-prod-aws-e2e-upgrade-prod-minus-three-to-next)|
|[osde2e-prod-aws-e2e-upgrade-prod-minus-two-to-next](https://prow.ci.openshift.org/?job=osde2e-prod-aws-e2e-upgrade-prod-minus-two-to-next)| <span style="color:#08f700;">99.69%</span>|[More Detail](#osde2e-prod-aws-e2e-upgrade-prod-minus-two-to-next)|
|[osde2e-prod-aws-e2e-upgrade-rescheduled](https://prow.ci.openshift.org/?job=osde2e-prod-aws-e2e-upgrade-rescheduled)| <span style="color:#10ef00;">99.38%</span>|[More Detail](#osde2e-prod-aws-e2e-upgrade-rescheduled)|
|[osde2e-stage-aws-e2e-default](https://prow.ci.openshift.org/?job=osde2e-stage-aws-e2e-default)| <span style="color:#01fe00;">100.00%</span>|[More Detail](#osde2e-stage-aws-e2e-default)|
|[osde2e-stage-aws-e2e-next-y](https://prow.ci.openshift.org/?job=osde2e-stage-aws-e2e-next-y)| <span style="color:#01fe00;">100.00%</span>|[More Detail](#osde2e-stage-aws-e2e-next-y)|
|[osde2e-stage-aws-e2e-next-z](https://prow.ci.openshift.org/?job=osde2e-stage-aws-e2e-next-z)| <span style="color:#10ef00;">99.38%</span>|[More Detail](#osde2e-stage-aws-e2e-next-z)|



## osde2e-prod-aws-e2e-default

Overall pass rate: <span style="color:#04fb00;">99.88%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1407911410154868736](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-aws-e2e-default/1407911410154868736) | 4.7.16-candidate |  | <span style="color:#01fe00;">100.00%</span>|
[1407971827719868416](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-aws-e2e-default/1407971827719868416) | 4.7.16-candidate |  | <span style="color:#01fe00;">100.00%</span>|
[1407609635107508224](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-aws-e2e-default/1407609635107508224) | 4.7.16-candidate |  | <span style="color:#01fe00;">100.00%</span>|
[1407730437643571200](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-aws-e2e-default/1407730437643571200) | 4.7.16-candidate |  | <span style="color:#10ef00;">99.38%</span>|<ul><li>[install] [Suite: operators] AlertmanagerInhibitions inhibits ClusterOperatorDegraded</li></ul>
[1407851103457906688](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-aws-e2e-default/1407851103457906688) | 4.7.16-candidate |  | <span style="color:#01fe00;">100.00%</span>|



## osde2e-prod-aws-e2e-next

Overall pass rate: <span style="color:#09f600;">99.68%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1407730438478237696](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-aws-e2e-next/1407730438478237696) | 4.8.0-rc.0-candidate |  | <span style="color:#01fe00;">100.00%</span>|
[1407851104313544704](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-aws-e2e-next/1407851104313544704) | 4.8.0-rc.0-candidate |  | <span style="color:#01fe00;">100.00%</span>|
[1407971828554534912](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-aws-e2e-next/1407971828554534912) | 4.8.0-rc.0-candidate |  | <span style="color:#01fe00;">100.00%</span>|
[1407609635812151296](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-aws-e2e-next/1407609635812151296) | 4.8.0-rc.0-candidate |  | <span style="color:#21de00;">98.73%</span>|<ul><li>[install] [Suite: e2e] Pods should be Running or Succeeded</li><li>[install] [Suite: operators] [OSD] Splunk Forwarder Operator Operator Upgrade should upgrade from the replaced version</li></ul>



## osde2e-prod-aws-e2e-upgrade-prod-minus-one-to-next

Overall pass rate: <span style="color:#10ef00;">99.38%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1407519012547465216](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-aws-e2e-upgrade-prod-minus-one-to-next/1407519012547465216) | 4.7.16-candidate | 4.8.0-rc.0 | <span style="color:#10ef00;">99.38%</span>|<ul><li>[upgrade] [Suite: operators] [OSD] Prune jobs pruner jobs should works builds-pruner should run successfully</li><li>[upgrade] [Suite: operators] [OSD] Splunk Forwarder Operator Operator Upgrade should upgrade from the replaced version</li></ul>
[1407881210885050368](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-aws-e2e-upgrade-prod-minus-one-to-next/1407881210885050368) | 4.7.16-candidate | 4.8.0-rc.0 | <span style="color:#10ef00;">99.38%</span>|<ul><li>[upgrade] [Suite: informing] [OSD] Splunk Forwarder Operator clusterServiceVersion openshift-splunk-forwarder-operator/splunk-forwarder-operator should be present and in succeeded state</li><li>[upgrade] [Suite: operators] [OSD] Splunk Forwarder Operator Operator Upgrade should upgrade from the replaced version</li></ul>



## osde2e-prod-aws-e2e-upgrade-prod-minus-three-to-next

Overall pass rate: <span style="color:#30cf00;">98.14%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1407639810004226048](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-aws-e2e-upgrade-prod-minus-three-to-next/1407639810004226048) | 4.7.14-candidate | 4.8.0-rc.0 | <span style="color:#30cf00;">98.14%</span>|<ul><li>[install] [Suite: operators] [OSD] Custom Domains Operator Should allow dedicated-admins to create domains Should be resolvable by external services</li><li>[upgrade] [Suite: e2e] [OSD] namespace validating webhook namespace validating webhook Members of SRE groups can manage all namespaces</li><li>[upgrade] [Suite: e2e] [OSD] namespace validating webhook namespace validating webhook Non-privileged users can manage all non-privileged namespaces</li><li>[upgrade] [Suite: operators] AlertmanagerInhibitions inhibits ClusterOperatorDegraded</li><li>[upgrade] [Suite: operators] CloudIngressOperator apischeme apischemes CR instance must be present on cluster</li><li>[upgrade] [Suite: operators] [OSD] Splunk Forwarder Operator Operator Upgrade should upgrade from the replaced version</li></ul>



## osde2e-prod-aws-e2e-upgrade-prod-minus-two-to-next

Overall pass rate: <span style="color:#08f700;">99.69%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1407579418573934592](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-aws-e2e-upgrade-prod-minus-two-to-next/1407579418573934592) | 4.7.15-candidate | 4.8.0-rc.0 | <span style="color:#08f700;">99.69%</span>|<ul><li>[upgrade] [Suite: scale-performance] Scaling should be tested with HTTP</li></ul>



## osde2e-prod-aws-e2e-upgrade-rescheduled

Overall pass rate: <span style="color:#10ef00;">99.38%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1407881210914410496](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-aws-e2e-upgrade-rescheduled/1407881210914410496) | 4.7.16-candidate | 4.7.18 | <span style="color:#10ef00;">99.38%</span>|<ul><li>[upgrade] [Suite: e2e] [OSD] Samesite Cookie Strict Validating samesite cookie should be set for openshift-monitoring OSD managed routes</li><li>[upgrade] [Suite: e2e] [OSD] namespace validating webhook namespace validating webhook Members of SRE groups can manage all namespaces</li></ul>



## osde2e-stage-aws-e2e-default

Overall pass rate: <span style="color:#01fe00;">100.00%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1407609643387064320](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-aws-e2e-default/1407609643387064320) | 4.7.16-candidate |  | <span style="color:#01fe00;">100.00%</span>|



## osde2e-stage-aws-e2e-next-y

Overall pass rate: <span style="color:#01fe00;">100.00%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1407911413271236608](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-aws-e2e-next-y/1407911413271236608) | 4.8.0-rc.0-candidate |  | <span style="color:#01fe00;">100.00%</span>|
[1407971836104282112](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-aws-e2e-next-y/1407971836104282112) | 4.8.0-rc.0-candidate |  | <span style="color:#01fe00;">100.00%</span>|



## osde2e-stage-aws-e2e-next-z

Overall pass rate: <span style="color:#10ef00;">99.38%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1407609645031231488](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-aws-e2e-next-z/1407609645031231488) | 4.7.17-candidate |  | <span style="color:#10ef00;">99.38%</span>|<ul><li>[install] [Suite: e2e] [OSD] namespace validating webhook namespace validating webhook Non-privileged users can manage all non-privileged namespaces</li></ul>




