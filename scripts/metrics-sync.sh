#!/bin/bash
#
# Metrics sync will takes metrics from the osde2e metrics bucket and sync them with the datahub Prometheus pushgateway.
#

METRICS_BUCKET=osde2e-metrics
INCOMING=incoming
PROCESSED=processed
VENV="$(mktemp -d)"
METRICS_DIR="$(mktemp -d)"
METRIC_TIMEOUT_IN_SECONDS=21600 # 6h in seconds
METRIC_ADDON_TIMEOUT_IN_SECONDS=46800 # 13h in seconds
METRIC_SCALE_TIMEOUT_IN_SECONDS=93600 # 26h in seconds
METRIC_IMAGESET_TIMEOUT_IN_SECONDS=93600 # 26h in seconds

PUSHGATEWAY_URL=${APPSRE_PUSHGW_URL%/}
PUSHGATEWAY_CREDS=${APPSRE_PUSHGW_CREDS}

# Cleanup the temporary directories
trap 'rm -rf "$VENV" "$METRICS_DIR"' EXIT

# First, we should detect any stale metrics and purge them if needed
METRICS_LAST_UPDATED=$(curl -s -H "Authorization: Basic ${PUSHGATEWAY_CREDS}" "$PUSHGATEWAY_URL/metrics" | grep "^push_time_seconds{.*" | grep -E 'osde2e|ocm-api-test' | sed 's/^.*job="\([[:alnum:]_.-]*\)".*\}\s*\(.*\)$/\1,\2/' | sort | uniq)
CURRENT_TIMESTAMP=$(date +%s)
for metric_and_timestamp in $METRICS_LAST_UPDATED; do
	JOB_NAME=$(echo -e "$metric_and_timestamp" | cut -f 1 -d,)
	TIMESTAMP=$(echo -e "$metric_and_timestamp" | cut -f 2 -d, | xargs -d '\n' printf "%.f")

	if echo "$JOB_NAME" | grep -qE ".*osde2e.*addon.*"; then
		TIMESTAMP_PLUS_TIMEOUT=$((TIMESTAMP + METRIC_ADDON_TIMEOUT_IN_SECONDS))
	elif echo "$JOB_NAME" | grep -qE ".*osde2e.*scale.*"; then
		TIMESTAMP_PLUS_TIMEOUT=$((TIMESTAMP + METRIC_SCALE_TIMEOUT_IN_SECONDS))
	elif echo "$JOB_NAME" | grep -qE ".*osde2e.*imageset.*"; then
		TIMESTAMP_PLUS_TIMEOUT=$((TIMESTAMP + METRIC_IMAGESET_TIMEOUT_IN_SECONDS))
	else
		TIMESTAMP_PLUS_TIMEOUT=$((TIMESTAMP + METRIC_TIMEOUT_IN_SECONDS))
	fi

	if (( TIMESTAMP_PLUS_TIMEOUT < CURRENT_TIMESTAMP )); then
		echo "Metrics for job $JOB_NAME have expired. Removing them from the pushgateway."
		if ! curl -s -H "Authorization: Basic ${PUSHGATEWAY_CREDS}" -X DELETE "$PUSHGATEWAY_URL/metrics/job/$JOB_NAME"; then
			echo "Error deleting old results for $JOB_NAME."
			exit 3
		fi
	fi
done

virtualenv "$VENV"

# shellcheck source=/dev/null
. "$VENV/bin/activate"

pip install awscli

if ! aws s3 ls s3://$METRICS_BUCKET > /dev/null 2>&1 ; then
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

	# shellcheck disable=SC2001
	JOB_NAME=$(echo "$file" | sed 's/^[^\.]*\.\(.*\)\.metrics\.prom$/\1/')
	if [[ ! $JOB_NAME = delete_* ]]; then
		if ! curl -s -H "Authorization: Basic ${PUSHGATEWAY_CREDS}" -X DELETE "$PUSHGATEWAY_URL/metrics/job/$JOB_NAME"; then
			echo "Error deleting old results for $JOB_NAME."
			exit 3
		fi

		if ! curl -s -H "Authorization: Basic ${PUSHGATEWAY_CREDS}" -T "$METRICS_DIR/$file" "$PUSHGATEWAY_URL/metrics/job/$JOB_NAME"; then
			echo "Error pushing new results for $JOB_NAME."
			exit 4
		fi

		if ! aws s3 mv "s3://$INCOMING_FILE" "s3://$PROCESSED_FILE"; then
			echo "Error moving $INCOMING_FILE to $PROCESSED_FILE in S3."
			exit 5
		fi
		echo "File has been processed and moved into the processed drectory."
	else
		echo "$file is a test file. Deleting it from S3."

		if ! aws s3 rm "s3://$INCOMING_FILE"; then
			echo "Error removing test file from S3."
			exit 6
		fi
	fi
done
