package healthchecks

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/logging"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
)

// CheckReplicaCountForDaemonSets checks if all the daemonsets running on the cluster have expected replicas
func CheckReplicaCountForDaemonSets(dsClient appsv1.AppsV1Interface, logger *log.Logger) (bool, error) {
	allErrors := &multierror.Error{}
	helper := helper.NewOutsideGinkgo()
	logger = logging.CreateNewStdLoggerOrUseExistingLogger(logger)
	logger.Print("Checking that all Daemonsets are running with expected replicas...")

	dsList, err := dsClient.DaemonSets(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return false, multierror.Append(allErrors, err)
	}

	dsTotalCount := len(dsList.Items)
	if dsTotalCount != 0 {
		for _, ds := range dsList.Items {
			// Ignore daemonsets in the OSDE2E project
			if helper != nil && ds.Namespace == helper.CurrentProject() {
				continue
			}
			// Ignore daemonsets not managed by OSD
			if !strings.HasPrefix(ds.Namespace, "openshift-") {
				continue
			}
			if ds.Status.NumberReady != ds.Status.DesiredNumberScheduled {
				err = fmt.Errorf("daemonset %s has %d out of %d replicas ready", ds.Name, ds.Status.NumberReady, ds.Status.DesiredNumberScheduled)
				allErrors = multierror.Append(allErrors, err)
			}
		}
	} else {
		err = fmt.Errorf("there are no daemonsets running on the cluster, the cluster is not running well")
		return false, multierror.Append(allErrors, err)
	}

	return allErrors.ErrorOrNil() == nil, allErrors.ErrorOrNil()
}

// CheckReplicaCountForReplicaSets checks if all the replicasets running on the cluster have expected replicas
func CheckReplicaCountForReplicaSets(dsClient appsv1.AppsV1Interface, logger *log.Logger) (bool, error) {
	helper := helper.NewOutsideGinkgo()
	allErrors := &multierror.Error{}
	logger = logging.CreateNewStdLoggerOrUseExistingLogger(logger)
	logger.Print("Checking that all Replicasets are running with expected replicas...")

	rsList, err := dsClient.ReplicaSets(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return false, multierror.Append(allErrors, err)
	}

	rsTotalCount := len(rsList.Items)
	if rsTotalCount != 0 {
		for _, rs := range rsList.Items {
			// Ignore replicasets in the OSDE2E project
			if helper != nil && rs.Namespace == helper.CurrentProject() {
				continue
			}
			// Ignore replicasets not managed by OSD
			if !strings.HasPrefix(rs.Namespace, "openshift-") {
				continue
			}
			if rs.Status.ReadyReplicas != rs.Status.Replicas {
				err = fmt.Errorf("replicaset %s has %d out of %d replicas ready", rs.Name, rs.Status.ReadyReplicas, rs.Status.Replicas)
				allErrors = multierror.Append(allErrors, err)
			}
		}
	} else {
		err = fmt.Errorf("there are no replicasets running on the cluster, the cluster is not running well")
		return false, multierror.Append(allErrors, err)
	}

	return allErrors.ErrorOrNil() == nil, allErrors.ErrorOrNil()
}
