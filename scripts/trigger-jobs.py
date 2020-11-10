#!/usr/bin/env python

import os
import re

import requests

BUILD_URL = os.environ['BUILD_URL']
JENKINS_URL = os.environ['JENKINS_URL']

UPSTREAM_JOBS = {
    "openshift-saas-deploy-saas-clusterimagesets-stage-osd-stage-01": "stage",
    "openshift-saas-deploy-saas-clusterimagesets-stage-osd-stage-hives02ue1": "stage",
    "openshift-saas-deploy-saas-clusterimagesets-prod-osd-production-hivep01ue1": "production",
    "openshift-saas-deploy-saas-clusterimagesets-integration-osd-integration-hivei01ue1": "integration",
}


def verify_upstream_jobs():
    for job in UPSTREAM_JOBS.keys():
        requests.head(f"{JENKINS_URL}/job/{job}").raise_for_status()


def get_upstream_job():
    job = requests.get(f"{BUILD_URL}/api/json").json()

    try:
        upstream_job = job['actions'][0]['causes'][0]['upstreamProject']
        upstream_build = job['actions'][0]['causes'][0]['upstreamBuild']
    except KeyError:
        return (None, None)

    return (upstream_job, upstream_build)


def get_changed_cis(job, build):
    r = requests.get(f"{JENKINS_URL}/job/{job}/{build}/consoleText")

    cis = []
    for encoded_line in r.iter_lines():
        line = encoded_line.decode()
        m = re.search(r"\['apply',.*'ClusterImageSet'.*'(openshift-.*)'", line)
        if m:
            cis.append(m.group(1))

    return cis


def trigger_cis_test(environment, cis):
    print(f"triggering cis test: {environment}-{cis}")


if __name__ == '__main__':
    verify_upstream_jobs()

    upstream_job, upstream_build = get_upstream_job()

    if upstream_job and upstream_build:
        if upstream_job in UPSTREAM_JOBS:
            environment = UPSTREAM_JOBS[upstream_job]
            for cis in get_changed_cis(upstream_job, upstream_build):
                trigger_cis_test(environment, cis)
        else:
            print("NOT A CIS JOB")
