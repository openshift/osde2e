#!/usr/bin/env bash

set -euo pipefail

echo "Logging into Azure"

az login --service-principal --username "$(cat /usr/local/osde2e-credentials/aro-app-id)" --password "$(cat /usr/local/osde2e-credentials/aro-password)" --tenant "$(cat /usr/local/osde2e-credentials/aro-tenant)"

echo "Deleting ARO cluster $(cat "$SHARED_DIR/cluster-name")"

az aro delete --yes --name="$(cat "$SHARED_DIR/cluster-name")" --resource-group="$(cat "$SHARED_DIR/cluster-name")"
az group delete --yes --name="$(cat "$SHARED_DIR/cluster-name")"
