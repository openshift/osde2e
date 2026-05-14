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
	-e OCM_CLIENT_ID -e OCM_CLIENT_SECRET \
	-e AWS_ACCESS_KEY_ID \
	-e AWS_SECRET_ACCESS_KEY \
	-e AWS_ACCOUNT_ID \
	-e AWS_REGION \
	-e ROSA_STS="${STS}" \
	-e CHANNEL \
	-e REPORT_DIR='/tmp/osde2e-report' \
	quay.io/redhat-services-prod/osde2e-cicada-tenant/osde2e:latest provision --reserve --configs "${CONFIGS}"

# Start the container and capture exit code
docker start -a "${CONTAINER_NAME}"
EXIT_CODE=$?

# Copy the junit results xml for publishing (even on failure, for debugging)
docker cp "${CONTAINER_NAME}":/tmp/osde2e-report . || true

# Clean up by removing the container
docker rm "${CONTAINER_NAME}"

# Exit with the osde2e exit code
exit ${EXIT_CODE}
