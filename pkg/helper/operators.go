package helper

import (
	"log"

	osconfig "github.com/openshift/client-go/config/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CheckOperatorReadiness attempts to look at the state of all operator and returns true if things are healthy.
func CheckOperatorReadiness(oscfg *osconfig.Clientset) (bool, error) {
	fail := true
	log.Print("Checking that all Operators are running or completed...")

	configClient, getOpts := oscfg.ConfigV1(), metav1.ListOptions{}
	list, err := configClient.ClusterOperators().List(getOpts)
	if err != nil {
		log.Printf("Error getting CVS: %v\n", err)
		return false, nil
	}

	for _, co := range list.Items {
		for _, cos := range co.Status.Conditions {
			if (cos.Type != "Available" || cos.Type != "True") && cos.Type != "Upgradable" {
				log.Printf("Operator %v type %v is %v: %v", co.ObjectMeta.Name, cos.Type, cos.Status, cos.Message)
				fail = false
			}
		}
	}

	return fail, nil
}
