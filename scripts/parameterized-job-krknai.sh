#!/bin/bash

set +e

args=(
  --configs "${CONFIGS}"
  --skip-destroy-cluster
  --log-analysis-enable
)

if [ "$SKIP_MUST_GATHER" = "true" ]; then
  args+=(--skip-must-gather)
fi

podman rm -f osde2e-krknai-run

PODMAN_SOCKET_STARTED=false
# shellcheck disable=SC2329
cleanup_podman_socket() {
	if [ "${PODMAN_SOCKET_STARTED}" = true ]; then
		systemctl --user stop podman.socket 2>/dev/null || true
	fi
}
trap cleanup_podman_socket EXIT

if ! systemctl --user is-active --quiet podman.socket 2>/dev/null; then
	if systemctl --user start podman.socket; then
		PODMAN_SOCKET_STARTED=true
	fi
fi

systemctl --user status podman.socket

export PODMAN_SOCK=/run/user/${UID}/podman/podman.sock

CONTAINER_SOCK_INNER="unix:///var/run/podman.sock"

HOST_REPORT="/tmp/${REPORT_DIR}"
HOST_SHARED="/tmp/${SHARED_DIR}"
mkdir -p "${HOST_REPORT}" "${HOST_SHARED}"
chmod a+rwx "${HOST_REPORT}" "${HOST_SHARED}" 2>/dev/null || true

# Run as the Jenkins/agent UID and keep host mapping so we can use the rootless podman socket (0600).
# Without --userns=keep-id, rootless Podman maps container UIDs to subuids; the process no longer matches
# the socket owner even with --user $(id -u), causing "connect: permission denied".
podman create --pull=always --name osde2e-krknai-run \
	--userns=keep-id \
	--security-opt label=disable \
	--user "$(id -u):$(id -g)" \
	-v "${PODMAN_SOCK}:/var/run/podman.sock:z" \
	-v "${HOST_REPORT}:${HOST_REPORT}:z" \
	-v "${HOST_SHARED}:${HOST_SHARED}:z" \
	-e "CONTAINER_HOST=${CONTAINER_SOCK_INNER}" \
	-e "DOCKER_HOST=${CONTAINER_SOCK_INNER}" \
	-e HOME=/tmp \
	-e XDG_CONFIG_HOME=/tmp/.config \
	-e XDG_DATA_HOME=/tmp/.local/share \
	-e TMPDIR=/tmp \
	-e OCM_CLIENT_ID -e OCM_CLIENT_SECRET \
	-e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_ACCOUNT_ID \
	-e GCP_CREDS_JSON \
	-e CLOUD_PROVIDER_REGION \
	-e ROSA_AWS_REGION="${CLOUD_PROVIDER_REGION}" \
	-e ROSA_ENV="${ENVIRONMENT}" \
	-e OCM_ENV="${ENVIRONMENT}" \
	-e ROSA_STS="${STS}" \
	-e HYPERSHIFT \
	-e CLUSTER_ID \
	-e KRKN_NAMESPACE \
	-e KRKN_POD_LABEL \
	-e KRKN_NODE_LABEL \
	-e KRKN_SKIP_POD_NAME \
	-e KRKN_FITNESS_QUERY \
	-e KRKN_SCENARIOS \
	-e KRKN_GENERATIONS \
	-e KRKN_POPULATION \
	-e KRKN_HEALTH_CHECK \
	-e KRKN_TOP_SCENARIOS_COUNT \
	-e GEMINI_API_KEY \
	-e REPORT_DIR="${HOST_REPORT}" \
	-e SHARED_DIR="${HOST_SHARED}" \
	quay.io/vkadapar_openshift/osde2e:local krkn-ai "${args[@]}"

podman start -a osde2e-krknai-run
rc=$?

# copy results for publishing (same as parameterized-job.sh docker cp)
podman cp "osde2e-krknai-run:${HOST_REPORT}" .

exit $rc
