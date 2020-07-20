package healthchecks

import (
	"context"
	"fmt"
	"log"

	v1 "github.com/openshift/api/config/v1"
	configclient "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	machineapi "github.com/openshift/cluster-api/pkg/apis/machine/v1beta1"
	"github.com/openshift/osde2e/pkg/common/logging"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// GetClusterVersionObject wlil get the cluster version object for the cluster.
func GetClusterVersionObject(configClient configclient.ConfigV1Interface) (*v1.ClusterVersion, error) {
	getOpts := metav1.GetOptions{}
	return configClient.ClusterVersions().Get(context.TODO(), "version", getOpts)
}

// CheckCVOReadiness attempts to look at the state of the ClusterVersionOperator and returns true if things are healthy.
func CheckCVOReadiness(configClient configclient.ConfigV1Interface, logger *log.Logger) (bool, error) {
	logger = logging.CreateNewStdLoggerOrUseExistingLogger(logger)

	success := true
	logger.Print("Checking that CVO says the cluster is healthy...")

	cvInfo, err := GetClusterVersionObject(configClient)
	if err != nil {
		return false, err
	}

	for _, v := range cvInfo.Status.Conditions {
		if (v.Type != "Available" && v.Status != "False") && v.Type != "Upgradeable" && v.Type != "RetrievedUpdates" {
			logger.Printf("CVO State not complete: %v: %v %v", v.Type, v.Status, v.Message)
			success = false
		}
	}

	return success, nil
}

// CheckMachinesObjectState lists all openshift machines and validates that they are "Running"
func CheckMachinesObjectState(dynamicClient dynamic.Interface, logger *log.Logger) (bool, error) {
	mc := dynamicClient.Resource(schema.GroupVersionResource{Group: "machine.openshift.io", Resource: "machines", Version: "v1beta1"})
	obj, err := mc.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return false, err
	}
	var runningPhase string
	runningPhase = "Running"

	for _, item := range obj.Items {
		var machine machineapi.Machine
		err = runtime.DefaultUnstructuredConverter.
			FromUnstructured(item.UnstructuredContent(), &machine)
		if err != nil {
			return false, fmt.Errorf("Error casting object: %s", err.Error())
		}

		if machine.Status.Phase != &runningPhase {
			logger.Printf("machine not ready")
			return false, nil
		}
	}
	return true, nil
}
