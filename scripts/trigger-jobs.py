#!/usr/bin/env python3

import os
import re
import sys
import pathlib

import subprocess
import requests

BUILD_URL = os.environ['BUILD_URL']
JENKINS_URL = os.environ['JENKINS_URL']
CWD = pathlib.Path(__file__).parent.absolute()
UID = os.getuid()

UPSTREAM_JOBS = {
    "openshift-saas-deploy-saas-clusterimagesets-stage-osd-stage-hives02ue1": "stage",
    "openshift-saas-deploy-saas-clusterimagesets-prod-osd-production-hivep01ue1": "prod",
    "openshift-saas-deploy-saas-clusterimagesets-integration-osd-integration-hivei01ue1": "int",
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


def trigger_cis_test(environment, cis, cloud_provider):
    run_array = [
        "docker", "run", f"-u {UID}", f"-v {CWD}/report:/report",
        "-e OCM_TOKEN", "-e REPORT_DIR=/report", f"-e CLOUD_PROVIDER_ID={cloud_provider}", f"-e OSD_ENV={environment}",
        f"-e CLUSTER_VERSION={cis}", "-e CLUSTER_EXPIRY_IN_MINUTES=240", "-e OCM_USER_OVERRIDE=ci-int-jenkins", 
        "quay.io/app-sre/osde2e", "test"
    ]

    print(" ".join(run_array))
    
    return os.system(" ".join(run_array))

if __name__ == '__main__':
    verify_upstream_jobs()

    upstream_job, upstream_build = get_upstream_job()

    if upstream_job and upstream_build:
        if upstream_job in UPSTREAM_JOBS:
            environment = UPSTREAM_JOBS[upstream_job]
            pathlib.Path(".").mkdir(parents=True, exist_ok=True)
            success = True
            for cis in get_changed_cis(upstream_job, upstream_build):
                if trigger_cis_test(environment, cis, "aws") > 0:
                    print("AWS Job failed!")
                    success = False
                if trigger_cis_test(environment, cis, "gcp") > 0:
                    print ("GCP Job failed!")
                    success = False
            
            sys.exit(not success)
        else:
            print("NOT A CIS JOB")
