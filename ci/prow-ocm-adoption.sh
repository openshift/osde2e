#!/usr/bin/env bash

set -o pipefail

OCM_URL="https://api.stage.openshift.com"

if [ -z "${OCM_TOKEN+x}" ]; then
    echo "Assuming the OCM token should be read from Prow";
    OCM_TOKEN=$(cat /usr/local/osde2e-credentials/ocm-refresh-token)
fi

export OCM_CONFIG=./.ocm.json
export KUBECONFIG="${SHARED_DIR}/kubeconfig"

curl -s https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/openshift-client-linux.tar.gz | tar zxvf - oc

chmod +x oc

GO111MODULE=on go get github.com/openshift-online/ocm-cli/cmd/ocm@master

if [ -z "${OCM_URL}" ]; then
    ocm login --token="${OCM_TOKEN}"
else 
    ocm login --url="${OCM_URL}" --token="${OCM_TOKEN}"
fi

cat <<EOF | ./oc create -n openshift-monitoring -f -
apiVersion: v1
kind: ConfigMap
metadata:
    name: cluster-monitoring-config
    namespace: openshift-monitoring
data:
    config.yaml: |
        telemeterClient:
            telemeterServerURL: https://infogw.api.stage.openshift.com
EOF

sleep 600;
echo "Output new config"
./oc get -n openshift-monitoring configmap/cluster-monitoring-config -o yaml

COUNT=$(CLUSTER_ID=$(./oc get clusterversion -o jsonpath='{.items[].spec.clusterID}{""}'); ocm get clusters --parameter search="external_id is '${CLUSTER_ID}'" | jq '.size')

if [[ "$COUNT" == "0" ]]; then
    echo "No cluster found!";
    exit 1
else
    echo "Cluster found!";
    exit 0
fi