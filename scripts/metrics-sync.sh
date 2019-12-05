#!/bin/bash

METRICS_BUCKET=osde2e-metrics
INCOMING=incoming
PROCESSED=processed
VENV="$(mktemp -d)"

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
METRICS_DIR="$(mktemp -d)"

for file in $METRICS_FILES; do
	INCOMING_FILE="$METRICS_BUCKET/$INCOMING/$file"
	PROCESSED_FILE="$METRICS_BUCKET/$PROCESSED/$file"
	echo "Processing $file"
	
	if ! aws s3 cp "s3://$INCOMING_FILE" "$METRICS_DIR/$file"; then
		echo "Error copying $INCOMING_FILE from S3."
		exit 2
	fi

	# TODO: push this file to the datahub pushgateway, exit 3 is for errors in during this process

	if ! aws s3 mv "s3://$INCOMING_FILE" "s3://$PROCESSED_FILE"; then
		echo "Error moving $INCOMING_FILE to $PROCESSED_FILE in S3."
		exit 4
	fi
	echo "File has been processed and moved into the processed drectory."
done
