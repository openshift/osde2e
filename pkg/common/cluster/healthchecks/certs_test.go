package healthchecks

import (
	"testing"

	"github.com/openshift/osde2e/pkg/common/util"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubernetes "k8s.io/client-go/kubernetes/fake"
)

func secrets(numItems int) []v1.Secret {
	secrets := []v1.Secret{}
	for i := 0; i < numItems; i++ {
		secrets = append(secrets, v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Labels:    map[string]string{"certificate_request": ""},
				Name:      util.RandomStr(5),
				Namespace: "openshift-config",
			},
		})
	}
	return secrets
}

func secretList(numItems int) *v1.SecretList {
	return &v1.SecretList{
		Items: secrets(numItems),
	}
}

func TestCerts(t *testing.T) {
	tests := []struct {
		description string
		expected    bool
		objs        []runtime.Object
	}{
		{
			description: "no certs present",
			expected:    false,
			objs:        []runtime.Object{secretList(0)},
		},
		{
			description: "one cert present",
			expected:    true,
			objs:        []runtime.Object{secretList(1)},
		},
		{
			description: "two certs present",
			expected:    true,
			objs:        []runtime.Object{secretList(2)},
		},
	}

	for _, test := range tests {
		kubeClient := kubernetes.NewSimpleClientset(test.objs...)
		state, err := CheckCerts(kubeClient.CoreV1(), nil)
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
			return
		}

		if state != test.expected {
			t.Errorf("%v: Expected value doesn't match returned value (%v, %v)", test.description, test.expected, state)
		}
	}
}
