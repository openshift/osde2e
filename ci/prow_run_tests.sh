#!/bin/bash

set -o pipefail

. "$(dirname "$0")/prow_setup.sh"

{
    # Usage
    if [ "$#" -lt "1" ] || [ "$#" -gt "2" ]; then
        echo "Usage: $0 <comma-separated-secrets-directories> [type-of-test]"
        exit 1
    fi

    SECRETS=$1
    TEST="test"

    if [ ! -z "$2" ]; then
        TEST="$TEST-$2"
    fi

    if [ -z "$SECRETS" ]; then
        echo "Secrets directories were not provided."
        exit 2
    fi

    # Extract the secrets
    extract_secret_from_dirs OCM_TOKEN "$SECRETS" ocm-refresh-token "OCM token file"
    extract_secret_from_dirs METRICS_AWS_ACCESS_KEY_ID "$SECRETS" metrics-aws-access-key "metrics AWS access key file"
    extract_secret_from_dirs METRICS_AWS_SECRET_ACCESS_KEY "$SECRETS" metrics-aws-secret-access-key "metrics AWS secret access key file"
    extract_secret_from_dirs METRICS_AWS_REGION "$SECRETS" metrics-aws-region "metrics AWS region file"

    extract_secret_from_dirs MOA_AWS_ACCESS_KEY_ID "$SECRETS" moa-aws-access-key "MOA AWS access key file" false
    extract_secret_from_dirs MOA_AWS_SECRET_ACCESS_KEY "$SECRETS" moa-aws-secret-access-key "MOA AWS secret access key file" false
    extract_secret_from_dirs MOA_AWS_REGION "$SECRETS" moa-aws-region "MOA AWS region file" false

    # We explicitly want to make sure we're always uploading metrics from prow jobs.
    export UPLOAD_METRICS=true

    make $TEST
} 2>&1 | tee -a $REPORT_DIR/test_output.log
