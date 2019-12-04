#!/bin/bash

# Install the dependencies
(cd "$(dirname "$0")/.."; go get -u github.com/Masterminds/glide && glide install --strip-vendor)

# Ensure the report directory exists
mkdir -p "$REPORT_DIR"
