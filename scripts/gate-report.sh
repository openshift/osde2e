#!/bin/bash
#
# gate-report will run the gate report command against the provided arguments and upload the resulting report to the osde2e-metrics bucket.
#

set -e

SRC_DIR="$(cd $(dirname $0)/..; pwd)"
METRICS_BUCKET=osde2e-metrics
GATE_REPORT=gate-report
VENV="$(mktemp -d)"
REPORT_DIR="$(mktemp -d)"

trap 'rm -rf "$VENV" "$REPORT_DIR"' EXIT

if [[ $# -ne 2 ]]; then
	echo "Usage: $0 <environment> <version>"
	exit 1
fi

ENVIRONMENT="$1"
VERSION="$2"

virtualenv "$VENV"
. "$VENV/bin/activate"

pip install awscli

if ! aws s3 ls s3://$METRICS_BUCKET 2>&1 > /dev/null ; then
	echo "AWS CLI not configured properly."
	exit 1
fi

REPORT_FILE="$ENVIRONMENT-$VERSION-report.json"

docker pull quay.io/app-sre/osde2e
docker run -e PROMETHEUS_ADDRESS -e PROMETHEUS_BEARER_TOKEN -v "$REPORT_DIR:/report-output" quay.io/app-sre/osde2e gate-report -output "/report-output/$REPORT_FILE" "$ENVIRONMENT" "$VERSION"

aws s3 cp "$REPORT_DIR/$REPORT_FILE" "s3://$METRICS_BUCKET/$GATE_REPORT/$REPORT_FILE"
