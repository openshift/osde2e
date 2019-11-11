// Package helper provides utilities to assist with osde2e testing.
package helper

import (
	"math/rand"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	projectv1 "github.com/openshift/api/project/v1"
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
	helper.installedWorkloads = make(map[string]bool)
	ginkgo.BeforeEach(helper.Setup)
	ginkgo.AfterEach(helper.Cleanup)
	return helper
}

// H configures clients and sets up and destroys Projects for test isolation.
type H struct {
	// embed test configuration
	*config.Config

	// internal
	restConfig         *rest.Config
	proj               *projectv1.Project
	installedWorkloads map[string]bool
}

// Setup configures a *rest.Config using the embedded kubeconfig then sets up a Project for tests to run in.
func (h *H) Setup() {
	var err error
	h.restConfig, err = clientcmd.RESTConfigFromKubeConfig(h.Kubeconfig)
	Expect(err).ShouldNot(HaveOccurred(), "failed to configure client")

	// setup project to run tests
	suffix := randomStr(5)
	proj, err := h.createProject(suffix)
	Expect(err).ShouldNot(HaveOccurred(), "failed to create project")
	Expect(proj).ShouldNot(BeNil())

	h.proj = proj
}

// Cleanup deletes a Project after tests have been ran.
func (h *H) Cleanup() {
	err := h.cleanup(h.proj.Name)
	Expect(err).ShouldNot(HaveOccurred(), "could not delete project '%s'", h.proj)

	h.restConfig = nil
	h.proj = nil
}

// CurrentProject returns the project being used for testing.
func (h *H) CurrentProject() string {
	Expect(h.proj).NotTo(BeNil(), "no project is currently set")
	return h.proj.Name
}

// GetWorkloads returns a list of workloads this osde2e run has installed
func (h *H) GetWorkloads() map[string]bool {
	return h.installedWorkloads
}

// GetWorkload takes a workload name and returns true or false depending on if it's installed
func (h *H) GetWorkload(name string) bool {
	_, ok := h.installedWorkloads[name]
	return ok
}

// AddWorkload uniquely appends a workload to the workloads list
func (h *H) AddWorkload(name string) {
	h.installedWorkloads[name] = true
}
