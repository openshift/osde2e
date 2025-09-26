#!/bin/bash

set +e

CONTAINER_NAME="osde2e-run"

# ensure we have a clean environment
docker rm "${CONTAINER_NAME}"

# Create the container with environment variables and unique name
docker create --name "${CONTAINER_NAME}" -e OCM_TOKEN \
	-e OCM_CLIENT_ID -e OCM_CLIENT_SECRET \
	-e AWS_ACCESS_KEY_ID \
	-e AWS_SECRET_ACCESS_KEY \
	-e AWS_ACCOUNT_ID \
	-e AWS_REGION \
	-e OCM_CCS \
	-e SKIP_DESTROY_CLUSTER \
	-e OCM_ENV="${ENVIRONMENT}" \
	-e ROSA_STS="${STS}" \
	-e AWS_VPC_SUBNET_IDS \
	-e CHANNEL \
	-e INSTALL_LATEST_XY \
	-e CLUSTER_VERSION \
	-e ROSA_BILLING_ACCOUNT_ID \
	-e GCP_CREDS_JSON \
	-e INSTALL_LATEST_NIGHTLY \
	-e REPORT_DIR='/tmp/osde2e-report' \
	-e USE_PROXY_FOR_INSTALL \
	-e SUBNET_IDS \
	-e TEST_HTTP_PROXY \
	-e TEST_HTTPS_PROXY \
	-e USER_CA_BUNDLE \
	quay.io/redhat-services-prod/osde2e-cicada-tenant/osde2e:latest test --configs "${CONFIGS}"  "${ADDITIONAL_ARGS}"

# Start the container
docker start -a "${CONTAINER_NAME}"

# Copy the junit results xml for publishing
docker cp "${CONTAINER_NAME}":/tmp/osde2e-report .

