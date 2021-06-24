#!/usr/bin/env bash

set -euo pipefail

source ./ci/install-azure-cli.sh

echo "Logging into Azure"

az login --service-principal --username "$(cat /usr/local/osde2e-credentials/aro-app-id)" --password "$(cat /usr/local/osde2e-credentials/aro-password)" --tenant "$(cat /usr/local/osde2e-credentials/aro-tenant)"

echo "Deleting ARO cluster $(cat {$SHARED_DIR}/cluster-name)"

az aro delete --name="$(cat {$SHARED_DIR}/cluster-name)" --resource-group="$(cat {$SHARED_DIR}/cluster-name)"
