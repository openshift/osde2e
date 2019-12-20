#!/bin/bash

# General functions
function extract_secret_from_dirs {
    VAR_NAME="$1"
    DIRECTORIES="$2"
    FILE_NAME="$3"
    FILE_DESCRIPTION="$4"

    IFS=',' read -ra SECRET_DIR_ARRAY <<< "$DIRECTORIES"
    for SECRET_DIR in "${SECRET_DIR_ARRAY[@]}"; do
        SECRET_FILE="$SECRET_DIR/$FILE_NAME"
        if [ -f "$SECRET_FILE" ]; then
            export $VAR_NAME="$(cat "$SECRET_FILE")"
        fi
    done

    if [ -z "${!VAR_NAME}" ]; then
        echo "Required $FILE_DESCRIPTION does not exist or has no value."
        exit 3
    fi
}

if [ -z "$REPORT_DIR" ]; then
    export REPORT_DIR=/tmp/artifacts
fi

# Install the dependencies
(cd "$(dirname "$0")/.."; go mod tidy && go mod vendor)

# Ensure the report directory exists
mkdir -p "$REPORT_DIR"
