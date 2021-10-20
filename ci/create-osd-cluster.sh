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

ocm create cluster --region=us-east-1 osd-conformance > "$SHARED_DIR/cluster-info"

cat "$SHARED_DIR/cluster-info"

awk 'FNR == 2 {print $2}' "$SHARED_DIR/cluster-info" > "$SHARED_DIR/cluster-id"

echo "Confirming cluster id is $(cat "$SHARED_DIR/cluster-id")";

COUNT=1;
TOTAL=91
STATUS="unknown";
while [[ $COUNT -lt $TOTAL ]]
do  
    STATUS=$(ocm get cluster "$(cat "$SHARED_DIR/cluster-id")"| jq -r '.status.state')
    if [[ "${STATUS}" == "ready" ]]; then
        ocm get "/api/clusters_mgmt/v1/clusters/$(cat "$SHARED_DIR/cluster-id")/credentials" | jq -r .kubeconfig > "$SHARED_DIR/kubeconfig"

        echo "KUBECONFIG retrieved!"
        exit 0
    else
        echo "Try #${COUNT} - Cluster is currently ${STATUS}";
        sleep 60
    fi
    COUNT=$((COUNT+1))
done

echo "Cluster failed to come up in time"
exit 1;
