#!/bin/bash
#
# gate-report-analysis will run the gate report analysis command against the provided arguments and exit successfully if the report indicates a viable release.
#

set -e

SRC_DIR="$(cd $(dirname $0)/..; pwd)"
METRICS_BUCKET=osde2e-metrics
GATE_REPORT=gate-report
REPORT_DIR="$(mktemp -d)"

trap 'rm -rf "$REPORT_DIR"' EXIT

if [[ $# -ne 2 ]]; then
	echo "Usage: $0 <environment> <version>"
	exit 1
fi

ENVIRONMENT="$1"
VERSION="$2"
REPORT_FILE="$ENVIRONMENT-$VERSION-report.json"

docker pull quay.io/app-sre/osde2e
docker run -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_REGION quay.io/app-sre/osde2e gate-report-analysis ""s3://$METRICS_BUCKET/$GATE_REPORT/$REPORT_FILE""
