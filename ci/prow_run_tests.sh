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
    extract_secret_from_dirs AWS_ACCESS_KEY_ID "$SECRETS" aws-access-key "AWS access key file"
    extract_secret_from_dirs AWS_SECRET_ACCESS_KEY "$SECRETS" aws-secret-access-key "AWS secret access key file"
    extract_secret_from_dirs AWS_REGION "$SECRETS" aws-region "AWS region file"

    if [ "$TEST" = "test-addon" ]; then
        extract_secret_from_dirs ADDON_IDS "$SECRETS" addon-ids "Addon IDs file"
        extract_secret_from_dirs ADDON_TEST_HARNESSES "$SECRETS" addon-test-harnesses "Addon test harnesses file"
    fi

    if [[ "$TEST" == test-scale* ]]; then
        extract_secret_from_dirs PBENCH_SSH_PRIVATE_KEY "$SECRETS" pbench-ssh-private-key "pbench private key file"
        extract_secret_from_dirs PBENCH_SSH_PUBLIC_KEY "$SECRETS" pbench-ssh-public-key "pbench public key file"
    fi

    # We explicitly want to make sure we're always uploading metrics from prow jobs.
    export UPLOAD_METRICS=true

    make $TEST
} 2>&1 | tee -a $REPORT_DIR/test_output.log
