package runner

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	image "github.com/openshift/client-go/image/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// TestKubeconfigVar is the environment variable that is used for a kubeconfig.
	TestKubeconfigVar = "TEST_KUBECONFIG"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func TestRunner(t *testing.T) {
	g := NewGomegaWithT(t)

	// setup runner
	runner := setupRunner(t)
	runner.Name = "runner-test"
	runner.Namespace = "default"

	// write series as results
	count := 15
	for i := 'A'; i < 'A'+int32(count); i++ {
		runner.Cmd += fmt.Sprintf(">%s/%c echo '%c' && ", runner.OutputDir, i, i)
	}

	// stdout and stderr should be just outMsg and errMsg
	outMsg := fmt.Sprintf("%s-out.txt", runner.Name)
	errMsg := fmt.Sprintf("%s-err.txt", runner.Name)
	runner.Cmd += fmt.Sprintf("echo '%s' && echo '%s' >&2", outMsg, errMsg)

	// execute runner
	stopCh := make(chan struct{})
	err := runner.Run(1800, stopCh)
	g.Expect(err).NotTo(HaveOccurred())

	// get results
	results, err := runner.RetrieveResults()
	g.Expect(err).NotTo(HaveOccurred())

	// verify total count and stdout/stderr
	g.Expect(results).To(HaveLen(count+2), "should have result for each count + stdout + stderr")
	g.Expect(results).To(HaveKey(outMsg))
	g.Expect(results[outMsg]).To(BeEquivalentTo(outMsg))
	g.Expect(results).To(HaveKey(errMsg))

	// verify result files
	for i := 'A'; i < 'A'+int32(count); i++ {
		cStr := fmt.Sprintf("%c", i)
		g.Expect(results).To(HaveKeyWithValue(cStr, cStr))
	}
}

func setupRunner(t *testing.T) *Runner {
	r := DefaultRunner.DeepCopy()
	if filename := os.Getenv(TestKubeconfigVar); len(filename) == 0 {
		t.Skipf("TEST_KUBECONFIG must be set to test against a cluster.")
	} else if restConfig, err := clientcmd.BuildConfigFromFlags("", filename); err != nil {
		t.Fatalf("failed reading '%s' which has been set as the TEST_KUBECONFIG: %v", filename, err)
	} else if r.Kube, err = kubernetes.NewForConfig(restConfig); err != nil {
		t.Fatalf("couldn't setup kube client: %v", err)
	} else if r.Image, err = image.NewForConfig(restConfig); err != nil {
		t.Fatalf("couldn't setup image client: %v", err)
	}
	return r
}

func randomStr(length int) (str string) {
	chars := "0123456789abcdefghijklmnopqrstuvwxyz"
	for i := 0; i < length; i++ {
		c := string(chars[rand.Intn(len(chars))])
		str += c
	}
	return
}
