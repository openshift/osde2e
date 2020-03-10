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

AWS="docker run -e AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID}" -e AWS_SECRET_ACCESS_KEY -v "$REPORT_DIR:/report-input" quay.io/app-sre/mesosphere-aws-cli"

if ! aws s3 ls s3://$METRICS_BUCKET 2>&1 > /dev/null ; then
	echo "AWS CLI not configured properly."
	exit 1
fi


REPORT_FILE="$ENVIRONMENT-$VERSION-report.json"

$AWS s3 cp "s3://$METRICS_BUCKET/$GATE_REPORT/$REPORT_FILE" "/report-input/$REPORT_FILE"

docker pull quay.io/app-sre/osde2e
docker run -v "$REPORT_DIR:/report-input" quay.io/app-sre/osde2e gate-report-analysis "/report-input/$REPORT_FILE"
