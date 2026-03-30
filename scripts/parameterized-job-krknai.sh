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

: "${XDG_RUNTIME_DIR:=/run/user/$(id -u)}"
PODMAN_SOCK="${XDG_RUNTIME_DIR}/podman/podman.sock"

# If we start podman system service in the background, stop it on exit.
PODMAN_SVC_PID=""
cleanup_podman_service() {
	if [ -n "${PODMAN_SVC_PID}" ]; then
		kill "${PODMAN_SVC_PID}" 2>/dev/null
		wait "${PODMAN_SVC_PID}" 2>/dev/null || true
		PODMAN_SVC_PID=""
	fi
	return 0
}

# Start the podman socket if it isn't already running (e.g. before systemctl --user start podman.socket)
if [ ! -S "${PODMAN_SOCK}" ]; then
	mkdir -p "$(dirname "${PODMAN_SOCK}")"
	podman system service --time=0 "unix://${PODMAN_SOCK}" &
	PODMAN_SVC_PID=$!
	for _ in $(seq 1 20); do
		[ -S "${PODMAN_SOCK}" ] && break
		sleep 0.5
	done
	if [ ! -S "${PODMAN_SOCK}" ]; then
		echo "ERROR: podman API socket not available at ${PODMAN_SOCK}" >&2
		cleanup_podman_service
		exit 1
	fi
fi

trap cleanup_podman_service EXIT

podman create --pull=always --name osde2e-krknai-run \
	--privileged \
	-v "${PODMAN_SOCK}:/run/podman/podman.sock" \
	-e CONTAINER_HOST="unix:///run/podman/podman.sock" \
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
	quay.io/redhat-services-prod/osde2e-cicada-tenant/osde2e:latest krkn-ai "${args[@]}"

podman start -a osde2e-krknai-run
rc=$?

# copy results for publishing (same as parameterized-job.sh docker cp)
podman cp osde2e-krknai-run:/tmp/osde2e-report .

exit $rc
