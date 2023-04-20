package aws

import (
	"fmt"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
)

type ccsAwsSession struct {
	session   *session.Session
	accountId string
	iam       *iam.IAM
	ec2       *ec2.EC2
	once      sync.Once
}

// CCSAWSSession is the global AWS session for interacting with AWS.
var CcsAwsSession ccsAwsSession

// GetAWSSessions returns a new AWS type with the first AWS account in the config file. The session is cached for the rest of the program.
func (CcsAwsSession *ccsAwsSession) GetAWSSessions() error {
	var err error

	CcsAwsSession.once.Do(func() {
		awsProfile := viper.GetString(config.AWSProfile)
		awsAccessKey := viper.GetString(config.AWSAccessKey)
		awsSecretAccessKey := viper.GetString(config.AWSSecretAccessKey)

		options := session.Options{
			Config: aws.Config{
				Region: aws.String(viper.GetString(config.AWSRegion)),
			},
		}

		if awsProfile != "" {
			options.Profile = awsProfile
		} else if awsAccessKey != "" || awsSecretAccessKey != "" {
			options.Config.Credentials = credentials.NewStaticCredentials(awsAccessKey, awsSecretAccessKey, "")
		}

		CcsAwsSession.session, err = session.NewSessionWithOptions(options)
		CcsAwsSession.iam = iam.New(CcsAwsSession.session)
		CcsAwsSession.ec2 = ec2.New(CcsAwsSession.session)
		CcsAwsSession.accountId = viper.GetString(config.AWSAccountId)
	})
	if err != nil {
		log.Printf("error initializing AWS session: %v", err)
	}

	return nil
}

// GetCredentials returns the credentials for the current aws session
func (CcsAwsSession *ccsAwsSession) GetCredentials() (*credentials.Value, error) {
	if err := CcsAwsSession.GetAWSSessions(); err != nil {
		return nil, fmt.Errorf("failed to create aws session to retrieve credentials: %v", err)
	}

	creds, err := CcsAwsSession.session.Config.Credentials.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get aws credentials: %v", err)
	}
	return &creds, nil
}

// GetRegion returns the region set when the session was created
func (CcsAwsSession *ccsAwsSession) GetRegion() *string {
	return CcsAwsSession.session.Config.Region
}

// GetAccountId returns the aws account id in session
func (CcsAwsSession *ccsAwsSession) GetAccountId() string {
	return CcsAwsSession.accountId
}
