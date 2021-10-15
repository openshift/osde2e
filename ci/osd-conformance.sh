#!/usr/bin/env bash

export KUBECONFIG="${SHARED_DIR}/kubeconfig"

git clone https://github.com/openshift/origin.git
cd origin || exit 1

RELEASE_URL="https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest"

curl -s "${RELEASE_URL}/openshift-client-linux.tar.gz" | tar zxvf - oc

chmod +x oc

PATH="$(pwd):$PATH"

test/extended/conformance-k8s.sh