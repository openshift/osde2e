#!/usr/bin/env bash

echo "Installing Azure CLI"

curl -L https://aka.ms/InstallAzureCli | bash

echo "Installing oc binary"

curl -s https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/openshift-client-linux.tar.gz | tar zxvf - oc

chmod +x oc

echo "Setting config variables"

CLUSTER_NAME="osde2e-$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 6 | head -n 1)"

echo $CLUSTER_NAME > ${SHARED_DIR}/cluster-name

LOCATION=$2
PULL_SECRET_FILE="/usr/local/osde2e-credentials/stage-ocm-pull-secret"
 
if [ "$LOCATION" == "" ];
then
    LOCATION="eastus"
fi

# just automating the steps in https://docs.microsoft.com/en-us/azure/openshift/tutorial-create-cluster

echo "Logging into Azure"

az login --service-principal --username "$(cat /usr/local/osde2e-credentials/aro-app-id)" --password "$(cat /usr/local/osde2e-credentials/aro-password)" --tenant "$(cat /usr/local/osde2e-credentials/aro-tenant)"

echo "Creating required Azure objects"

RESOURCEGROUP_NAME=$CLUSTER_NAME

az provider register -n Microsoft.RedHatOpenShift --wait
az provider register -n Microsoft.Compute --wait
az provider register -n Microsoft.Storage --wait
az provider register -n Microsoft.Authorization --wait

az group create \
    --name $RESOURCEGROUP_NAME \
    --location $LOCATION
    
az network vnet create \
    --resource-group $RESOURCEGROUP_NAME \
    --name aro-vnet \
    --address-prefixes 10.0.0.0/22
    
az network vnet subnet create \
    --resource-group $RESOURCEGROUP_NAME \
    --vnet-name aro-vnet $CLUSTER_NAME \
    --resource-group $RESOURCEGROUP_NAME \
    --vnet-name aro-vnet \
    --disable-private-link-service-network-policies true

CREATE_CMD="az aro create --resource-group $RESOURCEGROUP_NAME --name $CLUSTER_NAME --vnet aro-vnet --master-subnet master-subnet --worker-subnet worker-subnet "

if [ "$PULL_SECRET_FILE" != "" ];
then
    CREATE_CMD="$CREATE_CMD --pull-secret @$PULL_SECRET_FILE"
fi

echo "Running ARO create command"

AROINFO="$(eval "$CREATE_CMD")"
KUBEAPI=$(echo "$AROINFO" | jq -r '.apiserverProfile.url')
KUBECRED=$(az aro list-credentials --name $CLUSTER_NAME --resource-group $CLUSTER_NAME)
KUBEUSER=$(echo "$KUBECRED" | jq -r '.kubeadminUsername')
KUBEPASS=$(echo "$KUBECRED" | jq -r '.kubeadminPassword')

oc login "$KUBEAPI" --username="$KUBEUSER" --password="$KUBEPASS"

oc config view --raw > ${SHARED_DIR}/kubeconfig