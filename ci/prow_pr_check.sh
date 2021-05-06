#!/bin/bash

set -o pipefail

. "$(dirname "$0")/prow_setup.sh"

{
    export GOFLAGS=""
    make check

    make build

    CLUSTER_ID=1kgq75e84eloshd81c8gcjl10fglg0jb \
    ./out/osde2e test --configs=prod,aws --secret-locations=/usr/local/osde2e-common,/usr/local/osde2e-credentials

} 2>&1 | tee -a "$REPORT_DIR/test_output.log"