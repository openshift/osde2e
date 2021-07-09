#!/usr/bin/env bash

echo "Installing oc binary"

curl -s https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/openshift-client-linux.tar.gz | tar zxvf - oc

chmod +x oc

export KUBECONFIG=${SHARED_DIR}/kubeconfig

EXISTING_VERSION=$(./oc get clusterversion version -o json | jq -r ".status.desired.version")

./oc patch clusterversion/version --patch "$(echo "${EXISTING_VERSION}" | awk -F \. {'print "{\"spec\":{\"channel\":\"candidate-"$1"."$2+1"\"}}"'})" --type=merge

AVAILABLE_UPGRADES=$(./oc adm upgrade | tail -n +6 | tr -s ' ' | cut -d' ' -f1)

STREAM=${1:-latest}

echo "Upgrades available:";
echo "${AVAILABLE_UPGRADES}";
echo "";

case $STREAM in
  latest)
    echo "Looking for latest available version";
    REGEX="^.*$"
  ;;
  y)
    echo "Looking for latest y version";
    REGEX="^$(echo "${EXISTING_VERSION}" | awk -F \. {'print $1"."$2+1'}).*$"
  ;;
  z)
    echo "Looking for latest z version";
    REGEX="^$(echo "${EXISTING_VERSION}" | awk -F \. {'print $1"."$2'}).*$"
  ;;
  *)
  echo "Invalid stream: $STREAM"; exit 1;
  ;;
esac

echo "";

UPGRADE_TARGET=$(echo "${AVAILABLE_UPGRADES}" | grep "${REGEX}" | tac | head -n 1)

echo "Targeting ${UPGRADE_TARGET}";

./oc adm upgrade --to="$UPGRADE_TARGET"

runtime="120 minute"
endtime=$(date -ud "$runtime" +%s)

while [[ $(date -u +%s) -le $endtime ]]
do  
    STATUS=$(./oc get clusterversion version -o json)
    PROGRESSING=$(echo "${STATUS}" | jq -r '.status.conditions[] | select(.type=="Progressing") | .status')
    PROGRESS_MESSAGE=$(echo "${STATUS}" | jq -r '.status.conditions[] | select(.type=="Progressing") | .message')
    FAILING=$(echo "${STATUS}"  | jq -r '.status.conditions[] | select(.type=="Failing") | .status')
    
    if [[ "${PROGRESSING}" == "False" ]]; then
        if [[ "${FAILING}" == "True" ]]; then
            echo "Upgrade failed: ${PROGRESS_MESSAGE}";
            exit 1;
        else
            echo "Upgrade complete";
            exit 0;
        fi
    else
        echo "Upgrade in progress: ${PROGRESS_MESSAGE}";
    fi

    sleep 1m
done