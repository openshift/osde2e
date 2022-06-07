// Package config provides the configuration for tests run as part of the osde2e suite.
package config

import (
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
)

const (
	//AWSAccount, AWSAccessKey, AwsSecretAccessKey and AWSRegion are used to create an AWS session for testing Operators and CCS.
	// AWSAccount is used in CCS clusters
	AWSAccount = "config.aws.account"
	// AWSAccessKey is used in CCS clusters
	AWSAccessKey = "config.aws.accessKey"
	// AWSSecretKey is used in CCS clusters
	AWSSecretAccessKey = "config.aws.secretKey"
	// AWSRegion for provisioning clusters.
	AWSRegion = "config.aws.Region"

	// AWSVPCSubnetIDs is used in CCS clusters
	AWSVPCSubnetIDs = "config.aws.vpcSubnetIDs"
)

func InitViperAws() {
	viper.BindEnv(AWSAccount, "OCM_AWS_ACCOUNT", "AWS_ACCOUNT")
	RegisterSecret(AWSAccount, "ocm-aws-account")

	viper.BindEnv(AWSAccessKey, "OCM_AWS_ACCESS_KEY", "AWS_ACCESS_KEY_ID", "ROSA_AWS_ACCESS_KEY_ID", "AWS_ACCESS_KEY")
	RegisterSecret(AWSAccessKey, "aws-access-key")

	viper.BindEnv(AWSSecretAccessKey, "OCM_AWS_SECRET_KEY", "AWS_SECRET_ACCESS_KEY", "ROSA_AWS_SECRET_ACCESS_KEY", "AWS_SECRET_ACCESS_KEY")
	RegisterSecret(AWSSecretAccessKey, "aws-secret-access-key")

	viper.BindEnv(AWSVPCSubnetIDs, "OCM_AWS_VPC_SUBNET_IDS", "BYO_VPC")
	RegisterSecret(AWSVPCSubnetIDs, "aws-vpc-subnet-ids")

	viper.BindEnv(AWSRegion, "AWS_REGION", "ROSA_AWS_REGION")
	RegisterSecret(AWSRegion, "aws-region")
}
