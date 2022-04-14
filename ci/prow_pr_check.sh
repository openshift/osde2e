#!/bin/bash

set -eo pipefail

. "$(dirname "$0")/prow_setup.sh"

{
    export GOFLAGS=""
    PG_HOST=$(cat /usr/local/osde2e-common/rds-host)
    PG_USER=$(cat /usr/local/osde2e-common/rds-user)
    PG_PORT=$(cat /usr/local/osde2e-common/rds-port)
    PG_PASS=$(cat /usr/local/osde2e-common/rds-pass)
    FORCE_REAL_DB_TESTS=1
    export PG_HOST PG_USER PG_PASS PG_PORT FORCE_REAL_DB_TESTS
    make check

    make build

    CLUSTER_ID=1rj34urrnnb6nfqjnrdbimosp50hcsus \
    GINKGO_SKIP="Must Gather Operator" \
    OCM_CCS="true" \
    ./out/osde2e test --configs=prod,aws,pr-check,e2e-suite --secret-locations=/usr/local/osde2e-common,/usr/local/osde2e-credentials,/usr/local/sd-cicd-aws-prod

} 2>&1 | tee -a "$REPORT_DIR/test_output.log"
