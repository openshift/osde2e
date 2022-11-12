// Package config provides the configuration for tests run as part of the osde2e suite.
package config

import (
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
)

var (
	// AWSAccount is the AWS account
	AWSAccount = "config.aws.account"
	// AWSAccessKey is the AWS access key
	AWSAccessKey = "config.aws.accessKey"
	// AWSSecretKey is the AWS secret access key
	AWSSecretAccessKey = "config.aws.secretAccessKey"
	// AWSRegion sets the AWS region to use
	AWSRegion = "config.aws.region"

	// AWSVPCSubnetIDs is comma-separated list of strings to specify the subnets for cluster provision
	AWSVPCSubnetIDs = "config.aws.vpcSubnetIDs"
)

func InitAWSViper() {
	viper.BindEnv(AWSAccount, "AWS_ACCOUNT")
	RegisterSecret(AWSAccount, "aws-account")

	viper.BindEnv(AWSAccessKey, "OCM_AWS_ACCESS_KEY", "AWS_ACCESS_KEY_ID", "ROSA_AWS_ACCESS_KEY_ID", "AWS_ACCESS_KEY_ID")
	RegisterSecret(AWSAccessKey, "aws-access-key")

	viper.BindEnv(AWSSecretAccessKey, "OCM_AWS_SECRET_KEY", "AWS_SECRET_ACCESS_KEY", "ROSA_AWS_SECRET_ACCESS_KEY", "AWS_SECRET_ACCESS_KEY")
	RegisterSecret(AWSSecretAccessKey, "aws-secret-access-key")

	viper.BindEnv(AWSRegion, "ROSA_AWS_REGION", "AWS_REGION")
	RegisterSecret(AWSRegion, "aws-region")

	viper.BindEnv(AWSVPCSubnetIDs, "ROSA_SUBNET_IDS", "SUBNET_IDS")
	RegisterSecret(AWSVPCSubnetIDs, "subnet-ids")
}
