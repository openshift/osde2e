#!/bin/bash

set -o pipefail

{
    make check
} 2>&1 | tee -a $REPORT_DIR/test_output.log
