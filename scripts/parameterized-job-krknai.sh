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

podman create --pull=always --name osde2e-krknai-run \
	-v "${PODMAN_SOCK}:/var/run/podman.sock" \
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
	-e REPORT_DIR="/tmp/${REPORT_DIR}" \
	-e SHARED_DIR="/tmp/${SHARED_DIR}" \
	quay.io/vkadapar_openshift/osde2e:local krkn-ai "${args[@]}"

podman start -a osde2e-krknai-run
rc=$?

# copy results for publishing (same as parameterized-job.sh docker cp)
podman cp osde2e-krknai-run:/tmp/osde2e-report .

exit $rc
