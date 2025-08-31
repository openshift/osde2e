#!/bin/bash

set +e
docker rm osde2e-run

docker create --pull=always --name osde2e-run \
  -e OCM_TOKEN \
  -e OCM_CLIENT_ID -e OCM_CLIENT_SECRET \
  -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_ACCOUNT_ID \
  -e GCP_CREDS_JSON \
  -e SKIP_DESTROY_CLUSTER \
  -e SKIP_CLUSTER_HEALTH_CHECKS \
  -e CLUSTER_ID \
  -e SKIP_MUST_GATHER \
  -e INSTALL_LATEST_Y_FROM_DELTA \
  -e UPGRADE_TO_LATEST_Y \
  -e CONFIGS \
  -e REPORT_DIR="/tmp/${REPORT_DIR}" \
  quay.io/redhat-services-prod/osde2e-cicada-tenant/osde2e:latest \
    test --configs "${CONFIGS}"

docker start -a osde2e-run
docker cp osde2e-run:/tmp/osde2e-report .
