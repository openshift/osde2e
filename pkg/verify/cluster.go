package verify

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	projectv1 "github.com/openshift/api/project/v1"
	image "github.com/openshift/client-go/image/clientset/versioned"
	project "github.com/openshift/client-go/project/clientset/versioned"
	route "github.com/openshift/client-go/route/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/openshift/osde2e/pkg/config"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func NewCluster() (*config.Config, *Cluster) {
	cfg, cluster := config.Cfg, new(Cluster)
	ginkgo.BeforeEach(func() {
		cluster.Setup(cfg.Kubeconfig)
	})
	ginkgo.AfterEach(cluster.Cleanup)
	return cfg, cluster
}

type Cluster struct {
	restConfig *rest.Config
	proj       string
}

func (c *Cluster) Setup(kubeconfig []byte) {
	var err error
	c.restConfig, err = clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	Expect(err).ShouldNot(HaveOccurred(), "failed to configure client")

	// setup project to run tests
	prefix := randomStr(5)
	proj, err := c.createProject(prefix)
	Expect(err).ShouldNot(HaveOccurred(), "failed to create project")
	Expect(proj).ShouldNot(BeNil())

	c.proj = proj.Name
}

func (c *Cluster) Cleanup() {
	err := c.cleanup(c.proj)
	Expect(err).ShouldNot(HaveOccurred(), "could not delete project '%s'", c.proj)

	c.proj = ""
}

func (c *Cluster) Kube() kubernetes.Interface {
	client, err := kubernetes.NewForConfig(c.restConfig)
	Expect(err).ShouldNot(HaveOccurred(), "failed to configure Kubernetes clientset")
	return client
}

func (c *Cluster) Image() image.Interface {
	client, err := image.NewForConfig(c.restConfig)
	Expect(err).ShouldNot(HaveOccurred(), "failed to configure Image clientset")
	return client
}

func (c *Cluster) Route() route.Interface {
	client, err := route.NewForConfig(c.restConfig)
	Expect(err).ShouldNot(HaveOccurred(), "failed to configure Route clientset")
	return client
}

func (c *Cluster) Project() project.Interface {
	client, err := project.NewForConfig(c.restConfig)
	Expect(err).ShouldNot(HaveOccurred(), "failed to configure Project clientset")
	return client
}

func randomStr(length int) (str string) {
	chars := "0123456789abcdefghijklmnopqrstuvwxyz"
	for i := 0; i < length; i++ {
		c := string(chars[rand.Intn(len(chars))])
		str += c
	}
	return
}

func (c *Cluster) createProject(suffix string) (*projectv1.Project, error) {
	proj := &projectv1.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: "osde2e-" + suffix,
		},
	}
	return c.Project().ProjectV1().Projects().Create(proj)
}

func (c *Cluster) cleanup(projectName string) error {
	err := c.Project().ProjectV1().Projects().Delete(projectName, &metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to cleanup project '%s': %v", projectName, err)
	}
	return nil
}
