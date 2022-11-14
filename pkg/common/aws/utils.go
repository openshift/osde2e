package aws

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"k8s.io/apimachinery/pkg/util/wait"
)

type ccsAwsSession struct {
	session *session.Session
	iam     *iam.IAM
	once    sync.Once
}

// CCSAWSSession is the global AWS session for interacting with AWS.
var (
	CcsAwsSession ccsAwsSession
	ccsKeys       *iam.CreateAccessKeyOutput
)

// GetSession returns a new AWS session with the first AWS account in the config file. The session is cached for the rest of the program.
func (CcsAwsSession *ccsAwsSession) getIamClient() (*session.Session, *iam.IAM) {
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

	return CcsAwsSession.session, CcsAwsSession.iam
}

// AWS check for osdCCSAdmin credentials
func VerifyCCS() (string, error) {
	var err error
	CcsAwsSession.session, CcsAwsSession.iam = CcsAwsSession.getIamClient()

	result, err := CcsAwsSession.iam.GetUser(&iam.GetUserInput{})
	if err != nil {
		return "", err
	}

	if *result.User.UserName != "osdCcsAdmin" {
		log.Printf("The user %s is not osdCcsAdmin", *result.User.UserName)
	}

	return string(*result.User.UserName), nil
}

func CcsScale() (string, string, error) {
	err := wait.PollImmediate(2*time.Minute, 120*time.Minute, func() (bool, error) {
		// Grabs existing keys
		keys, err := CcsAwsSession.iam.ListAccessKeys(&iam.ListAccessKeysInput{
			UserName: aws.String("osdCcsAdmin"),
		})
		if err != nil {
			log.Printf("error listing keys: %v", err)
			return false, err
		}

		switch {
		case len(keys.AccessKeyMetadata) < 2:
			err = createCcsKeys()
			if err != nil {
				log.Printf("error creating keys: %v", err)
				return false, err
			} else {
				return true, nil
			}
		case len(keys.AccessKeyMetadata) == 2:
			for _, key := range keys.AccessKeyMetadata {
				// If the create date is older than 5 minutes, delete the key
				// This should be enough time for OCM to finish with old CCS keys used to create a cluster.
				if key.CreateDate.Before(time.Now().Add(-10 * time.Minute)) {
					_, err := CcsAwsSession.iam.DeleteAccessKey(&iam.DeleteAccessKeyInput{
						AccessKeyId: key.AccessKeyId,
						UserName:    aws.String("osdCcSAdmin"),
					})
					if err != nil {
						log.Printf("error deleting key: %v", err)
						return false, nil
					} else {
						log.Printf("Deleted old key pair for osdCcsAdmin")
						err = createCcsKeys()
						if err != nil {
							log.Printf("error creating keys: %v", err)
							return false, err
						} else {
							return true, nil
						}
					}
				} else {
					log.Printf("Existing key pair for osdCcsAdmin is not safe to delete")
				}
			}
		}
		return false, nil
	})
	if err != nil {
		return "", "", err
	}

	return *ccsKeys.AccessKey.AccessKeyId, *ccsKeys.AccessKey.SecretAccessKey, err
}

func createCcsKeys() error {
	var err error

	// Create new CCS key pair
	ccsKeys, err = CcsAwsSession.iam.CreateAccessKey(&iam.CreateAccessKeyInput{
		UserName: aws.String("osdCcSAdmin"),
	})
	if err != nil {
		return fmt.Errorf("error creating keys: %v", err)
	} else {
		log.Printf("Created new key pair for osdCcsAdmin")
	}
	return err
}
