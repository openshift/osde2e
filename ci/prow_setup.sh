#!/bin/bash

# Install the dependencies
(cd "$(dirname "$0")/.."; go mod tidy && go mod vendor)

# Ensure the report directory exists
mkdir -p "$REPORT_DIR"
