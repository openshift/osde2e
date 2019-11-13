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
    if [ "$#" -lt "1" ] || [ "$#" -gt "2" ]; then
        echo "1 or 2 arguments expected."
        exit 1
    fi

    SECRETS=$1
    TEST="test"

    if [ ! -z "$2" ]; then
        TEST="$TEST-$2"
    fi

    if [ ! -d "$SECRETS" ]; then
        echo "Secrets directory ($SECRETS) does not exist."
        exit 2
    fi

    # Test the existence of each expected file
    UHC_TOKEN_FILE=$SECRETS/uhc-refresh-token
    AWS_ACCESS_KEY_FILE=$SECRETS/aws-access-key
    AWS_SECRET_ACCESS_KEY_FILE=$SECRETS/aws-secret-access-key

    exit_if_file_missing "UHC token file" $UHC_TOKEN_FILE
    exit_if_file_missing "AWS access key file" $AWS_ACCESS_KEY_FILE
    exit_if_file_missing "AWS secret access key file" $AWS_SECRET_ACCESS_KEY_FILE

    export UHC_TOKEN=$(cat $UHC_TOKEN_FILE)
    export AWS_ACCESS_KEY_ID=$(cat $AWS_ACCESS_KEY_FILE)
    export AWS_SECRET_ACCESS_KEY=$(cat $AWS_SECRET_ACCESS_KEY_FILE)

    make $TEST
} 2>&1 | tee -a $REPORT_DIR/test_output.log
