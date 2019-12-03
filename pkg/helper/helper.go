// Package helper provides utilities to assist with osde2e testing.
package helper

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	projectv1 "github.com/openshift/api/project/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/openshift/osde2e/pkg/config"
)

func init() {
	rand.Seed(time.Now().Unix())
}

// New creates H, a helper used to expose common testing functions.
func New() *H {
	helper := &H{
		Config: config.Cfg,
	}
	
	ginkgo.BeforeEach(helper.Setup)
	ginkgo.AfterEach(helper.Cleanup)
	return helper
}

// H configures clients and sets up and destroys Projects for test isolation.
type H struct {
	// embed test configuration
	*config.Config

	// internal
	restConfig *rest.Config
	proj       *projectv1.Project
}

// SetupNoProj configures a *rest.Config using the embedded kubeconfig
func (h *H) SetupNoProj() {
	var err error
	
	if len(h.InstalledWorkloads) < 1 {
		h.InstalledWorkloads = make(map[string]string)
	}

	h.restConfig, err = clientcmd.RESTConfigFromKubeConfig(h.Kubeconfig)
	Expect(err).ShouldNot(HaveOccurred(), "failed to configure client")
}

// Setup configures a *rest.Config using the embedded kubeconfig then sets up a Project for tests to run in.
func (h *H) Setup() {
	var err error
	h.restConfig, err = clientcmd.RESTConfigFromKubeConfig(h.Kubeconfig)
	Expect(err).ShouldNot(HaveOccurred(), "failed to configure client")

	if len(h.InstalledWorkloads) < 1 {
		h.InstalledWorkloads = make(map[string]string)
	}


	// setup project to run tests
	suffix := randomStr(5)
	proj, err := h.createProject(suffix)
	Expect(err).ShouldNot(HaveOccurred(), "failed to create project")
	Expect(proj).ShouldNot(BeNil())

	h.proj = proj
}

// Cleanup deletes a Project after tests have been ran.
func (h *H) Cleanup() {
	if h.proj != nil {
		err := h.cleanup(h.proj.Name)
		Expect(err).ShouldNot(HaveOccurred(), "could not delete project '%s'", h.proj)
	}

	h.restConfig = nil
	h.proj = nil
}

// SetRestConfig

// CreateProject returns the project being used for testing.
func (h *H) CreateProject(name string) {
	proj, err := h.createProject(name)
	Expect(err).To(BeNil(), "error creating project")
	h.proj = proj
}

// CurrentProject returns the project being used for testing.
func (h *H) CurrentProject() string {
	Expect(h.proj).NotTo(BeNil(), "no project is currently set")
	return h.proj.Name
}

// SetProjectByName gets a project by name and sets it for the h.proj attribute
func (h *H) SetProjectByName(projectName string) error {
	proj, err := h.Project().ProjectV1().Projects().Get(projectName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get project '%s': %v", projectName, err)
	}
	h.proj = proj
	return nil
}

// GetWorkloads returns a list of workloads this osde2e run has installed
func (h *H) GetWorkloads() map[string]string {
	return h.InstalledWorkloads
}

// GetWorkload takes a workload name and returns true or false depending on if it's installed
func (h *H) GetWorkload(name string) (string, bool) {
	if val, ok := h.InstalledWorkloads[name]; ok {
		return val, true
	}

	return "", false
}

// AddWorkload uniquely appends a workload to the workloads list
func (h *H) AddWorkload(name, project string) {
	h.InstalledWorkloads[name] = project
}
