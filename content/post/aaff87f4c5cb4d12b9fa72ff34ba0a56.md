+++
title = "OSDe2e gcp Weather Report 2021-12-10 12:00:35.402723126 +0000 UTC"
author = "OSDe2e Automation"
date = "2021-12-10 12:00:35.402723126 +0000 UTC"
tags = ["weather-report", "gcp"]
summary = "<table class=\"summary\"><tr><td bgcolor=\"#24db00\"></td><td>int (Pass rate: 98.60)</td></tr><tr><td bgcolor=\"#12ed00\"></td><td>prod (Pass rate: 99.33)</td></tr><tr><td bgcolor=\"#2dd200\"></td><td>stage (Pass rate: 98.24)</td></tr></table>"
+++
## Summary

| Job Name | Pass Rate | More detail |
|----------|-----------|-------------|
|[osde2e-int-gcp-e2e-next-z](https://prow.ci.openshift.org/?job=osde2e-int-gcp-e2e-next-z)| <span style="color:#24db00;">98.60%</span>|[More Detail](#osde2e-int-gcp-e2e-next-z)|
|[osde2e-prod-gcp-e2e-default](https://prow.ci.openshift.org/?job=osde2e-prod-gcp-e2e-default)| <span style="color:#15ea00;">99.20%</span>|[More Detail](#osde2e-prod-gcp-e2e-default)|
|[osde2e-prod-gcp-e2e-next](https://prow.ci.openshift.org/?job=osde2e-prod-gcp-e2e-next)| <span style="color:#0ef100;">99.47%</span>|[More Detail](#osde2e-prod-gcp-e2e-next)|
|[osde2e-stage-gcp-e2e-default](https://prow.ci.openshift.org/?job=osde2e-stage-gcp-e2e-default)| <span style="color:#34cb00;">98.00%</span>|[More Detail](#osde2e-stage-gcp-e2e-default)|
|[osde2e-stage-gcp-e2e-next-z](https://prow.ci.openshift.org/?job=osde2e-stage-gcp-e2e-next-z)| <span style="color:#29d600;">98.40%</span>|[More Detail](#osde2e-stage-gcp-e2e-next-z)|



## osde2e-int-gcp-e2e-next-z

Overall pass rate: <span style="color:#24db00;">98.60%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1468973829131866112](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-int-gcp-e2e-next-z/1468973829131866112) | 4.9.11-candidate |  | <span style="color:#29d600;">98.40%</span>|<ul><li>[install] [Suite: e2e] [OSD] namespace validating webhook namespace validating webhook dedicated admins cannot manage privileged namespaces</li><li>[install] [Suite: operators] CloudIngressOperator deployment should have all desired replicas ready</li></ul>
[1469094630724210688](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-int-gcp-e2e-next-z/1469094630724210688) | 4.9.11-candidate |  | <span style="color:#3ec100;">97.60%</span>|<ul><li>[install] [Suite: e2e] [OSD] namespace validating webhook namespace validating webhook dedicated admins cannot manage privileged namespaces</li><li>[install] [Suite: operators] CloudIngressOperator deployment should have all desired replicas ready</li><li>[install] [Suite: operators] [OSD] Configure AlertManager Operator Operator Upgrade should upgrade from the replaced version</li></ul>
[1469215507516231680](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-int-gcp-e2e-next-z/1469215507516231680) | 4.9.11-candidate |  | <span style="color:#29d600;">98.40%</span>|<ul><li>[install] [Suite: e2e] [OSD] namespace validating webhook namespace validating webhook dedicated admins cannot manage privileged namespaces</li><li>[install] [Suite: operators] CloudIngressOperator rh-api-test cidr block changes should updated the service</li></ul>
[1468853021944320000](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-int-gcp-e2e-next-z/1468853021944320000) | 4.9.11-candidate |  | <span style="color:#01fe00;">100.00%</span>|



## osde2e-prod-gcp-e2e-default

Overall pass rate: <span style="color:#15ea00;">99.20%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1468853025316540416](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-default/1468853025316540416) | 4.9.9-candidate |  | <span style="color:#01fe00;">100.00%</span>|
[1469094637435097088](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-default/1469094637435097088) | 4.9.9-candidate |  | <span style="color:#29d600;">98.40%</span>|<ul><li>[install] [Suite: operators] CloudIngressOperator deployment should have all desired replicas ready</li><li>[install] [Suite: operators] CloudIngressOperator rh-api-test cidr block changes should updated the service</li></ul>
[1469215510854897664](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-default/1469215510854897664) | 4.9.9-candidate |  | <span style="color:#15ea00;">99.20%</span>|<ul><li>[install] [Suite: operators] CloudIngressOperator rh-api-test cidr block changes should updated the service</li></ul>



## osde2e-prod-gcp-e2e-next

Overall pass rate: <span style="color:#0ef100;">99.47%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1468853026172178432](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-next/1468853026172178432) | 4.9.11-candidate |  | <span style="color:#01fe00;">100.00%</span>|
[1468973833259061248](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-next/1468973833259061248) | 4.9.11-candidate |  | <span style="color:#01fe00;">100.00%</span>|
[1469215512561979392](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-prod-gcp-e2e-next/1469215512561979392) | 4.9.11-candidate |  | <span style="color:#29d600;">98.40%</span>|<ul><li>[install] [Suite: operators] CloudIngressOperator deployment should have all desired replicas ready</li><li>[install] [Suite: operators] CloudIngressOperator rh-api-test cidr block changes should updated the service</li></ul>



## osde2e-stage-gcp-e2e-default

Overall pass rate: <span style="color:#34cb00;">98.00%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1468973836551589888](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-default/1468973836551589888) | 4.9.9-candidate |  | <span style="color:#29d600;">98.40%</span>|<ul><li>[install] [Suite: e2e] [OSD] namespace validating webhook namespace validating webhook dedicated admins cannot manage privileged namespaces</li><li>[install] [Suite: operators] [OSD] Must Gather Operator Operator Upgrade should upgrade from the replaced version</li></ul>
[1469215516760477696](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-default/1469215516760477696) | 4.9.9-candidate |  | <span style="color:#3ec100;">97.60%</span>|<ul><li>[install] [Suite: e2e] [OSD] namespace validating webhook namespace validating webhook dedicated admins cannot manage privileged namespaces</li><li>[install] [Suite: operators] CloudIngressOperator deployment should have all desired replicas ready</li><li>[install] [Suite: operators] CloudIngressOperator rh-api-test cidr block changes should updated the service</li></ul>



## osde2e-stage-gcp-e2e-next-z

Overall pass rate: <span style="color:#29d600;">98.40%</span>

| Job ID | Install Version | Upgrade Version | Pass Rate | Failures |
|--------|-----------------|-----------------|-----------|----------|
[1468973838304808960](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-next-z/1468973838304808960) | 4.9.11-candidate |  | <span style="color:#15ea00;">99.20%</span>|<ul><li>[install] [Suite: e2e] [OSD] namespace validating webhook namespace validating webhook dedicated admins cannot manage privileged namespaces</li></ul>
[1469094655038590976](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-next-z/1469094655038590976) | 4.9.11-candidate |  | <span style="color:#29d600;">98.40%</span>|<ul><li>[install] [Suite: e2e] [OSD] namespace validating webhook namespace validating webhook dedicated admins cannot manage privileged namespaces</li><li>[install] [Suite: operators] CloudIngressOperator rh-api-test cidr block changes should updated the service</li></ul>
[1469215518438199296](https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/osde2e-stage-gcp-e2e-next-z/1469215518438199296) | 4.9.11-candidate |  | <span style="color:#3ec100;">97.60%</span>|<ul><li>[install] [Suite: e2e] [OSD] namespace validating webhook namespace validating webhook dedicated admins cannot manage privileged namespaces</li><li>[install] [Suite: operators] CloudIngressOperator deployment should have all desired replicas ready</li><li>[install] [Suite: operators] CloudIngressOperator rh-api-test cidr block changes should updated the service</li></ul>




