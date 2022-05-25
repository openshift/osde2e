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
var ccsKeys *iam.CreateAccessKeyOutput

//GetSession returns a new AWS session with the first AWS account in the config file. The session is cached for the rest of the program.
func (CcsAwsSession *ccsAwsSession) getIamClient() (*session.Session, *iam.IAM, *ec2.EC2) {
	var err error

	CcsAwsSession.once.Do(func() {
		CcsAwsSession.session, err = session.NewSession(aws.NewConfig().
			WithCredentials(credentials.NewStaticCredentials(viper.GetString("ocm.aws.accessKey"), viper.GetString("ocm.aws.secretKey"), "")).
			WithRegion(viper.GetString(config.CloudProvider.Region)))
		CcsAwsSession.iam = iam.New(CcsAwsSession.session)
	})
	if err != nil {
		log.Printf("error initializing AWS session: %v", err)
	}

	return CcsAwsSession.session, CcsAwsSession.iam, CcsAwsSession.ec2
}

// AWS check for osdCCSAdmin credentials
func VerifyCCS() (string, error) {
	var err error
	CcsAwsSession.session, CcsAwsSession.iam, CcsAwsSession.ec2 = CcsAwsSession.getIamClient()

	result, err := CcsAwsSession.iam.GetUser(&iam.GetUserInput{})
	if err != nil {
		return "", err
	}

	if *result.User.UserName != "osdCcsAdmin" {
		log.Printf("The user %s is not osdCcsAdmin", *result.User.UserName)
	}

	return string(*result.User.UserName), nil
}
