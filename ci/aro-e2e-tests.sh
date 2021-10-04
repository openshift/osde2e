#!/usr/bin/env bash

export GOFLAGS=""
export KUBECONFIG=${SHARED_DIR}/kubeconfig
export CLUSTER=$(cat "$SHARED_DIR/cluster-name")
export RESOURCEGROUP=$(cat "$SHARED_DIR/cluster-name")
export CI=""
export LOCATION="eastus"
export AZURE_CLIENT_ID="$(cat /usr/local/osde2e-credentials/aro-app-id)"
export AZURE_CLIENT_SECRET="$(cat /usr/local/osde2e-credentials/aro-password)"
export AZURE_TENANT_ID="$(cat /usr/local/osde2e-credentials/aro-tenant)"
export AZURE_SUBSCRIPTION_ID="$(cat /usr/local/osde2e-credentials/aro-subscription)"

RP_COMMIT=$(curl --silent https://arorpversion.blob.core.windows.net/rpversion/$LOCATION)
git clone https://github.com/Azure/ARO-RP.git
cd ARO-RP
git checkout $RP_COMMIT
make test-e2e
