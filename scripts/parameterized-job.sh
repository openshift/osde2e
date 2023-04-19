#!/bin/bash

set +e
# ensure we have a clean environment
podman rm -i osde2e-run

# bind mounts run into permissions issues, this creates
# the container and copies the secrets over to ensure it has perms
podman create --name osde2e-run -e OCM_TOKEN \
-e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY \
-e CLOUD_PROVIDER_REGION -e INSTANCE_TYPE \
-e REPORT_DIR=/tmp/osde2e-report quay.io/app-sre/osde2e test --configs rosa,stage,e2e-suite --skip-health-check
 
podman start -a osde2e-run

# copy the junit results xml for publishing
podman cp osde2e-run:/tmp/osde2e-report .