// Package helper provides utilities to assist with osde2e testing.
package helper

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	projectv1 "github.com/openshift/api/project/v1"
	"github.com/openshift/osde2e-common/pkg/clients/openshift"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
	"github.com/openshift/osde2e/pkg/common/runner"
	"github.com/openshift/osde2e/pkg/common/templates"
	"github.com/openshift/osde2e/pkg/common/util"
)

// Init is a common helper function to import the run state into Helper
func Init() *H {
	h := &H{
		mutex: sync.Mutex{},
	}
	return h
}

// New instantiates a helper function to be used within a Ginkgo Test block
func New() *H {
	h := Init()
	err := h.Setup()
	if err != nil {
		log.Fatalf("Error creating helper: %s", err.Error())
	}

	return h
}

// NewOutsideGinkgo instantiates a helper function while not within a Ginkgo Test Block
func NewOutsideGinkgo() (*H, error) {
	defer ginkgo.GinkgoRecover()

	h := Init()
	h.OutsideGinkgo = true
	err := h.Setup()
	if err != nil {
		return nil, err
	}

	return h, nil
}

// H configures clients and sets up and destroys Projects for test isolation.
type H struct {
	ServiceAccount string
	OutsideGinkgo  bool
	Client         *openshift.Client
	// internal
	restConfig *rest.Config
	proj       *projectv1.Project
	mutex      sync.Mutex
}

// Setup configures a *rest.Config using the embedded kubeconfig then sets up a Project for tests to run in.
func (h *H) Setup() error {
	var err error

	defer ginkgo.GinkgoRecover()

	ctx := context.TODO()
	if err = config.LoadKubeconfig(); err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	h.restConfig, err = clientcmd.RESTConfigFromKubeConfig([]byte(viper.GetString(config.Kubeconfig.Contents)))
	if h.OutsideGinkgo && err != nil {
		return fmt.Errorf("error generating restconfig: %s", err.Error())
	}

	Expect(err).ShouldNot(HaveOccurred(), "failed to configure client")

	h.Client, err = openshift.NewFromRestConfig(h.restConfig, ginkgo.GinkgoLogr)
	if h.OutsideGinkgo && err != nil {
		return fmt.Errorf("failed to create openshift client: %w", err)
	}
	Expect(err).ShouldNot(HaveOccurred(), "failed to configure openshift client")

	project := viper.GetString(config.Project)
	if project == "" {
		suffix := util.RandomStr(5)
		h.proj, err = h.SetupNewProject(ctx, suffix)
		if h.OutsideGinkgo && err != nil {
			return fmt.Errorf("failed to set up project: %s", err.Error())
		}
		Expect(err).ShouldNot(HaveOccurred(), "failed to set up project")
		viper.Set(config.Project, h.proj.Name)
	} else {
		if h.proj == nil {
			_ = wait.PollUntilContextTimeout(ctx, 1*time.Second, 30*time.Second, true, func(ctx context.Context) (bool, error) {
				h.proj, err = h.Project().ProjectV1().Projects().Get(ctx, project, metav1.GetOptions{})
				if err != nil && (apierrors.IsNotFound(err) || apierrors.IsServiceUnavailable(err)) {
					return false, nil
				}
				return true, err
			})

			if h.OutsideGinkgo && err != nil {
				return fmt.Errorf("error retrieving project: %s, %v", project, err.Error())
			}
			Expect(err).ShouldNot(HaveOccurred(), "failed to retrieve project: %s", project)
			Expect(h.proj).ShouldNot(BeNil())
		}
	}

	// Set the default service account for future helper-method-calls
	h.SetServiceAccount(ctx, viper.GetString(config.Tests.ServiceAccount))

	return nil
}

// SetupNewProject creates new project prefixed with osde2e-. Also creates essential serviceaccounts and rolebindings in the project.
func (h *H) SetupNewProject(ctx context.Context, suffix string) (*projectv1.Project, error) {
	var err error
	// setup project and dedicated-admin account to run tests
	// the service account is provisioned but only used when specified
	var v1project *projectv1.Project
	ginkgo.GinkgoLogr.Info("Setup called", "project", "osde2e-"+suffix)
	v1project, err = h.createProject(ctx, suffix)
	if h.OutsideGinkgo && err != nil {
		return nil, fmt.Errorf("failed to create project: %s", err.Error())
	}
	Expect(err).ShouldNot(HaveOccurred(), "failed to create project")
	Expect(v1project).ShouldNot(BeNil())

	h.CreateServiceAccountsInProject(ctx, v1project)

	// Quick fix for Hypershift pipelines failing. Currently the RBAC is not being created in the projects.
	if !viper.GetBool(config.Hypershift) {
		err = wait.PollUntilContextTimeout(ctx, 1*time.Second, 1*time.Minute, false, func(ctx context.Context) (bool, error) {
			_, err = h.Kube().RbacV1().RoleBindings(v1project.Name).Get(ctx, "dedicated-admins-project-dedicated-admins", metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				return false, nil
			}
			return true, err
		})
		if h.OutsideGinkgo && err != nil {
			return nil, fmt.Errorf("failed to create role bindings: %s", err.Error())
		}
	}
	return v1project, nil
}

// Adds essential secrets to test namespace to load onto ad hoc test pod
func (h *H) SetPassthruSecretInProject(ctx context.Context, project *projectv1.Project) error {
	err := h.GetClient().Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ci-secrets",
			Namespace: project.Name,
		},
		StringData: viper.GetStringMapString(config.NonOSDe2eSecrets),
	})
	return err
}

// Cleanup deletes a Project after tests have been ran.
func (h *H) Cleanup(ctx context.Context) {
	var err error

	h.restConfig, err = clientcmd.RESTConfigFromKubeConfig([]byte(viper.GetString(config.Kubeconfig.Contents)))
	if err != nil {
		ginkgo.GinkgoLogr.Error(err, "error setting Cleanup() restConfig")
		return
	}

	// Set the SA back to the default. This is required for cleanup in case other helper calls switched SAs
	h.SetServiceAccount(ctx, viper.GetString(config.Tests.ServiceAccount))
	projects, err := h.Project().ProjectV1().Projects().List(ctx, metav1.ListOptions{})
	if err != nil {
		ginkgo.GinkgoLogr.Error(err, "error listing existing projects")
	}
	// todo use k8s client to get namespace object and delete instead of for looping
	for _, project := range projects.Items {
		if h.proj.Name == project.Name {
			ginkgo.GinkgoLogr.Info(fmt.Sprintf("Deleting project `%s`", project.Name))
			err = h.Project().ProjectV1().Projects().Delete(ctx, project.Name, metav1.DeleteOptions{})
			if err != nil {
				ginkgo.GinkgoLogr.Error(err, fmt.Sprintf("error deleting project %q", project.Name))
			}
		}
	}

	h.restConfig = nil
	h.proj = nil
}

// CreateServiceAccountsInProject creates a set of serviceaccounts for test usage in given namespace
func (h *H) CreateServiceAccountsInProject(ctx context.Context, project *projectv1.Project) *H {
	// Create project-specific dedicated-admin account
	err := h.Client.Create(ctx, &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dedicated-admin-project",
			Namespace: project.Name,
		},
	})
	Expect(err).NotTo(HaveOccurred(), "failed to create SA: dedicated-admin-project")
	h.CreateClusterRoleBindingInProject(ctx, "dedicated-admin-project", "dedicated-admins-project", project)
	ginkgo.GinkgoLogr.Info(fmt.Sprintf("Created SA: %v", "dedicated-admin-project"))

	// Create cluster dedicated-admin account
	err = h.Client.Create(ctx, &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dedicated-admin-cluster",
			Namespace: project.Name,
		},
	})
	Expect(err).NotTo(HaveOccurred(), "failed to create SA: dedicated-admin-cluster")
	h.CreateClusterRoleBindingInProject(ctx, "dedicated-admin-cluster", "dedicated-admins-cluster", project)
	ginkgo.GinkgoLogr.Info(fmt.Sprintf("Created SA: %v", "dedicated-admin-cluster"))

	// Create cluster-admin account
	err = h.Client.Create(ctx, &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-admin",
			Namespace: project.Name,
		},
	})
	Expect(err).NotTo(HaveOccurred(), "failed to create SA: cluster-admin")
	h.CreateClusterRoleBindingInProject(ctx, "cluster-admin", "cluster-admin", project)
	ginkgo.GinkgoLogr.Info(fmt.Sprintf("Created SA: %v", "cluster-admin"))

	return h
}

// CreateClusterRoleBindingInProject takes an sa (presumably created by us) and applies a clusterRole to it
// The cr is bound to the given project and, thus, cleaned up when the project gets removed.
func (h *H) CreateClusterRoleBindingInProject(ctx context.Context, saName string, clusterRole string, project *projectv1.Project) {
	gvk := schema.FromAPIVersionAndKind("project.openshift.io/v1", "Project")
	projRef := *metav1.NewControllerRef(project.DeepCopy(), gvk)

	// create binding with OwnerReference
	err := h.Client.Create(ctx, &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "osde2e-test-access-",
			OwnerReferences: []metav1.OwnerReference{
				projRef,
			},
			Namespace: project.Name,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      saName,
				Namespace: project.Name,
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
func (h *H) SetServiceAccount(ctx context.Context, sa string) *H {
	if h.restConfig == nil {
		ginkgo.GinkgoLogr.Info("No restconfig found in SetServiceAccount")
		return nil
	}

	if strings.Contains(sa, "%s") {
		sa = fmt.Sprintf(sa, h.CurrentProject())
	}

	h.ServiceAccount = sa

	if h.ServiceAccount != "" {
		parts := strings.Split(h.ServiceAccount, ":")
		Expect(len(parts)).Should(Equal(4), "not a valid service account name: %v", h.ServiceAccount)
		sa := &corev1.ServiceAccount{}
		err := h.Client.Get(ctx, parts[3], parts[2], sa)
		Expect(err).ShouldNot(HaveOccurred(), "could not get sa '%s'", h.ServiceAccount)
	}

	h.restConfig.Impersonate = rest.ImpersonationConfig{
		UserName: h.ServiceAccount,
	}

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
func (h *H) CreateProject(ctx context.Context, name string) {
	var err error
	defer ginkgo.GinkgoRecover()
	h.proj, err = h.createProject(ctx, name)
	Expect(err).To(BeNil(), fmt.Sprintf("error creating project: %s", err))
}

// DeleteProject deletes the project provided
func (h *H) DeleteProject(ctx context.Context, name string) error {
	err := h.Project().ProjectV1().Projects().Delete(ctx, name, metav1.DeleteOptions{})
	if err == nil {
		return wait.PollUntilContextTimeout(ctx, 5*time.Second, 60*time.Second, false, func(ctx context.Context) (bool, error) {
			project, err := h.Project().ProjectV1().Projects().Get(ctx, name, metav1.GetOptions{})
			if err != nil && project.Name == "" {
				return true, nil
			}

			return false, err
		})
	}
	return err
}

// CurrentProject returns the project being used for testing.
func (h *H) CurrentProject() string {
	Expect(h.proj).NotTo(BeNil(), "no project is currently set")
	return h.proj.Name
}

// SetProjectByName gets a project by name and sets it for the h.proj attribute
func (h *H) SetProjectByName(ctx context.Context, projectName string) (*H, error) {
	var err error
	h.proj, err = h.Project().ProjectV1().Projects().Get(ctx, projectName, metav1.GetOptions{})
	Expect(err).To(BeNil(), "error retrieving project")
	return h, nil
}

// GetWorkloads returns a list of workloads this osde2e run has installed
func (h *H) GetWorkloads() map[string]string {
	return viper.GetStringMapString(config.InstalledWorkloads)
}

// GetWorkload takes a workload name and returns true or false depending on if it's installed
func (h *H) GetWorkload(name string) (string, bool) {
	if val, ok := h.GetWorkloads()[name]; ok {
		return val, true
	}

	return "", false
}

// AddWorkload uniquely appends a workload to the workloads list
func (h *H) AddWorkload(name, project string) {
	h.mutex.Lock()
	installedWorkloads := h.GetWorkloads()
	installedWorkloads[name] = project
	viper.Set(config.InstalledWorkloads, installedWorkloads)
	h.mutex.Unlock()
}

// ConvertTemplateToString takes a template and uses the provided data interface to construct a command string
func (h *H) ConvertTemplateToString(template *template.Template, data interface{}) (string, error) {
	var cmd bytes.Buffer
	if err := template.Execute(&cmd, data); err != nil {
		return "", fmt.Errorf("failed templating command: %v", err)
	}
	return cmd.String(), nil
}

// InspectState inspects the project used for testing, and saves the state to disk for later debugging
func (h *H) InspectState(ctx context.Context) {
	var err error

	h.restConfig, err = clientcmd.RESTConfigFromKubeConfig([]byte(viper.GetString(config.Kubeconfig.Contents)))
	Expect(err).ShouldNot(HaveOccurred(), "failed to configure client")

	// Set the SA back to the default. This is required for inspection in case other helper calls switched SAs
	h.SetServiceAccount(ctx, viper.GetString(config.Tests.ServiceAccount))
	project := viper.GetString(config.Project)

	if h.proj == nil && project != "" {
		h.proj, err = h.Project().ProjectV1().Projects().Get(ctx, project, metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred(), "failed to retrieve project")
		Expect(h.proj).ShouldNot(BeNil())
	}

	// Always inspect the E2E project
	inspectProjects := []string{h.CurrentProject()}

	// Add any additional configured projects to inspect
	projectsToInspectStr := viper.GetString(config.Cluster.InspectNamespaces)
	if projectsToInspectStr != "" {
		inspectProjects = append(inspectProjects, strings.Split(projectsToInspectStr, ",")...)
	}

	err = h.inspect(ctx, inspectProjects)
	Expect(err).ShouldNot(HaveOccurred(), "could not inspect project '%s'", h.proj)
}

// GetClusterVersion returns the Cluster Version object
func (h *H) GetClusterVersion(ctx context.Context) (*configv1.ClusterVersion, error) {
	cfgClient := h.Cfg()
	getOpts := metav1.GetOptions{}
	clusterVersionObj, err := cfgClient.ConfigV1().ClusterVersions().Get(ctx, "version", getOpts)
	if err != nil {
		return nil, fmt.Errorf("couldn't get current ClusterVersion '%s': %v", "version", err)
	}
	return clusterVersionObj, nil
}

// WithToken returns helper with a given bearer token
func (h *H) WithToken(token string) *H {
	config := rest.AnonymousClientConfig(h.restConfig)
	config.BearerToken = token
	return &H{
		restConfig: config,
	}
}

func (h *H) GetConfig() *rest.Config {
	return h.restConfig
}

func (h *H) GetClient() *openshift.Client {
	return h.Client
}

// GetRunnerCommandString Generates templated command string to provide to test adHocTest container
func (h *H) GetRunnerCommandString(templatePath string, timeout int, latestImageStream string, adHocTestImage string, suffix string, jobName string, serviceAccountDir string, command string, serviceAccountNamespacedName string) string {
	ginkgo.GinkgoHelper()
	values := struct {
		Name                 string
		JobName              string
		Arguments            string
		Timeout              int
		Image                string
		Command              string
		OutputDir            string
		ServiceAccount       string
		PushResultsContainer string
		Suffix               string
		Server               string
		CA                   string
		TokenFile            string
		EnvironmentVariables []struct {
			Name  string
			Value string
		}
		EnvironmentVariablesFromSecret []struct {
			SecretName string
			SecretKey  string
		}
	}{
		Name:                 jobName,
		JobName:              jobName,
		Timeout:              timeout,
		Image:                adHocTestImage,
		OutputDir:            runner.DefaultRunner.OutputDir,
		ServiceAccount:       serviceAccountNamespacedName,
		PushResultsContainer: latestImageStream,
		Suffix:               suffix,
		Server:               "https://kubernetes.default",
		CA:                   serviceAccountDir + "/ca.crt",
		TokenFile:            serviceAccountDir + "/token",
		EnvironmentVariables: []struct {
			Name  string
			Value string
		}{
			{
				Name:  "OCM_CLUSTER_ID",
				Value: viper.GetString(config.Cluster.ID),
			},
			{
				Name:  "CLOUD_PROVIDER_ID",
				Value: viper.GetString(config.CloudProvider.CloudProviderID),
			},
			{
				Name:  "CLOUD_PROVIDER_REGION",
				Value: viper.GetString(config.CloudProvider.Region),
			},
			{
				Name:  "OCM_ENV",
				Value: viper.GetString(ocmprovider.Env),
			},
		},
		// loaded as secrets to pod from env vars
		EnvironmentVariablesFromSecret: []struct {
			SecretName string
			SecretKey  string
		}{
			{
				SecretName: "ci-secrets",
				SecretKey:  "OCM_TOKEN",
			},
			{
				SecretName: "ci-secrets",
				SecretKey:  "OCM_CLIENT_SECRET",
			},
			{
				SecretName: "ci-secrets",
				SecretKey:  "OCM_CLIENT_ID",
			},
			{
				SecretName: "ci-secrets",
				SecretKey:  "AWS_ACCESS_KEY_ID",
			},
			{
				SecretName: "ci-secrets",
				SecretKey:  "AWS_PROFILE",
			},
			{
				SecretName: "ci-secrets",
				SecretKey:  "AWS_REGION",
			},
			{
				SecretName: "ci-secrets",
				SecretKey:  "AWS_SECRET_ACCESS_KEY",
			},
			{
				SecretName: "ci-secrets",
				SecretKey:  "GCP_CREDS_JSON",
			},
			{
				SecretName: "ci-secrets",
				SecretKey:  "CAD_PAGERDUTY_ROUTING_KEY",
			},
		},
		Command: command,
	}
	testTemplate, err := templates.LoadTemplate(templatePath)
	Expect(err).NotTo(HaveOccurred(), "Could not load pod template")
	cmd, err := h.ConvertTemplateToString(testTemplate, values)
	Expect(err).NotTo(HaveOccurred(), "Could not convert pod template")
	return cmd
}
