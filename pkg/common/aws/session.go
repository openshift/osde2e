package aws

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/prometheus/common/log"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
)

const (
	metricsAWSAccessKeyID     = "metrics.awsAccessKeyId"
	metricsAWSSecretAccessKey = "metrics.awsSecretAccessKey"
	metricsAWSRegion          = "metrics.awsRegion"

	metricsAWSAccessKeyIDEnv     = "METRICS_AWS_ACCESS_KEY_ID"
	metricsAWSSecretAccessKeyEnv = "METRICS_AWS_SECRET_ACCESS_KEY"
	metricsAWSRegionEnv          = "METRICS_AWS_REGION"
)

type awsSession struct {
	session *session.Session
	once    sync.Once
}

// AWSSession is the global AWS session for interacting with S3.
var AWSSession awsSession

func init() {
	viper.BindEnv(metricsAWSAccessKeyID, metricsAWSAccessKeyIDEnv)
	config.RegisterSecret(metricsAWSAccessKeyID, "metrics-aws-access-key")

	viper.BindEnv(metricsAWSSecretAccessKey, metricsAWSSecretAccessKeyEnv)
	config.RegisterSecret(metricsAWSSecretAccessKey, "metrics-aws-secret-access-key")

	viper.BindEnv(metricsAWSRegion, metricsAWSRegionEnv)
	config.RegisterSecret(metricsAWSRegion, "metrics-aws-region")
}

func (a *awsSession) getSession() (*session.Session, error) {
	var err error

	// Initialize this once, and initialize it in getSession so that osde2e capabilities that don't use S3
	// don't have to worry about the noise or configuring S3 properly. The cost here is that things will fail late
	// on misconfiguration, but since we're the primary consumers of osde2e, at the moment this isn't a big deal.
	a.once.Do(func() {
		// We're using static credentials here so that we can use AWS credentials for cluster providers.
		// When we have more time, we should make this not metrics focused, as the intent of this library is to be purpose agnostic.
		a.session, err = session.NewSession(aws.NewConfig().
			WithCredentials(credentials.NewStaticCredentials(viper.GetString(metricsAWSAccessKeyID), viper.GetString(metricsAWSSecretAccessKey), "")).
			WithRegion(viper.GetString(metricsAWSRegion)))

		if err != nil {
			log.Errorf("error initializing AWS session: %v", err)
		}
	})

	if a.session == nil {
		err = fmt.Errorf("unable to initialize AWS session")
	}

	return a.session, err
}
