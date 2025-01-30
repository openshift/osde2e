// Package config provides the configuration for tests run as part of the osde2e suite.
package config

import (
	"encoding/base64"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
)

const (
	profileA = "profile-a"
	profileB = "profile-b"
)

var (
	// AWSAccountId is the AWS account (Env var: AWS_ACCOUNT_ID)
	AWSAccountId = "config.aws.account"

	// AWSAccessKey is the AWS access key
	AWSAccessKey = "config.aws.accessKey"

	// AWSSecretAccessKey is the AWS secret access key
	AWSSecretAccessKey = "config.aws.secretAccessKey"

	// AWSSharedCredentials is the base64 encoded AWS credentials file content.
	// This is used to optimize AWS resource spending
	// by nuking each of the 2 accounts alternately.
	// Should contain two named profiles: "profile-a" and "profile-b"
	// If provided, supersedes AWS secret set in env.
	AWSSharedCredentials = "config.aws.sharedCredentials"

	// AWSCredentialsFile is the custom AWS credentials full file path including filename
	// This is where the provided credentials will be saved
	// AWS client uses this as a custom shared credentials file
	AWSCredentialsFile = "config.aws.sharedCredentialsFilePath"

	// AWSProfile is the AWS profile to use
	AWSProfile = "config.aws.profile"

	// AWSRegion is the AWS region to use
	AWSRegion = "config.aws.region"

	// AWSVPCSubnetIDs is comma-separated list of strings to specify the subnets for cluster provision
	AWSVPCSubnetIDs = "config.aws.vpcSubnetIDs"
)

func InitAWSViper() error {
	_ = viper.BindEnv(AWSSharedCredentials, "AWS_SHARED_CREDENTIALS")
	RegisterSecret(AWSSharedCredentials, "aws-shared-credentials")
	sharedCreds := viper.GetString(AWSSharedCredentials)

	viper.SetDefault(AWSCredentialsFile, os.Getenv("HOME")+"/.aws/osde2e/credentials")
	_ = viper.BindEnv(AWSCredentialsFile, "AWS_CREDENTIAL_FILE")
	customCredsPath := viper.GetString(AWSCredentialsFile)

	if sharedCreds != "" {
		// If shared credntials file is provided in env vars, it should contain two profiles named "profile-a" and profile-b"
		// Osde2e will use one of them based on current week.
		// While one profile is in use, the other is cleaned up using AWS nuke
		if err := os.MkdirAll(filepath.Dir(customCredsPath), os.FileMode(0o755)); err != nil {
			return fmt.Errorf("could not write given shared credentials file: %w", err)
		}

		data, err := base64.StdEncoding.DecodeString(sharedCreds)
		if err != nil {
			return fmt.Errorf("could not decode given shared credentials file. Ensure it is a valid base64 with no line breaks or spaces: %w", err)
		}

		// Write the string to the file
		if err = os.WriteFile(customCredsPath, data, os.ModePerm); err != nil {
			return fmt.Errorf("could not write given shared credentials file: %w", err)
		}

		// use profile based on week
		week := getWeekSince2024()
		currentProfile := profileA
		if week%2 != 0 {
			currentProfile = profileB
		}
		// remove secrets set in environment so that profile can take effect
		// by default, AWS gives higher precedence to secret env vars than profile.
		os.Setenv("AWS_ACCESS_KEY_ID", "")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "")
		os.Setenv("AWS_ACCOUNT_ID", "")
		os.Setenv("AWS_PROFILE", currentProfile)
	}
	_ = viper.BindEnv(AWSProfile, "AWS_PROFILE")
	RegisterSecret(AWSProfile, "aws-profile")

	_ = viper.BindEnv(AWSAccountId, "AWS_ACCOUNT_ID")
	RegisterSecret(AWSAccountId, "aws-account")

	_ = viper.BindEnv(AWSAccessKey, "AWS_ACCESS_KEY", "OCM_AWS_ACCESS_KEY", "AWS_ACCESS_KEY_ID", "ROSA_AWS_ACCESS_KEY_ID")
	RegisterSecret(AWSAccessKey, "aws-access-key")

	_ = viper.BindEnv(AWSSecretAccessKey, "AWS_SECRET_ACCESS_KEY", "OCM_AWS_SECRET_KEY", "ROSA_AWS_SECRET_ACCESS_KEY")
	RegisterSecret(AWSSecretAccessKey, "aws-secret-access-key")

	_ = viper.BindEnv(AWSRegion, "AWS_REGION", "ROSA_AWS_REGION", "CLOUD_PROVIDER_REGION")
	RegisterSecret(AWSRegion, "aws-region")

	_ = viper.BindEnv(AWSVPCSubnetIDs, "AWS_VPC_SUBNET_IDS", "ROSA_SUBNET_IDS", "SUBNET_IDS")
	RegisterSecret(AWSVPCSubnetIDs, "subnet-ids")

	return nil
}

// Since simply checking whether current week is odd or even
// within current year may result in unexpected outage if a year has
// odd number of weeks, use odd/even based on a constant start date.
func getWeekSince2024() int {
	timeFormat := "2006-01-02"
	t, _ := time.Parse(timeFormat, "2024-01-01")
	now := time.Now()
	duration := now.Sub(t)
	week := int(math.Floor(duration.Hours()/(24*7))) + 1
	return week
}
