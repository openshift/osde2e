#!/bin/bash

set +e
# ensure we have a clean environment
docker rm osde2e-run

# bind mounts run into permissions issues, this creates
# the container and copies the secrets over to ensure it has perms
docker create --pull=always --name osde2e-run -e OCM_TOKEN \
	-e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_ACCOUNT_ID \
	-e GCP_CREDS_JSON \
	-e CLOUD_PROVIDER_REGION \
	-e ROSA_AWS_REGION="${CLOUD_PROVIDER_REGION}" \
	-e ROSA_ENV="${ENVIRONMENT}" \
	-e OCM_ENV="${ENVIRONMENT}" \
	-e ROSA_STS="${STS}" \
	-e ROSA_MINT_MODE="${MINT_MODE}" \
	-e INSTANCE_TYPE \
	-e SKIP_DESTROY_CLUSTER \
	-e SKIP_CLUSTER_HEALTH_CHECKS \
	-e CLUSTER_ID \
	-e SKIP_MUST_GATHER \
	-e INSTALL_LATEST_XY \
	-e INSTALL_LATEST_NIGHTLY \
	-e TEST_HARNESSES \
	-e POLLING_TIMEOUT \
	-e OCM_CCS \
	-e MULTI_AZ \
	-e REPORT_DIR="/tmp/${REPORT_DIR}" \
	quay.io/redhat-services-prod/osde2e-cicada-tenant/osde2e:latest test --configs "${CONFIGS}" "${ADDITIONAL_FLAGS}"

docker start -a osde2e-run

# copy the junit results xml for publishing
docker cp osde2e-run:/tmp/osde2e-report .
