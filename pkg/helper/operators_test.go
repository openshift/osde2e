package helper

import (
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	fakeConfig "github.com/openshift/client-go/config/clientset/versioned/fake"
	"github.com/openshift/osde2e/pkg/config"
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

func TestCheckOperatorReadiness(t *testing.T) {
	var tests = []struct {
		description string
		expected    bool
		objs        []runtime.Object
		skip        string
	}{
		{"no operators", false, nil, ""},
		{"single operator success", true, []runtime.Object{clusterOperator("a")}, ""},
		{"single operator failure", false, []runtime.Object{unavailableClusterOperator("a")}, ""},
		{"single operator progressing", false, []runtime.Object{progressingClusterOperator("a")}, ""},
		{"multi operator success", true, []runtime.Object{clusterOperator("a"), clusterOperator("b")}, ""},
		{"multi operator one progressing", false, []runtime.Object{clusterOperator("a"), progressingClusterOperator("b")}, ""},
		{"multi operator one failure", false, []runtime.Object{clusterOperator("a"), unavailableClusterOperator("b")}, ""},
		{"multi operator, skip success", true, []runtime.Object{
			clusterOperator("a"),
			unavailableClusterOperator("b"),
			unavailableClusterOperator("c"),
			unavailableClusterOperator("d"),
			clusterOperator("e"),
		}, "b,c,d"},
		{"multi operator, skip failure", false, []runtime.Object{
			clusterOperator("a"),
			unavailableClusterOperator("b"),
			unavailableClusterOperator("c"),
			unavailableClusterOperator("d"),
		}, "b,c"},
	}

	for _, test := range tests {
		cfgClient := fakeConfig.NewSimpleClientset(test.objs...)
		c := config.Config{}
		c.Tests.OperatorSkip = test.skip
		state, err := CheckOperatorReadiness(&c, cfgClient.ConfigV1())

		if err != nil {
			t.Errorf("Unexpected error: %s", err)
			return
		}

		if state != test.expected {
			t.Errorf("%v: Expected value doesn't match returned value (%v, %v)", test.description, test.expected, state)
		}
	}
}
