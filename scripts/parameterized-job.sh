#!/bin/bash

set +e
# ensure we have a clean environment
docker rm osde2e-run

# bind mounts run into permissions issues, this creates
# the container and copies the secrets over to ensure it has perms
docker create --name osde2e-run -e OCM_TOKEN \
	-e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY \
	-e CLOUD_PROVIDER_REGION -e ROSA_AWS_REGION="${CLOUD_PROVIDER_REGION}" \
	-e ROSA_ENV=stage -e ROSA_STS=true -e INSTANCE_TYPE \
	-e REPORT_DIR="${REPORT_DIR}" quay.io/app-sre/osde2e test --configs rosa,stage,e2e-suite 

docker start -a osde2e-run

# copy the junit results xml for publishing
docker cp osde2e-run:/tmp/osde2e-report .
