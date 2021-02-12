#!/bin/bash

set -o pipefail

mkdir -p installer

export VERSION=latest
export RELEASE_IMAGE=$(curl -s https://mirror.openshift.com/pub/openshift-v4/clients/ocp/$VERSION/release.txt | grep 'Pull From: quay.io' | awk -F ' ' '{print $3}')

export cmd=openshift-install
export pullsecret_file=~/Downloads/crc-pull-secret.txt
export extract_dir=$(pwd)

if [ -z "$OCM_TOKEN" ]; then
    echo "Assuming the token should be read from Prow";
    export OCM_TOKEN=$(cat /usr/local/osde2e-credentials/ocm-refresh-token)
fi

if [ -z "AWS_ACCESS_KEY_ID" ]; then
    export AWS_ACCESS_KEY_ID=$(cat /usr/local/osde2e-credentials/aws-access-key-id)
fi

if [ -z "AWS_SECRET_ACCESS_KEY" ]; then
    export AWS_SECRET_ACCESS_KEY=$(cat /usr/local/osde2e-credentials/aws-secret-access-key)
fi


cp /usr/local/osde2e-credentials/stage-installer-config ./installer/installer-config.yaml



curl -s https://mirror.openshift.com/pub/openshift-v4/clients/ocp/$VERSION/openshift-client-linux.tar.gz | tar zxvf - oc

./oc adm release extract --registry-config "${pullsecret_file}" --command=$cmd --to "${extract_dir}" ${RELEASE_IMAGE}

./openshift-install create cluster --dir=$(pwd)/installer/ --log-level info

export OCM_CONFIG=$(pwd)/.ocm.json
export KUBECONFIG=$(pwd)/installer/auth/kubeconfig

{

go get -u github.com/openshift-online/ocm-cli/cmd/ocm

if [ -z "$OCM_URL" ]; then
    ocm login --token=$OCM_TOKEN
else 
    ocm login --url=$OCM_URL --token=$OCM_TOKEN

    cat <<EOF | oc create -n openshift-monitoring -f -
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