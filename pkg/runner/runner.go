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
	defaultName = "openshift-tests"

	// directory containing service account credentials
	serviceAccountDir = "/var/run/secrets/kubernetes.io/serviceaccount"
)

// DefaultRunner is a runner with the most commonly desired settings.
var DefaultRunner = &Runner{
	Name:                 defaultName,
	ImageStreamName:      testImageStreamName,
	ImageStreamNamespace: testImageStreamNamespace,
	OutputDir:            "./results",
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

	// Name of the operation being performed.
	Name string

	// Namespace runner resources should be created in.
	Namespace string

	// ImageStreamName is the name of the ImageStream containing the suite.
	ImageStreamName string

	// ImageStreamNamespace is the namespace of the ImageStream containing the suite.
	ImageStreamNamespace string

	// Cmd is run within the test pod.
	Cmd string

	// OutputDir is the directory that is copied from the Pod to the local host.
	OutputDir string

	// Tarball will create a single .tgz file for the entire OutputDir.
	Tarball bool

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
		GenerateName: r.Name + "-",
		Labels: map[string]string{
			"app": r.Name,
		},
	}
}

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
