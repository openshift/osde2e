#!/usr/bin/env bash

set -euo pipefail

echo "Logging into Azure"

az login --service-principal --username "$(cat /usr/local/osde2e-credentials/aro-app-id)" --password "$(cat /usr/local/osde2e-credentials/aro-password)" --tenant "$(cat /usr/local/osde2e-credentials/aro-tenant)"

az aro list | jq '.[] | select(.name|test("osde2e-.")) | .name' | while read line ; do
    cluster_id=$(echo "$line" | sed -e 's/^"//' -e 's/"$//')

    echo "Cleaning up old cluster cluster $cluster_id"

    az aro delete --yes --name="$cluster_id" --resource-group="$cluster_id"
    az group delete --yes --name="$cluster_id"
done