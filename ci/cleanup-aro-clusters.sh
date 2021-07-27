#!/usr/bin/env bash

set -euo pipefail

echo "Logging into Azure"

az login --service-principal --username "$(cat /usr/local/osde2e-credentials/aro-app-id)" --password "$(cat /usr/local/osde2e-credentials/aro-password)" --tenant "$(cat /usr/local/osde2e-credentials/aro-tenant)"

az aro list | jq '.[] | select(.name|test("osde2e-.")) | .name' | while read line ; do
    echo "Cleaning up old cluster cluster $line"

    az aro delete --yes --name="$line" --resource-group="$line"
    az group delete --yes --name="$line"
done