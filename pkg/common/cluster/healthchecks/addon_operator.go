package healthchecks

import (
	"context"
	"fmt"
	"log"

	addonsv1alpha1 "github.com/openshift/addon-operator/apis/addons/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// CheckAddonOperatorReadiness reports the state
// of the AddonOperators "Available" condition.
func CheckAddonOperatorReadiness(
	runtimeClient runtimeclient.Client, logger *log.Logger,
) (bool, error) {
	ao := &addonsv1alpha1.AddonOperator{}
	err := runtimeClient.Get(context.TODO(), runtimeclient.ObjectKey{
		Name: addonsv1alpha1.DefaultAddonOperatorName,
	}, ao)
	if errors.IsNotFound(err) {
		// Waiting for object to be created.
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("error getting AddonOperator object: %w", err)
	}

	return meta.IsStatusConditionTrue(
		ao.Status.Conditions,
		addonsv1alpha1.Available,
	), nil
}
