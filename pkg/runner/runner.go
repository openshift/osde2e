// Package runner provides a wrapper for the OpenShift extended test suite image.
package runner

import (
	"log"
	"os"

	image "github.com/openshift/client-go/image/clientset/versioned"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube "k8s.io/client-go/kubernetes"
)

const (
	// name used in Runner resources
	name = "openshift-tests"

	// directory containing service account credentials
	serviceAccountDir = "/var/run/secrets/kubernetes.io/serviceaccount"
)

// DefaultRunner is a runner with the most commonly desired settings.
var DefaultRunner = &Runner{
	ImageStreamName:      testImageStreamName,
	ImageStreamNamespace: testImageStreamNamespace,
	Type:                 RegularTest,
	Suite:                "openshift/conformance",
	Flags: []string{
		"--loglevel=10",
		"--include-success",
		"--junit-dir=./results",
	},
	AuthConfig: AuthConfig{
		Name:      "osde2e",
		Server:    "https://kubernetes.default",
		CA:        serviceAccountDir + "/ca.crt",
		TokenFile: serviceAccountDir + "/token",
	},
	Logger: log.New(os.Stderr, "", log.LstdFlags),
}

// Runner runs the OpenShift extended test suite within a cluster.
type Runner struct {
	// Kube client used to run test in-cluster.
	Kube kube.Interface

	// Image client used to gather ImageStream information.
	Image image.Interface

	// Namespace runner resources should be created in.
	Namespace string

	// ImageStreamName is the name of the ImageStream containing the suite.
	ImageStreamName string

	// ImageStreamNamespace is the namespace of the ImageStream containing the suite.
	ImageStreamNamespace string

	// TestType determines which suite the runner executes.
	Type TestType

	// Suite to be run inside the runner.
	Suite string

	// TestNames explicitly specify which tests to run as part of the suite. No other tests will be run.
	TestNames []string

	// Flags to run the suite with.
	Flags []string

	// Auth defines how to connect to a cluster.
	AuthConfig

	// Logger receives all messages.
	*log.Logger

	// internal
	stopCh    <-chan struct{}
	testImage string
	svc       *kubev1.Service
	status    Status
}

// Run deploys the suite into a cluster, waits for it to finish, and gathers the results.
func (r *Runner) Run(stopCh <-chan struct{}) (err error) {
	r.stopCh = stopCh
	r.status = StatusSetup
	if r.testImage, err = r.getLatestImageStreamTag(); err != nil {
		return
	}

	var pod *kubev1.Pod
	if pod, err = r.createPod(); err != nil {
		return
	}

	if err = r.waitForPodRunning(pod); err != nil {
		return
	}
	r.status = StatusRunning

	if r.svc, err = r.createService(pod); err != nil {
		return
	}

	if err = r.waitForEndpoints(); err != nil {
		return
	}
	r.status = StatusDone
	return nil
}

// Status returns the current state of the runner.
func (r *Runner) Status() Status {
	return r.status
}

// meta returns the ObjectMeta used for Runner resources.
func (r *Runner) meta() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		GenerateName: name + "-",
		Labels: map[string]string{
			"app": name,
		},
	}
}

// TestType determines which suite is run.
type TestType string

var (
	RegularTest TestType = "regular"
	UpgradeTest TestType = "upgrade"
)

// AuthConfig defines how the test Pod connects to the cluster.
type AuthConfig struct {
	Name      string
	Server    string
	CA        string
	TokenFile string
}

// Status is the current state of the runner.
type Status string

var (
	StatusSetup   Status = "setup"
	StatusRunning Status = "running"
	StatusDone    Status = "done"
)
