package aws

import (
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
	session *session.Session
	iam     *iam.IAM
	ec2     *ec2.EC2
	once    sync.Once
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
				Region: aws.String(viper.GetString(config.CloudProvider.Region)),
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
	})
	if err != nil {
		log.Printf("error initializing AWS session: %v", err)
	}

	return nil
}
