package healthchecks

import (
	"context"
	"fmt"
	"log"

	"github.com/openshift/osde2e/pkg/common/logging"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
)

//CheckReplicaCountForDaemonSets checks if all the daemonsets running on the cluster have expected replicas
func CheckReplicaCountForDaemonSets(dsClient appsv1.AppsV1Interface, logger *log.Logger) (bool, error) {
	logger = logging.CreateNewStdLoggerOrUseExistingLogger(logger)
	logger.Print("Checking that all Daemonsets are running with expected replicas...")

	dsList, err := dsClient.DaemonSets(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return false, err
	}

	dsTotalCount := len(dsList.Items)
	dsReadyCount := 0
	if dsTotalCount != 0 {
		for _, ds := range dsList.Items {
			if ds.Status.NumberReady == ds.Status.DesiredNumberScheduled {
				dsReadyCount = dsReadyCount + 1
			}
		}
	} else {
		return false, fmt.Errorf("there is no daemonset running on the cluster, the cluster is not running well")
	}

	if dsTotalCount != dsReadyCount {
		return false, fmt.Errorf("the number of total and ready daemonset replicas are different, some of the replicas are unhealthy")
	}

	return true, nil
}

//CheckReplicaCountForReplicaSets checks if all the replicasets running on the cluster have expected replicas
func CheckReplicaCountForReplicaSets(dsClient appsv1.AppsV1Interface, logger *log.Logger) (bool, error) {
	logger = logging.CreateNewStdLoggerOrUseExistingLogger(logger)
	logger.Print("Checking that all Replicasets are running with expected replicas...")

	rsList, err := dsClient.ReplicaSets(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return false, err
	}

	rsTotalCount := len(rsList.Items)
	rsReadyCount := 0
	if rsTotalCount != 0 {
		for _, rs := range rsList.Items {
			if rs.Status.ReadyReplicas == rs.Status.Replicas {
				rsReadyCount = rsReadyCount + 1
			}
		}
	} else {
		return false, fmt.Errorf("there is no replicaset running on the cluster, the cluster is not running well")
	}

	if rsTotalCount != rsReadyCount {
		return false, fmt.Errorf("the number of total and ready replicaset replicas are different, some of the replicas are unhealthy")
	}

	return true, nil
}
