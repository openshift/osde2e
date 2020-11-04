#!/bin/bash

set -eo pipefail

# Returns the url for the console log upstream job that triggered this job
upstream_log_url() {
    # BUILD_URL and JENKINS_URL are defined by JENKINS. Failing if not defined
    [ -z "$BUILD_URL" ] && echo "ERROR: BUILD_URL not defined" >&2 && return 1
    [ -z "$JENKINS_URL" ] && echo "ERROR: JENKINS_URL not defined" >&2 && return 1

    UPSTREAM_PROJECT=$(curl -s "${BUILD_URL}/api/json" | jq -r '.actions[0].causes[0].upstreamProject')
    UPSTREAM_BUILD=$(curl -s "${BUILD_URL}/api/json" | jq -r '.actions[0].causes[0].upstreamBuild')
    echo "${JENKINS_URL}/job/${UPSTREAM_PROJECT}/${UPSTREAM_BUILD}/consoleText"
}

# Checks whether the job was triggered by a ClusterImageSet job
triggered_by_clusterimagesets() {
    UPSTREAM_LOG_URL="$1"

    [[ "${UPSTREAM_LOG_URL}" =~ "openshift-saas-deploy-saas-clusterimagesets" ]] && return 0 || return 1
}

# Returns the list of ClusterImageSets that appear in the console log of the job
# that triggered this job in Jenkins.
changed_clusterimagesets() {
    UPSTREAM_LOG_URL="$1"

    # Looking for a line like this one:
    # [2020-11-04 16:04:29] [INFO] [openshift_base.py:apply:207] - ['apply', 'hive-stage-01', 'cluster-scope', 'ClusterImageSet', 'openshift-v4.5.0-0.nightly-2020-11-04-132914-nightly']
    curl -s "${UPSTREAM_LOG_URL}" | sed -n "/\['apply',.*'ClusterImageSet'.*'openshift-.*'/s/^.*'\(openshift.*\)'\]$/\1/p"
}

## Example:
# UPSTREAM_LOG_URL=$(upstream_log_url)

# if triggered_by_clusterimagesets "$UPSTREAM_LOG_URL"; then
#     changed_clusterimagesets "$UPSTREAM_LOG_URL"
# fi
