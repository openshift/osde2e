// Package helper provides utilities to assist with osde2e testing.
package helper

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"text/template"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	projectv1 "github.com/openshift/api/project/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/state"
	"github.com/openshift/osde2e/pkg/common/util"
)

func init() {
	rand.Seed(time.Now().Unix())
}

// New creates H, a helper used to expose common testing functions.
func New() *H {
	h := &H{
		State: state.Instance,
	}

	ginkgo.BeforeEach(h.Setup)
	return h
}

// H configures clients and sets up and destroys Projects for test isolation.
type H struct {
	// embed state
	*state.State
	Persona string

	// internal
	restConfig *rest.Config
	proj       *projectv1.Project
}

// Setup configures a *rest.Config using the embedded kubeconfig then sets up a Project for tests to run in.
func (h *H) Setup() {
	var err error

	h.restConfig, err = clientcmd.RESTConfigFromKubeConfig(h.Kubeconfig.Contents)
	Expect(err).ShouldNot(HaveOccurred(), "failed to configure client")

	if config.Instance.Tests.Persona != "" {
		h.Persona = config.Instance.Tests.Persona
	}

	if h.State.Project == "" {
		// setup project and dedicated-admin account to run tests
		// the service account is provisioned but only used when specified
		suffix := util.RandomStr(5)
		h.State.Project = "osde2e-" + suffix

		log.Printf("Setup called for %s", h.State.Project)

		sa, err := h.Kube().CoreV1().ServiceAccounts("dedicated-admin").Create(&v1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name: h.State.Project,
			},
		})
		Expect(err).NotTo(HaveOccurred())
		log.Printf("Created SA: %v", sa.GetName())

		proj, err := h.createProject(suffix)
		Expect(err).ShouldNot(HaveOccurred(), "failed to create project")
		Expect(proj).ShouldNot(BeNil())

		h.proj = proj
		time.Sleep(60 * time.Second)

	} else {
		log.Printf("Setting project name to %s", h.State.Project)
		proj, err := h.Project().ProjectV1().Projects().Get(h.State.Project, metav1.GetOptions{})
		if err != nil {
			log.Printf("failed to get project '%s': %v", h.State.Project, err)
		}
		h.proj = proj
	}

	h.SetPersona(h.Persona)

	if len(h.InstalledWorkloads) < 1 {
		h.InstalledWorkloads = make(map[string]string)
	}

}

// Cleanup deletes a Project after tests have been ran.
func (h *H) Cleanup() {
	var err error
	h.restConfig, err = clientcmd.RESTConfigFromKubeConfig(h.Kubeconfig.Contents)
	Expect(err).ShouldNot(HaveOccurred(), "failed to configure client")

	h.restConfig.Impersonate = rest.ImpersonationConfig{
		UserName: "",
	}

	if h.proj == nil && h.State.Project != "" {
		log.Printf("Setting project name to %s", h.State.Project)
		proj, err := h.Project().ProjectV1().Projects().Get(h.State.Project, metav1.GetOptions{})
		if err != nil {
			log.Printf("failed to get project '%s': %v", h.State.Project, err)
		}
		h.proj = proj

		err = h.Kube().CoreV1().ServiceAccounts("dedicated-admin").Delete(h.CurrentProject(), &metav1.DeleteOptions{})
		Expect(err).ShouldNot(HaveOccurred(), "could not delete sa '%s'", h.CurrentProject)

		err = h.cleanup(h.proj.Name)
		Expect(err).ShouldNot(HaveOccurred(), "could not delete project '%s'", h.proj)
	}

	h.Persona = config.Instance.Tests.Persona
	h.restConfig = nil
	h.proj = nil
}

// SetPersona sets a persona for helper-based tests to run as
func (h *H) SetPersona(persona string) *H {
	if h.restConfig == nil {
		log.Print("No restconfig found in SetPersona")
		return nil
	}
	account := fmt.Sprintf("system:serviceaccount:dedicated-admin:%s", persona)
	if persona == "" {
		account = ""
	} else {
		if persona != "dedicated-admin" {
			log.Fatalf("Assuming an invalid persona: %s", persona)
		}
	}

	h.restConfig.Impersonate = rest.ImpersonationConfig{
		UserName: account,
	}

	return h
}

// SetProject manually sets the project
func (h *H) SetProject(proj *projectv1.Project) *H {
	h.proj = proj
	return h
}

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
func (h *H) SetProjectByName(projectName string) (*H, error) {
	proj, err := h.Project().ProjectV1().Projects().Get(projectName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get project '%s': %v", projectName, err)
	}
	h.proj = proj
	return h, nil
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

// ConvertTemplateToString takes a template and uses the provided data interface to construct a command string
func (h *H) ConvertTemplateToString(template *template.Template, data interface{}) (string, error) {
	var cmd bytes.Buffer
	if err := template.Execute(&cmd, data); err != nil {
		return "", fmt.Errorf("failed templating command: %v", err)
	}
	return cmd.String(), nil
}
