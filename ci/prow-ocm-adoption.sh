#!/bin/bash

set -o pipefail

export OCM_CONFIG=$(pwd)/.ocm.json

{

go get -u github.com/openshift-online/ocm-cli/cmd/ocm

if [ -z "$OCM_TOKEN" ]; then
    echo "Assuming the token should be read from Prow";
    export OCM_TOKEN=$(cat /usr/local/osde2e-credentials/ocm-refresh-token)
fi

if [ -z "$OCM_URL" ]; then
    ocm login --token=$OCM_TOKEN
else 
    ocm login --url=$OCM_URL --token=$OCM_TOKEN
    echo "Old Config: "
    oc get -n openshift-monitoring configmap/cluster-monitoring-config -o yaml
    cat <<EOF | oc apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: cluster-monitoring-config
  namespace: openshift-monitoring
data:
  config.yaml: |
    prometheusK8s:
      retention: 15d
      volumeClaimTemplate:
        metadata:
          name: prometheus-data
        spec:
          storageClassName: gp2
          resources:
            requests:
              storage: 50Gi
    alertmanagerMain:
      volumeClaimTemplate:
        metadata:
          name: alertmanager-data
        spec:
          storageClassName: gp2
          resources:
            requests:
              storage: 10Gi
    telemeterClient:
      telemeterServerURL: https://infogw.api.stage.openshift.com
EOF
    sleep 600;
    echo "New config"
    oc get -n openshift-monitoring configmap/cluster-monitoring-config -o yaml
fi

COUNT=$(CLUSTER_ID=$(oc get clusterversion -o jsonpath='{.items[].spec.clusterID}{"\n"}'); ocm get clusters --parameter search="external_id is '$CLUSTER_ID'" | jq '.size')

if [[ "$COUNT" == "0" ]]; then
    echo "No cluster found!";
    exit 1
else
    echo "Cluster found!";
    exit 0
fi

} 2>&1 | tee -a /tmp/artifacts/test_output.log