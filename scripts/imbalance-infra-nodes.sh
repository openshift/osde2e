#!/bin/bash

echo "INFO: start putting the infra nodes into an imbalanced state"

AZ_COUNT=$(oc get nodes -o yaml | grep -E "^\s+failure-domain.beta.kubernetes.io/zone" | sort | uniq | wc -l)

echo "INFO: Number of AZs ${AZ_COUNT}"

# target workloads
TARGET_PODS="alertmanager-main|prometheus-k8s|splunkforwarder-deployment"

# check all the target pods on each infra node
for NODE in $(oc get node -l node-role.kubernetes.io/infra= --no-headers | awk '{print $1}'); do
    NUM=$(oc describe node "${NODE}" | grep -cE "${TARGET_PODS}")
    echo "INFO: there are ${NUM} target pods running on infra node ${NODE}"
done

# get the first infra node
INFRA_NODE=$(oc get node -l node-role.kubernetes.io/infra= --no-headers | awk '{print $1}' | head -n 1)

# get the infra node with the maximum target pods INFRA_NODE
for NODE in $(oc get node -l node-role.kubernetes.io/infra= --no-headers | awk '{print $1}'); do
    # set infra nodes NotSchedulable
    oc adm cordon "${NODE}"
    if [ "$(oc describe node "${INFRA_NODE}" | grep -cE "${TARGET_PODS}")" -lt "$(oc describe node "${NODE}" | grep -cE "${TARGET_PODS}")" ]; then
        INFRA_NODE=${NODE}
    fi
done

# set INFRA_NODE Schedulable
oc adm uncordon "${INFRA_NODE}"

# delete target pods on infra nodes except INFRA_NODE
for NODE in $(oc get node -l node-role.kubernetes.io/infra= --no-headers | awk '{print $1}'); do
    if [ "${INFRA_NODE}" != "${NODE}" ]; then
        if [ "$(oc describe node "${NODE}" | grep -cE "${TARGET_PODS}")" -ne 0 ]; then
            # delete the target workloads on infra nodes
            for POD in $(oc describe node "${NODE}" | grep -E "${TARGET_PODS}" | awk '{print $2}'); do
                NS=$(oc get pods --all-namespaces -o wide --field-selector spec.nodeName="${NODE}" | grep "${POD}" | awk '{print $1}')

                VOLUME_NAME=""
                if [[ "${POD}" == "alertmanager-main-"* ]]; then
                    VOLUME_NAME="alertmanager-data"
                fi
                if [[ "${POD}" == "prometheus-k8s-"* ]]; then
                    VOLUME_NAME="prometheus-data"
                fi
                if [ "${VOLUME_NAME}" ] && [ "${AZ_COUNT}" -ne 1 ]; then
                    # delete PVC
                    PVC=$(oc get pod -n "${NS}" "${POD}" -o jsonpath='{.spec.volumes[?(@.name=="'$VOLUME_NAME'")].persistentVolumeClaim.claimName}')
                    echo "INFO: Deleting PVC ${PVC} in NS ${NS}"
                    oc delete pvc --wait=false -n "${NS}" "${PVC}"
                fi

                echo "INFO: Deleting POD ${POD} in NS ${NS}"
                oc delete pod "${POD}" -n "${NS}"
            done
        fi
    fi
done

# check the number of the target pods scheduled to INFRA_NODE
for NODE in $(oc get node -l node-role.kubernetes.io/infra= --no-headers | awk '{print $1}'); do
    NUM=$(oc describe node "${NODE}" | grep -cE "${TARGET_PODS}")
    echo "INFO: there are ${NUM} target pods running on infra node ${NODE}"
done

# make the infra nodes schedulable
for NODE in $(oc get node -l node-role.kubernetes.io/infra= --no-headers | awk '{print $1}'); do
    if [ "${INFRA_NODE}" != "${NODE}" ]; then
        oc adm uncordon "${NODE}"
    fi
done

echo "INFO: the infra nodes are imbalanced!"