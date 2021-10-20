#!/usr/bin/env bash

set -o pipefail

OCM_URL="https://api.stage.openshift.com"

if [ -z "${OCM_TOKEN+x}" ]; then
    echo "Assuming the OCM token should be read from Prow";
    OCM_TOKEN=$(cat /usr/local/osde2e-credentials/ocm-refresh-token)
fi

export OCM_CONFIG=./.ocm.json

GO111MODULE=on go get github.com/openshift-online/ocm-cli/cmd/ocm@master

if [ -z "${OCM_URL}" ]; then
    ocm login --token="${OCM_TOKEN}"
else 
    ocm login --url="${OCM_URL}" --token="${OCM_TOKEN}"
fi

ocm delete cluster "$(cat "$SHARED_DIR/cluster-id")"