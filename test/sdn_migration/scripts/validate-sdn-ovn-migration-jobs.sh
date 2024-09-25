#!/bin/bash

set +e

UNIQUE_ID=$(date +%s%N) # Generate a unique identifier for this instance
CONTAINER_NAME="sdn-ovn-migration-${UNIQUE_ID}" # Name of the container based on the unique identifier

# Check if the container already exists
if podman ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "Container ${CONTAINER_NAME} already exists. Exiting."
    exit 1
fi

# Create the container with environment variables and unique name
podman create --name "${CONTAINER_NAME}" -e OCM_TOKEN \
	-e AWS_ACCESS_KEY_ID \
	-e AWS_SECRET_ACCESS_KEY \
	-e AWS_REGION \
	-e CLUSTER_ID \
	-e CLUSTER_NAME \
	-e REPLICAS \
	-e GINKGO_LABEL_FILTER \
	-e AWS_HTTPS_PROXY \
	-e AWS_HTTP_PROXY \
	-e CA_BUNDLE\
	-e SUBNETS\
	-e REPORT_DIR='/tmp/osde2e-report' \
	quay.io/redhat-user-workloads/osde2e-cicada-tenant/test-suites/sdn-migration:latest


# Start the container
podman start -a "${CONTAINER_NAME}"

# Copy the junit results xml for publishing
podman cp "${CONTAINER_NAME}":/tmp/osde2e-report .

# Optionally, clean up by removing the container after use
podman rm "${CONTAINER_NAME}"
