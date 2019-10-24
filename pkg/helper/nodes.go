package helper

import (
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CheckNodeHealth attempts to look at the state of all operator and returns true if things are healthy.
func CheckNodeHealth(kubeclient kubernetes.Interface) (bool, error) {
	fail := false
	log.Print("Checking that all Nodes are running or completed...")

	nodeClient, getOpts := kubeclient.CoreV1(), metav1.ListOptions{}
	list, err := nodeClient.Nodes().List(getOpts)
	if err != nil {
		log.Printf("Error getting CVS: %v\n", err)
		return false, nil
	}

	for _, node := range list.Items {
		for _, ns := range node.Status.Conditions {
			if ns.Type != "Ready" && ns.Status == "True" {
				log.Printf("Node (%v) issue: %v=%v %v\n", node.ObjectMeta.Name, ns.Type, ns.Status, ns.Message)
				fail = true
			} else if ns.Type == "Ready" && ns.Status != "True" {
				log.Printf("Node (%v) not ready: %v=%v %v\n", node.ObjectMeta.Name, ns.Type, ns.Status, ns.Message)
				fail = true
			}
		}
	}

	return fail, nil
}
