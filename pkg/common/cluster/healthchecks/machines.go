package healthchecks

import (
	"context"
	"fmt"
	"log"

	machineapi "github.com/openshift/cluster-api/pkg/apis/machine/v1beta1"
	"github.com/openshift/osde2e/pkg/common/logging"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// CheckMachinesObjectState lists all openshift machines and validates that they are "Running"
func CheckMachinesObjectState(dynamicClient dynamic.Interface, logger *log.Logger) (bool, error) {
	logger = logging.CreateNewStdLoggerOrUseExistingLogger(logger)

	logger.Print("Checking that machines are healthy...")

	mc := dynamicClient.Resource(schema.GroupVersionResource{Group: "machine.openshift.io", Resource: "machines", Version: "v1beta1"})
	obj, err := mc.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return false, err
	}
	var runningPhase string = "Running"

	for _, item := range obj.Items {
		var machine machineapi.Machine
		err = runtime.DefaultUnstructuredConverter.
			FromUnstructured(item.UnstructuredContent(), &machine)
		if err != nil {
			return false, fmt.Errorf("Error casting object: %s", err.Error())
		}

		if machine.Status.Phase != nil && *machine.Status.Phase != runningPhase {
			logger.Printf("machine %s not ready", machine.Name)
			return false, nil
		}
	}
	return true, nil
}
