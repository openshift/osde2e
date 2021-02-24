#!/usr/bin/env bash

RELEASE_URL="https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest"
RELEASE_IMAGE=$(curl -s "${RELEASE_URL}/release.txt" | grep 'Pull From: quay.io' | awk -F ' ' '{print $3}')

if [ -z "${AWS_ACCESS_KEY_ID+x}" ]; then
    echo "Assuming the AWS Access token should be read from Prow";
    AWS_ACCESS_KEY_ID=$(cat /usr/local/osde2e-credentials/aws-access-key-id)
    export AWS_ACCESS_KEY_ID
fi

if [ -z "${AWS_SECRET_ACCESS_KEY+x}" ]; then
    echo "Assuming the AWS Secret token should be read from Prow";
    AWS_SECRET_ACCESS_KEY=$(cat /usr/local/osde2e-credentials/aws-secret-access-key)
    export AWS_SECRET_ACCESS_KEY
fi

if [ -z "${PULL_SECRET_FILE+x}" ]; then
    echo "Assuming the Pull Secret should be read from Prow";
    PULL_SECRET_FILE=/usr/local/osde2e-credentials/stage-ocm-pull-secret
fi

export KUBECONFIG=${SHARED_DIR}/installer/auth/kubeconfig

curl -s "${RELEASE_URL}/openshift-client-linux.tar.gz" | tar zxvf - oc

chmod +x oc

./oc adm release extract --registry-config "${PULL_SECRET_FILE}" --command=openshift-install --to "$(pwd)/" "${RELEASE_IMAGE}"

chmod +x openshift-install

./openshift-install destroy cluster --dir="${SHARED_DIR}" --log-level info