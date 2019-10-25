package helper

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
		if v.Type == "Available" && v.Status != "True" {
			log.Printf("API Server troubles: %v\n", v.Message)
			success = false
		}
		if v.Type == "Progressing" && v.Status != "False" {
			log.Printf("Installation or upgrade in progress: %v\n", v.Message)
			success = false
		}
		if v.Type == "Failing" && v.Status != "False" {
			log.Printf("Cluster is in a failed state: %v\n", v.Message)
			success = false
		}
	}

	return success, nil
}
