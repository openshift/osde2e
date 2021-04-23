package healthchecks

import (
	"context"
	"fmt"
	"log"
	"strings"

	configclient "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/logging"
	"github.com/openshift/osde2e/pkg/common/metadata"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CheckOperatorReadiness attempts to look at the state of all operator and returns true if things are healthy.
func CheckOperatorReadiness(configClient configclient.ConfigV1Interface, logger *log.Logger) (bool, error) {
	logger = logging.CreateNewStdLoggerOrUseExistingLogger(logger)

	success := true
	logger.Print("Checking that all Operators are running or completed...")

	listOpts := metav1.ListOptions{}
	list, err := configClient.ClusterOperators().List(context.TODO(), listOpts)
	if err != nil {
		return false, fmt.Errorf("error getting cluster operator list: %v", err)
	}

	if len(list.Items) == 0 {
		return false, fmt.Errorf("no operators were found")
	}

	// Load the list of operators we want to ignore and skip.
	operatorSkipString := viper.GetString(config.Tests.OperatorSkip)
	operatorSkipList := make(map[string]string)
	if len(operatorSkipString) > 0 {
		operatorSkipVals := strings.Split(operatorSkipString, ",")
		for _, val := range operatorSkipVals {
			operatorSkipList[val] = ""
		}
	}

	var metadataState []string

	for _, co := range list.Items {
		if _, ok := operatorSkipList[co.GetName()]; !ok {
			for _, cos := range co.Status.Conditions {
				if cos.Type == "Disabled" && cos.Status == "True" {
					continue
				}
				if (cos.Type != "Available" && cos.Status != "False") && cos.Type != "Upgradeable" {
					metadataState = append(metadataState, fmt.Sprintf("%v", co))
					logger.Printf("Operator %v type %v is %v: %v", co.ObjectMeta.Name, cos.Type, cos.Status, cos.Message)
					success = false
				}
			}
		}
	}

	if len(metadataState) > 0 {
		metadata.Instance.SetHealthcheckValue("operators", metadataState)
	} else {
		metadata.Instance.ClearHealthcheckValue("operators")
	}

	return success, nil
}
