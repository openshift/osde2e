package healthchecks

import (
	"log"

	configclient "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CheckCVOReadiness attempts to look at the state of the ClusterVersionOperator and returns true if things are healthy.
func CheckCVOReadiness(configClient configclient.ConfigV1Interface) (bool, error) {
	success := true
	log.Print("Checking that CVO says the cluster is healthy...")

	getOpts := metav1.GetOptions{}
	cvInfo, err := configClient.ClusterVersions().Get("version", getOpts)
	if err != nil {
		log.Printf("Error getting CVS: %v\n", err)
		return false, nil
	}

	for _, v := range cvInfo.Status.Conditions {
		if (v.Type != "Available" && v.Status != "False") && v.Type != "Upgradeable" && v.Type != "RetrievedUpdates" {
			log.Printf("CVO State not complete: %v: %v %v", v.Type, v.Status, v.Message)
			success = false
		}
	}

	return success, nil
}
