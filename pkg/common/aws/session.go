package aws

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/prometheus/common/log"
)

type awsSession struct {
	session *session.Session
	once    sync.Once
}

// AWSSession is the global AWS session for interacting with S3.
var AWSSession awsSession

func (a *awsSession) getSession() (*session.Session, error) {
	var err error

	// Initialize this once, and initialize it in getSession so that osde2e capabilities that don't use S3
	// don't have to worry about the noise or configuring S3 properly. The cost here is that things will fail late
	// on misconfiguration, but since we're the primary consumers of osde2e, at the moment this isn't a big deal.
	a.once.Do(func() {
		// We're very intentionally using the shared configs here.
		// This allows us to configure the AWS client at a system level and this should behave as expected.
		// This is particularly useful if we want to, at some point in the future, run this on an AWS host with an instance profile
		// that doesn't need explicit credentials.
		a.session, err = session.NewSessionWithOptions(session.Options{SharedConfigState: session.SharedConfigEnable})

		if err != nil {
			log.Errorf("error initializing AWS session: %v", err)
		}
	})

	if a.session == nil {
		err = fmt.Errorf("unable to initialize AWS session")
	}

	return a.session, err
}
