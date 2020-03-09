#!/bin/bash
#
# gate-report-analysis will run the gate report analysis command against the provided arguments and exit successfully if the report indicates a viable release.
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

aws s3 cp "s3://$METRICS_BUCKET/$GATE_REPORT/$REPORT_FILE" "$REPORT_DIR/$REPORT_FILE"

docker run -v "$REPORT_DIR:/report-input" quay.io/app-sre/osde2e gate-report-analysis "/report-input/$REPORT_FILE"
