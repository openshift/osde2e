package helper

import (
	"log"

	configclient "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CheckOperatorReadiness attempts to look at the state of all operator and returns true if things are healthy.
func CheckOperatorReadiness(configClient configclient.ConfigV1Interface) (bool, error) {
	success := true
	log.Print("Checking that all Operators are running or completed...")

	listOpts := metav1.ListOptions{}
	list, err := configClient.ClusterOperators().List(listOpts)
	if err != nil {
		log.Printf("Error getting CVS: %v\n", err)
		return false, nil
	}

	if len(list.Items) == 0 {
		log.Printf("No operators found...?")
		return false, nil
	}

	for _, co := range list.Items {
		for _, cos := range co.Status.Conditions {
			if (cos.Type != "Available" && cos.Status != "False") && cos.Type != "Upgradeable" {
				log.Printf("Operator %v type %v is %v: %v", co.ObjectMeta.Name, cos.Type, cos.Status, cos.Message)
				success = false
			}
		}
	}

	return success, nil
}
