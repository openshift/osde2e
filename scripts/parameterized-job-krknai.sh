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

# ensure we have a clean environment
docker rm osde2e-krknai-run

# bind mounts run into permissions issues, this creates
# the container and copies the secrets over to ensure it has perms
docker create --pull=always --name osde2e-krknai-run \
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
	quay.io/redhat-services-prod/osde2e-cicada-tenant/osde2e:latest krkn-ai "${args[@]}"

docker start -a osde2e-krknai-run
rc=$?

# copy the krkn-ai results for publishing
docker cp osde2e-krknai-run:/tmp/"${REPORT_DIR}" .
exit $rc
