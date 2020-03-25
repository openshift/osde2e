// Package helper provides utilities to assist with osde2e testing.
package helper

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"text/template"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	projectv1 "github.com/openshift/api/project/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	ServiceAccount string

	// internal
	restConfig *rest.Config
	proj       *projectv1.Project
}

// Setup configures a *rest.Config using the embedded kubeconfig then sets up a Project for tests to run in.
func (h *H) Setup() {
	var err error

	h.restConfig, err = clientcmd.RESTConfigFromKubeConfig(h.Kubeconfig.Contents)
	Expect(err).ShouldNot(HaveOccurred(), "failed to configure client")

	if h.State.Project == "" {
		// setup project and dedicated-admin account to run tests
		// the service account is provisioned but only used when specified
		suffix := util.RandomStr(5)
		h.State.Project = "osde2e-" + suffix

		log.Printf("Setup called for %s", h.State.Project)

		h.proj, err = h.createProject(suffix)
		Expect(err).ShouldNot(HaveOccurred(), "failed to create project")
		Expect(h.proj).ShouldNot(BeNil())

		h.CreateServiceAccounts()

		time.Sleep(60 * time.Second)

	} else {
		log.Printf("Setting project name to %s", h.State.Project)
		h.proj, err = h.Project().ProjectV1().Projects().Get(h.State.Project, metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred(), "failed to retrieve project")
		Expect(h.proj).ShouldNot(BeNil())
	}

	h.SetServiceAccount(config.Instance.Tests.ServiceAccount)

	if len(h.InstalledWorkloads) < 1 {
		h.InstalledWorkloads = make(map[string]string)
	}

}

// Cleanup deletes a Project after tests have been ran.
func (h *H) Cleanup() {
	var err error

	h.restConfig, err = clientcmd.RESTConfigFromKubeConfig(h.Kubeconfig.Contents)
	Expect(err).ShouldNot(HaveOccurred(), "failed to configure client")

	h.SetServiceAccount(config.Instance.Tests.ServiceAccount)

	if h.proj == nil && h.State.Project != "" {
		log.Printf("Setting project name to %s", h.State.Project)
		h.proj, err = h.Project().ProjectV1().Projects().Get(h.State.Project, metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred(), "failed to retrieve project")
		Expect(h.proj).ShouldNot(BeNil())

		err = h.cleanup(h.CurrentProject())
		Expect(err).ShouldNot(HaveOccurred(), "could not delete project '%s'", h.proj)
	}

	h.restConfig = nil
	h.proj = nil
}

// CreateServiceAccounts creates a set of serviceaccounts for test usage
func (h *H) CreateServiceAccounts() *H {
	Expect(h.proj).NotTo(BeNil(), "no project is currently set")

	// Create project-specific dedicated-admin account
	sa, err := h.Kube().CoreV1().ServiceAccounts(h.CurrentProject()).Create(&v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dedicated-admin-project",
		},
	})
	Expect(err).NotTo(HaveOccurred())
	h.CreateClusterRoleBinding(sa, "dedicated-admins-project")
	log.Printf("Created SA: %v", sa.GetName())

	// Create cluster dedicated-admin account
	sa, err = h.Kube().CoreV1().ServiceAccounts(h.CurrentProject()).Create(&v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dedicated-admin-cluster",
		},
	})
	Expect(err).NotTo(HaveOccurred())
	h.CreateClusterRoleBinding(sa, "dedicated-admins-cluster")
	log.Printf("Created SA: %v", sa.GetName())

	// Create cluster-admin account
	sa, err = h.Kube().CoreV1().ServiceAccounts(h.CurrentProject()).Create(&v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster-admin",
		},
	})
	Expect(err).NotTo(HaveOccurred())
	h.CreateClusterRoleBinding(sa, "cluster-admin")
	log.Printf("Created SA: %v", sa.GetName())

	return h
}

// CreateClusterRoleBinding takes an sa (presumably created by us) and applies a clusterRole to it
// The cr is bound to the project and, thus, cleaned up when the project gets removed.
func (h *H) CreateClusterRoleBinding(sa *v1.ServiceAccount, clusterRole string) {
	gvk := schema.FromAPIVersionAndKind("project.openshift.io/v1", "Project")
	projRef := *metav1.NewControllerRef(h.proj, gvk)

	// create binding with OwnerReference
	_, err := h.Kube().RbacV1().ClusterRoleBindings().Create(&rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "osde2e-test-access-",
			OwnerReferences: []metav1.OwnerReference{
				projRef,
			},
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      sa.GetName(),
				Namespace: h.CurrentProject(),
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     clusterRole,
		},
	})
	Expect(err).NotTo(HaveOccurred(), "couldn't set correct permissions for OpenShift E2E")
}

// SetServiceAccount sets the serviceAccount you want all helper commands to run as
func (h *H) SetServiceAccount(sa string) *H {
	if h.restConfig == nil {
		log.Print("No restconfig found in SetServiceAccount")
		return nil
	}

	if strings.Contains(sa, "%s") {
		sa = fmt.Sprintf(sa, h.CurrentProject())
	}

	h.ServiceAccount = sa

	if h.ServiceAccount != "" {
		parts := strings.Split(h.ServiceAccount, ":")
		Expect(len(parts)).Should(Equal(4), "not a valid service account name: %v", h.ServiceAccount)
		_, err := h.Kube().CoreV1().ServiceAccounts(parts[2]).Get(parts[3], metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred(), "could not get sa '%s'", h.ServiceAccount)
	}

	h.restConfig.Impersonate = rest.ImpersonationConfig{
		UserName: h.ServiceAccount,
	}
	log.Printf("ServiceAccount is now set to `%v`", h.ServiceAccount)

	return h
}

// GetNamespacedServiceAccount just gets the name, not the "full name"
func (h *H) GetNamespacedServiceAccount() string {
	sa := ""
	if h.ServiceAccount != "" {
		parts := strings.Split(h.ServiceAccount, ":")
		Expect(len(parts)).Should(Equal(4), "not a valid service account name: %v", h.ServiceAccount)
		sa = parts[3]
	}
	return sa
}

// SetProject manually sets the project
func (h *H) SetProject(proj *projectv1.Project) *H {
	h.proj = proj
	return h
}

// CreateProject returns the project being used for testing.
func (h *H) CreateProject(name string) {
	var err error
	h.proj, err = h.createProject(name)
	Expect(err).To(BeNil(), "error creating project")
}

// CurrentProject returns the project being used for testing.
func (h *H) CurrentProject() string {
	Expect(h.proj).NotTo(BeNil(), "no project is currently set")
	return h.proj.Name
}

// SetProjectByName gets a project by name and sets it for the h.proj attribute
func (h *H) SetProjectByName(projectName string) (*H, error) {
	var err error
	h.proj, err = h.Project().ProjectV1().Projects().Get(projectName, metav1.GetOptions{})
	Expect(err).To(BeNil(), "error retrieving project")
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
