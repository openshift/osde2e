package verify

import (
	"io/ioutil"
	"os"
	"testing"
)

const (
	TestKubeconfigEnv = "TEST_KUBECONFIG"
)

func TestCluster(t *testing.T) {
	kubeconfigPath, ok := os.LookupEnv(TestKubeconfigEnv)
	if !ok {
		t.Skipf("Environment variable '%s' not set, skipping...", TestKubeconfigEnv)
		t.SkipNow()
	}

	data, err := ioutil.ReadFile(kubeconfigPath)
	if err != nil {
		t.Fatalf("Couldn't read kubeconfig: %v", err)
	}

	_, err = RunTests(data)
	if err != nil {
		t.Fatalf("Couldn't configure cluster: %v", err)
	}
}
