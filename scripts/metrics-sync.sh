#!/bin/bash

METRICS_BUCKET=osde2e-metrics
INCOMING=incoming
PROCESSED=processed
VENV="$(mktemp -d)"
METRICS_DIR="$(mktemp -d)"

# Cleanup the temporary directories
trap 'rm -rf "$VENV" "$METRICS_DIR"' EXIT

virtualenv "$VENV"
. "$VENV/bin/activate"

pip install awscli

if ! aws s3 ls s3://$METRICS_BUCKET 2>&1 > /dev/null ; then
	echo "AWS CLI not configured properly."
	exit 1
fi

# We're going to iterate over each file as opposed to trying to grab things by wildcard.
# This way, we're guaranteed to process a fixed set of files.
METRICS_FILES=$(aws s3 ls "s3://$METRICS_BUCKET/$INCOMING/" | awk '{print $4}')

for file in $METRICS_FILES; do
	INCOMING_FILE="$METRICS_BUCKET/$INCOMING/$file"
	PROCESSED_FILE="$METRICS_BUCKET/$PROCESSED/$file"
	echo "Processing $file"
	
	if ! aws s3 cp "s3://$INCOMING_FILE" "$METRICS_DIR/$file"; then
		echo "Error copying $INCOMING_FILE from S3."
		exit 2
	fi

	JOB_NAME=$(echo $file | sed 's/^[^\.]*\.\(.*\)\.metrics\.prom$/\1/')
	if ! curl -X DELETE "$PUSHGATEWAY_URL/metrics/job/$JOB_NAME"; then
		echo "Error deleting old results for $JOB_NAME."
		exit 3
	fi

	if ! curl -T "$METRICS_DIR/$file" "$PUSHGATEWAY_URL/metrics/job/$JOB_NAME"; then
		echo "Error pushing new results for $JOB_NAME."
		exit 4
	fi

	if ! aws s3 mv "s3://$INCOMING_FILE" "s3://$PROCESSED_FILE"; then
		echo "Error moving $INCOMING_FILE to $PROCESSED_FILE in S3."
		exit 5
	fi
	echo "File has been processed and moved into the processed drectory."
done
