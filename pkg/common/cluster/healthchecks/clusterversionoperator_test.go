package healthchecks

import (
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	fakeConfig "github.com/openshift/client-go/config/clientset/versioned/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func clusterVersion() *configv1.ClusterVersion {
	return &configv1.ClusterVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name: "version",
		},
		Status: configv1.ClusterVersionStatus{
			Conditions: []configv1.ClusterOperatorStatusCondition{
				{
					Type:    configv1.OperatorAvailable,
					Status:  configv1.ConditionTrue,
					Reason:  "Available",
					Message: "Available",
				},
				{
					Type:    configv1.OperatorProgressing,
					Status:  configv1.ConditionFalse,
					Reason:  "Available",
					Message: "Available",
				},
				{
					Type:    configv1.OperatorDegraded,
					Status:  configv1.ConditionFalse,
					Reason:  "Available",
					Message: "Available",
				},
				{
					Type:    configv1.OperatorUpgradeable,
					Status:  configv1.ConditionTrue,
					Reason:  "Available",
					Message: "Available",
				},
				{
					Type:    configv1.RetrievedUpdates,
					Status:  configv1.ConditionTrue,
					Reason:  "Available",
					Message: "Available",
				},
				{
					Type:    "ReleaseAccepted",
					Status:  configv1.ConditionTrue,
					Reason:  "Available",
					Message: "Available",
				},
			},
		},
	}
}

func unavailableClusterVersion() *configv1.ClusterVersion {
	op := clusterVersion()
	op.Status.Conditions[0].Status = configv1.ConditionFalse
	op.Status.Conditions[2].Status = configv1.ConditionTrue
	op.Status.Conditions[0].Message = "Degraded"
	op.Status.Conditions[1].Message = "Degraded"
	op.Status.Conditions[2].Message = "Degraded"
	return op
}

func progressingClusterVersion() *configv1.ClusterVersion {
	op := clusterVersion()
	op.Status.Conditions[0].Status = configv1.ConditionTrue
	op.Status.Conditions[1].Status = configv1.ConditionTrue
	op.Status.Conditions[0].Message = "Available"
	op.Status.Conditions[1].Message = "Progressing"
	return op
}

func TestCheckCVOReadiness(t *testing.T) {
	tests := []struct {
		description   string
		expected      bool
		expectedError bool
		objs          []runtime.Object
	}{
		{"no version", false, true, nil},
		{"single version success", true, false, []runtime.Object{clusterVersion()}},
		{"single version failure", false, false, []runtime.Object{unavailableClusterVersion()}},
		{"single version progressing", false, false, []runtime.Object{progressingClusterVersion()}},
	}

	for _, test := range tests {
		cfgClient := fakeConfig.NewSimpleClientset(test.objs...)
		state, err := CheckCVOReadiness(cfgClient.ConfigV1(), nil)

		if err != nil && !test.expectedError {
			t.Errorf("Unexpected error: %s", err)
			return
		}

		if state != test.expected {
			t.Errorf("%v: Expected value doesn't match returned value (%v, %v)", test.description, test.expected, state)
		}
	}
}
