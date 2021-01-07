#!/usr/bin/env bash
#
# Increase ROSA limits across environments.
#

request_quota_increase() {
	SERVICE_CODE="$1"
	QUOTA_CODE="$2"
	DESIRED_VALUE="$3"
	REGION="$4"

	SERVICE_QUOTA_FILE="$(mktemp)"
	trap 'rm "$SERVICE_QUOTA_FILE"' RETURN

	SHOULD_INCREASE_QUOTA="false"
	if ! aws service-quotas get-service-quota --service-code "$SERVICE_CODE" --quota-code "$QUOTA_CODE" --region "$REGION" 2>/dev/null > "$SERVICE_QUOTA_FILE"
	then
		SHOULD_INCREASE_QUOTA="true"
		echo "No quota found for $SERVICE_CODE:$QUOTA_CODE in $REGION. Will attempt to increase quota."
	else
		CURRENT_QUOTA_VALUE="$(jq -r '.Quota.Value' "$SERVICE_QUOTA_FILE")"

		if (( $(echo "$CURRENT_QUOTA_VALUE < $DESIRED_VALUE" | bc -l) ))
		then
			SHOULD_INCREASE_QUOTA="true"
			echo "Quota for $SERVICE_CODE:$QUOTA_CODE is $CURRENT_QUOTA_VALUE in $REGION. Will be increased to $DESIRED_VALUE"
		else
			echo "Quota for $SERVICE_CODE:$QUOTA_CODE is $CURRENT_QUOTA_VALUE in $REGION, which is greater than or equal to $DESIRED_VALUE. Skipping quota increase."
		fi
	fi

	if [ "$SHOULD_INCREASE_QUOTA" = "true" ]
	then
		aws service-quotas request-service-quota-increase --service-code "$SERVICE_CODE" --quota-code "$QUOTA_CODE" --desired-value "$DESIRED_VALUE" --region "$REGION"
	fi
}

REGIONS=()
mapfile -t REGIONS < <(aws ec2 describe-regions --region us-east-1 | jq -r '.Regions | .[] | .RegionName')

for REGION in "${REGIONS[@]}"
do
	echo "Increasing quota for $REGION..."

	# Quota numbers obtained from: https://github.com/openshift/rosa/blob/master/pkg/aws/quota.go
	# Number of EIPs - VPC EIPs
	request_quota_increase "ec2" "L-0263D0A3" "5.0" "$REGION"
	# Running On-Demand Standard (A, C, D, H, I, M, R, T, Z) instances
	request_quota_increase "ec2" "L-1216C47A" "200.0" "$REGION"
	# VPCs per Region
	request_quota_increase "vpc" "L-F678F1CE" "5.0" "$REGION"
	# Internet gateways per Region
	request_quota_increase "vpc" "L-A4707A72" "5.0" "$REGION"
	# Network interfaces per Region
	request_quota_increase "vpc" "L-DF5E4CA3" "5000.0" "$REGION"
	# General Purpose SSD (gp2) volume storage
	request_quota_increase "ebs" "L-D18FCD1D" "300.0" "$REGION"
	# Number of EBS snapshots
	request_quota_increase "ebs" "L-309BACF6" "300.0" "$REGION"
	# Provisioned IOPS
	request_quota_increase "ebs" "L-B3A130E6" "300000.0" "$REGION"
	# Provisioned IOPS SSD (io1) volume storage
	request_quota_increase "ebs" "L-FD252861" "300.0" "$REGION"
	# Application Load Balancers per Region
	request_quota_increase "elasticloadbalancing" "L-53DA6B97" "50.0" "$REGION"
	# Classic Load Balancers per Region
	request_quota_increase "elasticloadbalancing" "L-E9E9831D" "20.0" "$REGION"
done
