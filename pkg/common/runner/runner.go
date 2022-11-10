// Package runner provides a wrapper for the OpenShift extended test suite image.
package runner

import (
	"fmt"
	"log"
	"os"

	image "github.com/openshift/client-go/image/clientset/versioned"
	"github.com/openshift/osde2e/pkg/common/util"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube "k8s.io/client-go/kubernetes"
)

const (
	// name used in Runner resources
	defaultName = "osde2e-runner"
	// directory containing service account credentials
	serviceAccountDir = "/var/run/secrets/kubernetes.io/serviceaccount"
)

// DefaultRunner is a runner with the most commonly desired settings.
var DefaultRunner = &Runner{
	Name:                 defaultName,
	ImageStreamName:      testImageStreamName,
	ImageStreamNamespace: testImageStreamNamespace,
	PodSpec: kubev1.PodSpec{
		Containers: []kubev1.Container{
			DefaultContainer,
		},
		RestartPolicy: kubev1.RestartPolicyNever,
	},
	OutputDir: "/test-run-results",
	Server:    "https://kubernetes.default",
	CA:        serviceAccountDir + "/ca.crt",
	TokenFile: serviceAccountDir + "/token",
	Logger:    log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile),
}

// Runner runs the OpenShift extended test suite within a cluster.
type Runner struct {
	// Kube client used to run test in-cluster.
	Kube kube.Interface

	// Image client used to gather ImageStream information.
	Image image.Interface

	// Name of the operation being performed.
	Name string

	// Server is the endpoint within the pod the Kubernetes API should be
	Server string

	// CA is the CA bundle used to auth against the Kubernetes API
	CA string

	// TokenFile is the credentials used to auth against the Kubernetes API
	TokenFile string

	// Namespace runner resources should be created in.
	Namespace string

	// ImageStreamName is the name of the ImageStream containing the suite.
	ImageStreamName string

	// ImageStreamNamespace is the namespace of the ImageStream containing the suite.
	ImageStreamNamespace string

	// ImageName is a container image used for the runner.
	ImageName string

	// Cmd is run within the test pod. If PodSpec is also set it overrides the container of the same name.
	Cmd string

	// PodSpec defines the Pod used by the runner.
	PodSpec kubev1.PodSpec

	// OutputDir is the directory that is copied from the Pod to the local host.
	OutputDir string

	// Tarball will create a single .tgz file for the entire OutputDir.
	Tarball bool

	// SkipLogsFromPod should be set to true if logs should not be collected.
	SkipLogsFromPod bool

	// Repos are cloned and mounted into the test Pod.
	Repos

	// Logger receives all messages.
	*log.Logger

	// internal
	stopCh <-chan struct{}
	svc    *kubev1.Service
	status Status
}

// Run deploys the suite into a cluster, waits for it to finish, and gathers the results.
func (r *Runner) Run(timeoutInSeconds int, stopCh <-chan struct{}) (err error) {
	r.stopCh = stopCh
	r.status = StatusSetup

	// set image if imagestream is set
	if r.ImageName == "" {
		if r.ImageName, err = r.GetLatestImageStreamTag(); err != nil {
			return
		}
	}
	log.Printf("Using '%s' as image for runner", r.ImageName)

	log.Printf("Creating %s runner Pod...", r.Name)
	var pod *kubev1.Pod
	if pod, err = r.createPod(); err != nil {
		return
	}

	log.Printf("Waiting for %s runner Pod to start...", r.Name)
	if err = r.waitForPodRunning(pod); err != nil {
		return
	}
	r.status = StatusRunning

	log.Printf("Creating service for %s runner Pod...", r.Name)
	if r.svc, err = r.createService(pod); err != nil {
		return
	}

	log.Printf("Waiting for endpoints of %s runner Pod with a timeout of %d seconds...", r.Name, timeoutInSeconds)
	var completionErr error
	completionErr = r.waitForCompletion(pod.Name, timeoutInSeconds)

	if !r.SkipLogsFromPod {
		log.Printf("Collecting logs from containers on %s runner Pod...", r.Name)
		if err = r.getAllLogsFromPod(pod.Name); err != nil {
			return
		}
	} else {
		log.Printf("Skipping logs from pod %s", r.Name)
	}

	if completionErr != nil {
		return completionErr
	}

	log.Printf("%s runner is done", r.Name)
	r.status = StatusDone
	return nil
}

// Status returns the current state of the runner.
func (r *Runner) Status() Status {
	return r.status
}

// DeepCopy returns a deep copy of a runner.
func (r *Runner) DeepCopy() *Runner {
	newRunner := *DefaultRunner

	// copy repos & PodSpec
	newRunner.Repos = make(Repos, len(r.Repos))
	copy(newRunner.Repos, r.Repos)
	newRunner.PodSpec = *r.PodSpec.DeepCopy()

	return &newRunner
}

// meta returns the ObjectMeta used for Runner resources.
func (r *Runner) meta() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name: fmt.Sprintf("%s-%s", r.Name, util.RandomStr(5)),
		Labels: map[string]string{
			"app": r.Name,
		},
	}
}

// Status is the current state of the runner.
type Status string

var (
	StatusSetup   Status = "setup"
	StatusRunning Status = "running"
	StatusDone    Status = "done"
)
