#!/bin/bash

set +e

UNIQUE_ID=$(date +%s%N) # Generate a unique identifier for this instance
CONTAINER_NAME="sdn-ovn-migration-${UNIQUE_ID}" # Name of the container based on the unique identifier

# Create the container with environment variables and unique name
docker create --name "${CONTAINER_NAME}" -e OCM_TOKEN \
	-e AWS_ACCESS_KEY_ID \
	-e AWS_SECRET_ACCESS_KEY \
	-e AWS_REGION \
	-e CLUSTER_ID \
	-e CLUSTER_NAME \
	-e REPLICAS \
	-e GINKGO_LABEL_FILTER \
	-e TEST_HTTPS_PROXY \
	-e TEST_HTTP_PROXY \
	-e USER_CA_BUNDLE\
	-e SUBNET_IDS\
	-e REPORT_DIR='/tmp/osde2e-report' \
	quay.io/redhat-services-prod/osde2e-cicada-tenant/osde2e/sdn-migration:latest


# Start the container
docker start -a "${CONTAINER_NAME}"

# Copy the junit results xml for publishing
docker cp "${CONTAINER_NAME}":/tmp/osde2e-report .

# Optionally, clean up by removing the container after use
docker rm "${CONTAINER_NAME}"
