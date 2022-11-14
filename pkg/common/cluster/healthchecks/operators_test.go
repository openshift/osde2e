package healthchecks

import (
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	fakeConfig "github.com/openshift/client-go/config/clientset/versioned/fake"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func clusterOperator(name string) *configv1.ClusterOperator {
	return &configv1.ClusterOperator{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Status: configv1.ClusterOperatorStatus{
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
			},
		},
	}
}

func unavailableClusterOperator(name string) *configv1.ClusterOperator {
	op := clusterOperator(name)
	op.Status.Conditions[0].Status = configv1.ConditionFalse
	op.Status.Conditions[2].Status = configv1.ConditionTrue
	op.Status.Conditions[0].Message = "Degraded"
	op.Status.Conditions[1].Message = "Degraded"
	op.Status.Conditions[2].Message = "Degraded"
	return op
}

func progressingClusterOperator(name string) *configv1.ClusterOperator {
	op := clusterOperator(name)
	op.Status.Conditions[0].Status = configv1.ConditionTrue
	op.Status.Conditions[1].Status = configv1.ConditionTrue
	op.Status.Conditions[0].Message = "Available"
	op.Status.Conditions[1].Message = "Progressing"
	return op
}

func clusterOperatorWithUnknownStatus(name string) *configv1.ClusterOperator {
	op := clusterOperator(name)
	op.Status.Conditions[0].Status = configv1.ConditionFalse
	op.Status.Conditions[1].Status = configv1.ConditionUnknown
	op.Status.Conditions[0].Message = "Fake Condition"
	op.Status.Conditions[1].Message = "RecentBackup"
	return op
}

func TestCheckOperatorReadiness(t *testing.T) {
	tests := []struct {
		description   string
		expected      bool
		objs          []runtime.Object
		expectedError bool
		skip          string
	}{
		{"no operators", false, nil, true, ""},
		{"single operator success", true, []runtime.Object{clusterOperator("a")}, false, ""},
		{"single operator failure", false, []runtime.Object{unavailableClusterOperator("a")}, false, ""},
		{"single operator progressing", false, []runtime.Object{progressingClusterOperator("a")}, false, ""},
		{"multi operator success", true, []runtime.Object{clusterOperator("a"), clusterOperator("b")}, false, ""},
		{"multi operator one progressing", false, []runtime.Object{clusterOperator("a"), progressingClusterOperator("b")}, false, ""},
		{"multi operator one with condition status unknown", false, []runtime.Object{clusterOperator("a"), clusterOperatorWithUnknownStatus("b")}, false, ""},
		{"multi operator one failure", false, []runtime.Object{clusterOperator("a"), unavailableClusterOperator("b")}, false, ""},
		{"multi operator, skip success", true, []runtime.Object{
			clusterOperator("a"),
			unavailableClusterOperator("b"),
			unavailableClusterOperator("c"),
			unavailableClusterOperator("d"),
			clusterOperator("e"),
		}, false, "b,c,d"},
		{"multi operator, skip failure", false, []runtime.Object{
			clusterOperator("a"),
			unavailableClusterOperator("b"),
			unavailableClusterOperator("c"),
			unavailableClusterOperator("d"),
		}, false, "b,c"},
	}

	for _, test := range tests {
		viper.Reset()
		cfgClient := fakeConfig.NewSimpleClientset(test.objs...)
		viper.Set(config.Tests.OperatorSkip, test.skip)
		state, err := CheckOperatorReadiness(cfgClient.ConfigV1(), nil)

		if err != nil && !test.expectedError {
			t.Errorf("Unexpected error: %s", err)
			return
		}

		if state != test.expected {
			t.Errorf("%v: Expected value doesn't match returned value (%v, %v)", test.description, test.expected, state)
		}
	}
}
