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
	-e ROSA_STS="${STS}" \
	-e CHANNEL \
	-e REPORT_DIR='/tmp/osde2e-report' \
	quay.io/redhat-services-prod/osde2e-cicada-tenant/osde2e:latest provision --reserve --configs "${CONFIGS}" 

# Start the container
docker start -a "${CONTAINER_NAME}"

# Copy the junit results xml for publishing
docker cp "${CONTAINER_NAME}":/tmp/osde2e-report .

 
