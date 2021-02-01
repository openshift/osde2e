package healthchecks

import (
	"context"
	"fmt"
	"log"

	machineapi "github.com/openshift/cluster-api/pkg/apis/machine/v1beta1"
	"github.com/openshift/osde2e/pkg/common/logging"
	"github.com/openshift/osde2e/pkg/common/metadata"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

const (
	runningPhase      = "Running"
	machinesNamespace = "openshift-machine-api"
)

// CheckMachinesObjectState lists all openshift machines and validates that they are "Running"
func CheckMachinesObjectState(dynamicClient dynamic.Interface, logger *log.Logger) (bool, error) {
	logger = logging.CreateNewStdLoggerOrUseExistingLogger(logger)

	logger.Print("Checking that machines are healthy...")

	mc := dynamicClient.
		Resource(schema.GroupVersionResource{Group: "machine.openshift.io", Resource: "machines", Version: "v1beta1"}).
		Namespace(machinesNamespace)
	obj, err := mc.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return false, err
	}
	if len(obj.Items) == 0 {
		return false, fmt.Errorf("No machines found in the %s namespace", machinesNamespace)
	}

	var metadataState []string

	for _, item := range obj.Items {
		var machine machineapi.Machine
		err = runtime.DefaultUnstructuredConverter.
			FromUnstructured(item.UnstructuredContent(), &machine)
		if err != nil {
			return false, fmt.Errorf("Error casting object: %s", err.Error())
		}

		if machine.Status.Phase == nil || *machine.Status.Phase != runningPhase {
			metadataState = append(metadataState, fmt.Sprintf("%v", machine))
			logger.Printf("machine %s not ready", machine.Name)
		}
	}
	if len(metadataState) > 0 {
		metadata.Instance.SetHealthcheckValue("machines", metadataState)
		return false, nil
	}

	metadata.Instance.ClearHealthcheckValue("machines")
	return true, nil
}
