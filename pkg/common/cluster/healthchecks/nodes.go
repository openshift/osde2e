package healthchecks

import (
	"context"
	"fmt"
	"log"

	"github.com/openshift/osde2e-common/pkg/clients/openshift"
	"github.com/openshift/osde2e/pkg/common/logging"
	corev1 "k8s.io/api/core/v1"
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
		if openshift.IsRetryableAPIError(err) {
			logger.Printf("error listing nodes: %v", err)
			return false, nil
		}
		return false, fmt.Errorf("error getting node list: %v", err)
	}

	if len(list.Items) == 0 {
		return false, fmt.Errorf("no nodes found")
	}

	for _, node := range list.Items {
		for _, ns := range node.Status.Conditions {
			if ns.Type != corev1.NodeReady && ns.Status == corev1.ConditionTrue {
				logger.Printf("Node (%v) issue: %v=%v %v\n", node.Name, ns.Type, ns.Status, ns.Message)
				success = false
			} else if ns.Type == corev1.NodeReady && ns.Status != corev1.ConditionTrue {
				logger.Printf("Node (%v) not ready: %v=%v %v\n", node.Name, ns.Type, ns.Status, ns.Message)
				success = false
			}
		}
		// Check taints to ensure node is schedulable
		for _, nt := range node.Spec.Taints {
			if nt.Effect == corev1.TaintEffectNoSchedule && nt.Key == corev1.TaintNodeUnschedulable {
				logger.Printf("Node (%v) not schedulable with taint: %v=%v\n", node.Name, nt.Key, nt.Effect)
				success = false
			}
		}
	}

	return success, nil
}
