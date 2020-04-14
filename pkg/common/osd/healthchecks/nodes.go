package healthchecks

import (
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// CheckNodeHealth attempts to look at the state of all operator and returns true if things are healthy.
func CheckNodeHealth(nodeClient v1.CoreV1Interface) (bool, error) {
	success := true
	log.Print("Checking that all Nodes are running or completed...")

	listOpts := metav1.ListOptions{}
	list, err := nodeClient.Nodes().List(listOpts)
	if err != nil {
		log.Printf("Error getting CVS: %v\n", err)
		return false, nil
	}

	if len(list.Items) == 0 {
		log.Printf("Zero nodes found...?")
		return false, nil
	}

	for _, node := range list.Items {
		for _, ns := range node.Status.Conditions {
			if ns.Type != "Ready" && ns.Status == "True" {
				log.Printf("Node (%v) issue: %v=%v %v\n", node.ObjectMeta.Name, ns.Type, ns.Status, ns.Message)
				success = false
			} else if ns.Type == "Ready" && ns.Status != "True" {
				log.Printf("Node (%v) not ready: %v=%v %v\n", node.ObjectMeta.Name, ns.Type, ns.Status, ns.Message)
				success = false
			}
		}
	}

	return success, nil
}
