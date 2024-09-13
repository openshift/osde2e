#!/bin/bash

set +e

UNIQUE_ID=$(date +%s%N) # Generate a unique identifier for this instance
CONTAINER_NAME="osde2e-${UNIQUE_ID}" # Name of the container based on the unique identifier

# Check if the container already exists
if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "Container ${CONTAINER_NAME} already exists. Exiting."
    exit 1
fi

# Create the container with environment variables and unique name
docker create --name "${CONTAINER_NAME}" -e OCM_TOKEN \
	-e AWS_ACCESS_KEY_ID \
	-e AWS_SECRET_ACCESS_KEY \
	-e AWS_ACCOUNT_ID \
	-e AWS_REGION \
	-e AWS_VPC_SUBNET_IDS \
	-e GCP_CREDS_JSON \
	-e INSTALL_LATEST_NIGHTLY \
	-e REPORT_DIR='/tmp/osde2e-report' \
	quay.io/redhat-services-prod/osde2e-cicada-tenant/osde2e:latest test --configs "${CONFIGS}"

# Start the container
docker start -a "${CONTAINER_NAME}"

# Copy the junit results xml for publishing
docker cp "${CONTAINER_NAME}":/tmp/osde2e-report .

# Optionally, clean up by removing the container after use
docker rm "${CONTAINER_NAME}"
