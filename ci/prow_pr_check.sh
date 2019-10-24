#!/bin/bash

set -o pipefail

{
    make check
    make out/osde2e
} 2>&1 | tee -a $REPORT_DIR/test_output.log
