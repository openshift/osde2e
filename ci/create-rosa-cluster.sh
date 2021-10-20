#!/usr/bin/env bash

set -o pipefail

OCM_URL="https://api.stage.openshift.com"
AWS_DEFAULT_REGION="us-east-1"
export AWS_DEFAULT_REGION

if [ -z "${OCM_TOKEN+x}" ]; then
    echo "Assuming the OCM token should be read from Prow";
    OCM_TOKEN=$(cat /usr/local/osde2e-credentials/ocm-refresh-token)
    export OCM_TOKEN
fi
if [ -z "${AWS_ACCESS_KEY_ID+x}" ] || [ -z "${AWS_SECRET_ACCESS_KEY+x}" ]; then
    echo "Assuming AWS creds should be read from Prow";
    AWS_ACCESS_KEY_ID=$(cat /usr/local/osde2e-rosa-staging/rosa-aws-access-key)
    export AWS_ACCESS_KEY_ID
    AWS_SECRET_ACCESS_KEY=$(cat /usr/local/osde2e-rosa-staging/rosa-aws-secret-access-key)
    export AWS_SECRET_ACCESS_KEY
fi

export OCM_CONFIG=./.ocm.json

# We do also need OCM in order to get the kubeconfig
GO111MODULE=on go get github.com/openshift-online/ocm-cli/cmd/ocm@master

if [ -z "${OCM_URL}" ]; then
    ocm login --token="${OCM_TOKEN}"
else 
    ocm login --url="${OCM_URL}" --token="${OCM_TOKEN}"
fi

GO111MODULE=on go get github.com/openshift/rosa/cmd/rosa@master

rosa login --token="${OCM_TOKEN}"

rosa create cluster --yes --region="${AWS_DEFAULT_REGION}" --cluster-name=rosa-conform > "$SHARED_DIR/cluster-info"

cat "$SHARED_DIR/cluster-info"

awk 'FNR == 8 {print $2}' "$SHARED_DIR/cluster-info" > "$SHARED_DIR/cluster-id"

echo "Confirming cluster id is $(cat "$SHARED_DIR/cluster-id")";

COUNT=1;
TOTAL=91
STATUS="unknown";
while [[ $COUNT -lt $TOTAL ]]
do  
    STATUS=$(rosa describe cluster --region="${AWS_DEFAULT_REGION}" -c "$(cat "$SHARED_DIR/cluster-id")" -o json | jq -r '.status.state')
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
