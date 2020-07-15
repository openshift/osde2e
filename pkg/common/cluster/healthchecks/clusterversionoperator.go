package healthchecks

import (
	"context"
	"log"

	v1 "github.com/openshift/api/config/v1"
	configclient "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	"github.com/openshift/osde2e/pkg/common/logging"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetClusterVersionObject wlil get the cluster version object for the cluster.
func GetClusterVersionObject(configClient configclient.ConfigV1Interface) (*v1.ClusterVersion, error) {
	getOpts := metav1.GetOptions{}
	return configClient.ClusterVersions().Get(context.TODO(), "version", getOpts)
}

// CheckCVOReadiness attempts to look at the state of the ClusterVersionOperator and returns true if things are healthy.
func CheckCVOReadiness(configClient configclient.ConfigV1Interface, logger *log.Logger) (bool, error) {
	logger = logging.CreateNewStdLoggerOrUseExistingLogger(logger)

	success := true
	logger.Print("Checking that CVO says the cluster is healthy...")

	cvInfo, err := GetClusterVersionObject(configClient)
	if err != nil {
		return false, err
	}

	for _, v := range cvInfo.Status.Conditions {
		if (v.Type != "Available" && v.Status != "False") && v.Type != "Upgradeable" && v.Type != "RetrievedUpdates" {
			logger.Printf("CVO State not complete: %v: %v %v", v.Type, v.Status, v.Message)
			success = false
		}
	}

	return success, nil
}
