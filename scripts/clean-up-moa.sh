#!/usr/bin/env bash
#
# This will clean up the Velero S3 buckets left over during MOA testing.

if [ ! $# -eq 1 ]
then
	echo "No secrets directory supplied!"
	exit 1
fi

# Load secrets from a given directory.
MOA_AWS_ACCESS_KEY_ID="$(cat "$1/moa-aws-access-key")"
MOA_AWS_SECRET_ACCESS_KEY="$(cat "$1/moa-aws-secret-access-key")"
MOA_AWS_REGION="$(cat "$1/moa-aws-region")"

if [ -z "$MOA_AWS_ACCESS_KEY_ID" ]
then
	echo "No AWS access key found!"
	exit 1
fi

if [ -z "$MOA_AWS_SECRET_ACCESS_KEY" ]
then
	echo "No AWS secret access key found!"
	exit 1
fi

if [ -z "$MOA_AWS_REGION" ]
then
	echo "No AWS region found!"
	exit 1
fi

# The docker AWS CLI command to run. We want to make sure we populate AWS keys
# from the environment.
AWS() {
	docker run \
	-e AWS_ACCESS_KEY_ID="$MOA_AWS_ACCESS_KEY_ID" \
	-e AWS_SECRET_ACCESS_KEY="$MOA_AWS_SECRET_ACCESS_KEY" \
	-e AWS_REGION="$MOA_AWS_REGION" \
	--rm -it amazon/aws-cli "$@"
}

VELERO_BUCKETS=()
while IFS='' read -r line; do VELERO_BUCKETS+=("$line"); done < <(AWS s3 ls | grep managed-velero | awk '{print $3}')

for bucket in "${VELERO_BUCKETS[@]}"
do
	bucket="$(echo -e "$bucket" | tr -d '[:space:]')"
	echo "Deleting $bucket"
	if AWS s3 rb "s3://$bucket" --force
	then
		echo "Successfully deleted $bucket."
	else
		echo "Error deleting $bucket."
		exit 1
	fi
done
