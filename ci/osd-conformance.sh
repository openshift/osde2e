#!/usr/bin/env bash

export KUBECONFIG="${SHARED_DIR}/kubeconfig"

git clone https://github.com/openshift/origin.git
cd origin || exit 1

RELEASE_URL="https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest"

curl -s "${RELEASE_URL}/openshift-client-linux.tar.gz" | tar zxvf - oc

chmod +x oc

PATH="$(pwd):$PATH"

echo "Giving the cluster time to settle. Sleeping for 1800 seconds.";

sleep 1800

test/extended/conformance-k8s.sh