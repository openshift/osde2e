#!/bin/bash

set -o pipefail

{
    function exit_if_file_missing {
        FILE_DESCRIPTION="$1"
        FILE="$2"

        if [ ! -f "$FILE" ]; then
            echo "$FILE_DESCRIPTION ($FILE) does not exist."
            exit 3
        fi
    }

    # One argument, pointing to a secrets directory, is expected
    if [ "$#" -ne "1" ]; then
        echo "1 argument expected."
        exit 1
    fi

    SECRETS=$1

    if [ ! -d "$SECRETS" ]; then
        echo "Secrets directory ($SECRETS) does not exist."
        exit 2
    fi

    # Test the existence of each expected file
    UHC_TOKEN_FILE=$SECRETS/uhc-refresh-token
    TESTGRID_BUCKET_FILE=$SECRETS/testgrid-bucket
    TESTGRID_SERVICE_ACCOUNT_FILE=$SECRETS/testgrid-service-account
    AWS_ACCESS_KEY_FILE=$SECRETS/aws-access-key
    AWS_SECRET_ACCESS_KEY_FILE=$SECRETS/aws-secret-access-key

    exit_if_file_missing "UHC token file" $UHC_TOKEN_FILE
    exit_if_file_missing "Testgrid bucket file" $TESTGRID_BUCKET_FILE
    exit_if_file_missing "Testgrid service account file" $TESTGRID_SERVICE_ACCOUNT_FILE
    exit_if_file_missing "AWS access key file" $AWS_ACCESS_KEY_FILE
    exit_if_file_missing "AWS secret access key file" $AWS_SECRET_ACCESS_KEY_FILE

    export UHC_TOKEN=$(cat $UHC_TOKEN_FILE)
    export TESTGRID_BUCKET=$(cat $TESTGRID_BUCKET_FILE)
    export TESTGRID_SERVICE_ACCOUNT=$(cat $TESTGRID_SERVICE_ACCOUNT_FILE)
    export AWS_ACCESS_KEY_ID=$(cat $AWS_ACCESS_KEY_FILE)
    export AWS_SECRET_ACCESS_KEY=$(cat $AWS_SECRET_ACCESS_KEY_FILE)

    make test
} 2>&1 | tee -a $REPORT_DIR/test_output.log
