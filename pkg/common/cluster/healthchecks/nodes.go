package healthchecks

import (
	"context"
	"fmt"
	"log"

	"github.com/openshift/osde2e/pkg/common/logging"
	"github.com/openshift/osde2e/pkg/common/metadata"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// CheckNodeHealth attempts to look at the state of all operator and returns true if things are healthy.
func CheckNodeHealth(nodeClient v1.CoreV1Interface, logger *log.Logger) (bool, error) {
	logger = logging.CreateNewStdLoggerOrUseExistingLogger(logger)

	success := true
	logger.Print("Checking that all Nodes are running or completed...")

	listOpts := metav1.ListOptions{}
	list, err := nodeClient.Nodes().List(context.TODO(), listOpts)
	if err != nil {
		return false, fmt.Errorf("error getting node list: %v", err)
	}

	if len(list.Items) == 0 {
		return false, fmt.Errorf("no nodes found")
	}

	var metadataState []string

	for _, node := range list.Items {
		for _, ns := range node.Status.Conditions {
			if ns.Type != "Ready" && ns.Status == "True" {
				metadataState = append(metadataState, fmt.Sprintf("%v", node))
				logger.Printf("Node (%v) issue: %v=%v %v\n", node.ObjectMeta.Name, ns.Type, ns.Status, ns.Message)
				success = false
			} else if ns.Type == "Ready" && ns.Status != "True" {
				metadataState = append(metadataState, fmt.Sprintf("%v", node))
				logger.Printf("Node (%v) not ready: %v=%v %v\n", node.ObjectMeta.Name, ns.Type, ns.Status, ns.Message)
				success = false
			}
		}
		// Check taints to ensure node is schedulable
		for _, nt := range node.Spec.Taints {
			if nt.Effect == "NoSchedule" && nt.Key == "node.kubernetes.io/unschedulable" {
				metadataState = append(metadataState, fmt.Sprintf("%v", node))
				logger.Printf("Node (%v) not ready taint: %v=%v\n", node.ObjectMeta.Name, nt.Key, nt.Effect)
				success = false
			}
		}
	}

	if len(metadataState) > 0 {
		metadata.Instance.SetHealthcheckValue("nodes", metadataState)
	} else {
		metadata.Instance.ClearHealthcheckValue("nodes")
	}

	return success, nil
}
