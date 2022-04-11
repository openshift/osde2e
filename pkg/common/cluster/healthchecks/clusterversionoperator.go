package healthchecks

import (
	"context"
	"fmt"
	"log"

	v1 "github.com/openshift/api/config/v1"
	configclient "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	"github.com/openshift/osde2e/pkg/common/logging"
	"github.com/openshift/osde2e/pkg/common/metadata"
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

	var metadataState []string

	for _, v := range cvInfo.Status.Conditions {
		switch v.Type {
		case "Available":
			// Available must be true
			if v.Status == "True" {
				continue
			}
		case "Upgradeable", "RetrievedUpdates", "ReleaseAccepted":
			// These conditions don't matter to readiness state
			continue
		default:
			// Ignore any condition that isn't true
			// Any other true condition will return an error
			if v.Status != "True" {
				continue
			}
		}
		metadataState = append(metadataState, fmt.Sprintf("%v", v))
		logger.Printf("CVO State not complete: %v: %v %v", v.Type, v.Status, v.Message)
		success = false
	}

	if len(metadataState) > 0 {
		metadata.Instance.SetHealthcheckValue("cvo", metadataState)
	} else {
		metadata.Instance.ClearHealthcheckValue("cvo")
	}

	return success, nil
}
