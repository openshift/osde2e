package healthchecks

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/openshift/osde2e/pkg/common/logging"
	"github.com/openshift/osde2e/pkg/common/metadata"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type certCheckData struct {
	checkStarted bool
	startTime    time.Time
	certFound    bool
}

var certCheck = certCheckData{
	checkStarted: false,
	certFound:    false,
}

// CheckCerts will check for the presence of a cert issued by certman
func CheckCerts(secretClient v1.CoreV1Interface, logger *log.Logger) (bool, error) {
	logger = logging.CreateNewStdLoggerOrUseExistingLogger(logger)

	if !certCheck.checkStarted {
		certCheck.checkStarted = true
		certCheck.startTime = time.Now()
	}

	listOpts := metav1.ListOptions{
		LabelSelector: "certificate_request",
	}
	secrets, err := secretClient.Secrets("openshift-config").List(context.TODO(), listOpts)
	if err != nil {
		metadata.Instance.SetHealthcheckValue("certs", []string{"error"})
		return false, fmt.Errorf("error trying to find issued certificate(s): %v", err)
	}
	if len(secrets.Items) < 1 {
		logger.Printf("Certificate(s) not yet issued.")
		metadata.Instance.SetHealthcheckValue("certs", []string{"pending"})
		return false, nil
	}

	if !certCheck.certFound {
		certCheck.certFound = true
		metadata.Instance.SetTimeToCertificateIssued(time.Since(certCheck.startTime).Seconds())
		metadata.Instance.ClearHealthcheckValue("certs")
	}

	logger.Printf("Certificate(s) has been found.")

	return true, nil
}
