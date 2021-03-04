#!/bin/bash

set -o pipefail

. "$(dirname "$0")/prow_setup.sh"

{
    export GOFLAGS=""
    make check
} 2>&1 | tee -a "$REPORT_DIR/test_output.log"
