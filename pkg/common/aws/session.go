package aws

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	ec2v2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	iamv2 "github.com/aws/aws-sdk-go-v2/service/iam"
	s3v2 "github.com/aws/aws-sdk-go-v2/service/s3"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
)

type ccsAwsSession struct {
	cfg       aws.Config
	accountId string
	iam       *iamv2.Client
	s3        *s3v2.Client
	ec2       *ec2v2.Client
	once      sync.Once
}

// CcsAwsSession is the global AWS session for interacting with AWS.
var CcsAwsSession ccsAwsSession

// GetAWSSessions initializes the AWS config and service clients. The result is cached for the rest of the program.
func (CcsAwsSession *ccsAwsSession) GetAWSSessions() error {
	var err error

	CcsAwsSession.once.Do(func() {
		awsProfile := viper.GetString(config.AWSProfile)
		awsAccessKey := viper.GetString(config.AWSAccessKey)
		awsSecretAccessKey := viper.GetString(config.AWSSecretAccessKey)

		var opts []func(*awsconfig.LoadOptions) error
		opts = append(opts, awsconfig.WithRegion(viper.GetString(config.AWSRegion)))

		if awsProfile != "" {
			opts = append(opts, awsconfig.WithSharedConfigProfile(awsProfile))
		} else if awsAccessKey != "" || awsSecretAccessKey != "" {
			opts = append(opts, awsconfig.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(awsAccessKey, awsSecretAccessKey, ""),
			))
		}

		CcsAwsSession.cfg, err = awsconfig.LoadDefaultConfig(context.Background(), opts...)
		if err != nil {
			log.Printf("error initializing AWS config: %v", err)
			return
		}
		CcsAwsSession.iam = iamv2.NewFromConfig(CcsAwsSession.cfg)
		CcsAwsSession.s3 = s3v2.NewFromConfig(CcsAwsSession.cfg)
		CcsAwsSession.ec2 = ec2v2.NewFromConfig(CcsAwsSession.cfg)
		CcsAwsSession.accountId = viper.GetString(config.AWSAccountId)
	})

	return err
}

func (CcsAwsSession *ccsAwsSession) GetConfig() (aws.Config, error) {
	err := CcsAwsSession.GetAWSSessions()
	return CcsAwsSession.cfg, err
}

// GetCredentials returns the credentials for the current aws session.
func (CcsAwsSession *ccsAwsSession) GetCredentials(ctx context.Context) (*aws.Credentials, error) {
	if err := CcsAwsSession.GetAWSSessions(); err != nil {
		return nil, fmt.Errorf("failed to create aws session to retrieve credentials: %v", err)
	}

	creds, err := CcsAwsSession.cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get aws credentials: %v", err)
	}
	return &creds, nil
}

// GetRegion returns the region set when the session was created.
func (CcsAwsSession *ccsAwsSession) GetRegion() string {
	return CcsAwsSession.cfg.Region
}

// GetAccountId returns the aws account id in session.
func (CcsAwsSession *ccsAwsSession) GetAccountId() string {
	return CcsAwsSession.accountId
}
