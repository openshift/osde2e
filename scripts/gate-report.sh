#!/bin/bash
#
# gate-report will run the gate report command against the provided arguments and upload the resulting report to the osde2e-metrics bucket.
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
docker run -e PROMETHEUS_ADDRESS -e PROMETHEUS_BEARER_TOKEN -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_REGION quay.io/app-sre/osde2e gate-report -output "s3://$METRICS_BUCKET/$GATE_REPORT/$REPORT_FILE" "$ENVIRONMENT" "$VERSION"
