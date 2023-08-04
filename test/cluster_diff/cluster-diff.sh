#!/bin/bash

set -o pipefail

## Script to perform gap analysis on two clusters
## Usage: ./gap-analysis.sh ${OLD-VERSION-CLUSTERNAME} ${NEW-VERSION-CLUSTERNAME}
##
## The following are requisites for the script to work:
## 1. Both clusters should be in staging environment and not production.
## 2. Both clusters should have public scope of network.
## 3. Both clusters should be authenticated to OCM.
##
## Following will be checked by the script:
## 1. Difference in alerts
## 2. Difference in clusteroperators
## 3. Difference in projects
## 4. Difference in resources
## 5. Difference in apigroups
## 6. Difference in kube-apiserver configuration
## 7. Difference in kube-controller-manager configuration
## 8. Difference in kube-scheduler configuration
## 9. Difference in openshift-apiserver configuration
## 10. Difference in Cloud CredentialsRequests (offline)
##
## Artifacts for comparison will be created in current working directory

OLD_CLUSTER=$1
NEW_CLUSTER=$2

function Usage {
	echo "Usage: $0 OLD-VERSION-CLUSTERNAME NEW-VERSION-CLUSTERNAME"
	echo "Example: $0 test49 test410"
	exit 1
}

function Login {
	CLUSTER=$1
	echo "($CLUSTER) - Login to cluster"
	CLUSTERID=$(ocm list clusters --managed | grep -w "$CLUSTER" | awk '{ print $1 }')
	ocm get /api/clusters_mgmt/v1/clusters/"$CLUSTERID"/credentials | jq -r .kubeconfig >"$CLUSTER"-kubeconfig.txt
	export KUBECONFIG="$CLUSTER"-kubeconfig.txt
}

function Logout {
	CLUSTER=$1
	echo "($CLUSTER) - Logout from cluster"
	unset KUBECONFIG
	rm -rf "$CLUSTER"-kubeconfig.txt
}

function Alerts {
	CLUSTER=$1
	echo "($CLUSTER) - Gathering alerts"
	oc get prometheusrules -A -o json --as backplane-cluster-admin | jq -r '.items[].spec.groups[].rules[] | select(.alert!=null)|[.alert, .labels.severity] | @csv' | sort -f >"$CLUSTER"-alerts.txt
}

function Clusteroperators {
	CLUSTER=$1
	echo "($CLUSTER) - Gathering clusteroperators"
	oc get co -oname | awk -F"/" '{ print $2 }' >"$CLUSTER"-clusteroperators.txt
}

function Projects {
	CLUSTER=$1
	echo "($CLUSTER) - Gathering projects"
	oc get projects -oname | awk -F/ '{ print $2 }' >"$CLUSTER"-projects.txt
}

function Resources {
	CLUSTER=$1
	echo "($CLUSTER) - Gathering resources"
	oc api-resources -oname | sort >"$CLUSTER"-resources.txt
}

function Groups {
	CLUSTER=$1
	echo "($CLUSTER) - Gathering API groups"
	oc api-versions >"$CLUSTER"-groups.txt
}

function KubeAPI {
	CLUSTER=$1
	echo "($CLUSTER) - Gathering kube-apiserver configuration"
	oc -n openshift-kube-apiserver get cm config -ojson | jq -r '.data."config.yaml"' | jq >"$CLUSTER"-kube-apiserver.json
}

function KubeCM {
	CLUSTER=$1
	echo "($CLUSTER) - Gathering kube-controller-manager configuration"
	oc -n openshift-kube-controller-manager get cm config -ojson | jq -r '.data."config.yaml"' | jq >"$CLUSTER"-kube-controller-manager.json
}

function KubeSched {
	CLUSTER=$1
	echo "($CLUSTER) - Gathering kube-scheduler configuration"
	oc -n openshift-kube-scheduler get cm config -ojson | jq -r '.data."config.yaml"' | jq >"$CLUSTER"-kube-scheduler.json
}

function OpenshiftAPI {
	CLUSTER=$1
	echo "($CLUSTER) - Gathering openshift-apiserver configuration"
	oc -n openshift-apiserver get cm config -ojson | jq -r '.data."config.yaml"' | jq >"$CLUSTER"-openshift-apiserver.json
}

function CredReq {
	for version in "$OLD_CLUSTER_VERSION" "$NEW_CLUSTER_VERSION"; do
		for cloud in aws gcp azure; do
			echo "(${version}) - Gathering Cloud CredentialsRequest for ${cloud} (offline)"
			image="quay.io/openshift-release-dev/ocp-release:${version}-x86_64"
			if [[ "${version}" == *nightly* ]]; then
				image="registry.ci.openshift.org/ocp/release:${version}"
			fi
			oc adm release extract "${image}" --credentials-requests --cloud="$cloud" --to="${version}"-"${cloud}"-sts
			if [[ $? -ne 0 ]]; then
				echo "Failed to gather Cloud CredentialsRequest for $version-$cloud !! Please check if image exists.."
				exit 1
			fi
		done
		echo
	done
}

function CredReq-Diff {
	echo "===== Difference in Cloud CredentialsRequest for AWS, GCP and Azure ====="

	for cloud in aws gcp azure; do
		echo
		echo "$(echo $cloud | tr '[:lower:]' '[:upper:]'):"
		echo "----"
		diff --color "${OLD_CLUSTER_VERSION}"-"${cloud}"-sts "${NEW_CLUSTER_VERSION}"-"${cloud}"-sts
		if [[ $? -eq 0 ]]; then
			echo
			echo "No change in Cloud CredentialsRequest for $cloud"
			echo
		fi
	done
}

function Diff {
	CATEGORY=$1
	FORMAT=${2:-txt}

	echo "===== Difference in $CATEGORY ($OLD_CLUSTER_VERSION vs. $NEW_CLUSTER_VERSION) ====="
	echo
	diff --color --suppress-common-lines -y "${OLD_CLUSTER}"-"${CATEGORY}"."${FORMAT}" "${NEW_CLUSTER}"-"${CATEGORY}"."${FORMAT}"
	if [[ $? -eq 0 ]]; then
		echo "No change in $CATEGORY"
	fi
	echo
}

if ! [ -x "$(command -v oc)" ]; then
	echo "OpenShift CLI is not installed, downloading latest available"
	curl -Ls https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/openshift-client-linux.tar.gz | tar -zxvf - oc && chmod +x oc
	PATH="${PATH}:."
fi

if ! [ -x "$(command -v ocm)" ]; then
	echo "OCM CLI is not installed, downloading latest available"
	curl -Ls https://github.com/openshift-online/ocm-cli/releases/latest/download/ocm-linux-amd64 -o ocm && chmod +x ocm
	PATH="${PATH}:."
fi

if [[ -z "${OLD_CLUSTER}" || -z "${NEW_CLUSTER}" ]]; then
	Usage
fi

TEMP_DIRECTORY="$(mktemp -d --suffix=-gap-analysis --tmpdir=${ARTIFACT_DIR:-.})"
cd "$TEMP_DIRECTORY" || exit 1

echo
echo "++ Beginning Gap Analysis between $OLD_CLUSTER and $NEW_CLUSTER ++"
echo
echo "- Please wait for few minutes for script to finish..."
echo

OLD_CLUSTER_VERSION=$(ocm describe cluster "$OLD_CLUSTER" --json | jq -r .version.raw_id)
NEW_CLUSTER_VERSION=$(ocm describe cluster "$NEW_CLUSTER" --json | jq -r .version.raw_id)

echo "- Gathering artifacts from cluster $OLD_CLUSTER (version $OLD_CLUSTER_VERSION) and $NEW_CLUSTER (version $NEW_CLUSTER_VERSION)"
echo

for cluster in $OLD_CLUSTER $NEW_CLUSTER; do
	Login "$cluster"

	Alerts "$cluster"
	Clusteroperators "$cluster"
	Projects "$cluster"
	Resources "$cluster"
	Groups "$cluster"
	KubeAPI "$cluster"
	KubeCM "$cluster"
	KubeSched "$cluster"
	OpenshiftAPI "$cluster"

	Logout "$cluster"
	echo
done

## Offline gathering required CredentialRequests using `oc adm release extract` command
CredReq

## Showing diff of collected artifacts
for category in alerts clusteroperators projects resources groups; do
	Diff "$category"
done

for category in kube-apiserver kube-controller-manager kube-scheduler openshift-apiserver; do
	Diff "$category" json
done

## Perform diff for CredentialRequest artifacts
CredReq-Diff

echo
echo "- Gap analysis done! Refer to temp directory $TEMP_DIRECTORY to see all the artifacts.."
echo

## Uncomment the below line if want to remove artifacts
#cd ../ && rm -rf "$TEMP_DIRECTORY"
