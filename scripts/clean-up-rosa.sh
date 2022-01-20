#!/usr/bin/env bash
#
# This will clean up the Velero S3 buckets left over during ROSA testing.

if [ ! $# -eq 1 ]
then
	echo "No secrets directory supplied!"
	exit 1
fi

if [ -n "$(command -v apt-get)" ]; then
	apt update
	apt install -y jq
fi
if [ -n "$(command -v yum)" ]; then
	yum install -y jq
fi
if [ -n "$(command -v apk)" ]; then
	apk update
	apk add jq
fi

# Load secrets from a given directory.
AWS_ACCESS_KEY_ID="$(cat "$1/rosa-aws-access-key")"
AWS_SECRET_ACCESS_KEY="$(cat "$1/rosa-aws-secret-access-key")"
AWS_REGION="$(cat "$1/rosa-aws-region")"

if [ -z "$AWS_ACCESS_KEY_ID" ]
then
	echo "No AWS access key found!"
	exit 1
fi

if [ -z "$AWS_SECRET_ACCESS_KEY" ]
then
	echo "No AWS secret access key found!"
	exit 1
fi

if [ -z "$AWS_REGION" ]
then
	echo "No AWS region found!"
	exit 1
fi

export AWS_ACCESS_KEY_ID
export AWS_SECRET_ACCESS_KEY
export AWS_REGION

VELERO_BUCKETS=()
while IFS='' read -r line; do VELERO_BUCKETS+=("$line"); done < <(aws s3 ls | grep managed-velero | awk '{print $3}')

for bucket in "${VELERO_BUCKETS[@]}"
do
	bucket="$(echo -e "$bucket" | tr -d '[:space:]')"
	echo "Deleting $bucket"
	if aws s3 rb "s3://$bucket" --force
	then
		echo "Successfully deleted $bucket."
	else
		echo "Error deleting $bucket."
		exit 1
	fi
done

for data in $(aws iam list-roles | jq '.Roles[] | select(.RoleName|test("osde2e-.")) | @base64');
do
	role=$(echo "$data" | tr -d '"' | base64 --decode)
	role_name=$(echo "$role" | jq -r '.RoleName')
	roleCreatedAt=$(echo "$role" | jq -r '.CreateDate')
	if (( $(date -d "$roleCreatedAt" +%s) < $(date -d "1 day ago" +%s) ))
	then
		echo "Deleting role $role_name"
		if aws iam delete-role --role-name "$role_name"
		then
			echo "Successfully deleted role $role_name."
		else
			echo "Error deleting role $role_name."
		fi
	fi
	sleep 2;
done