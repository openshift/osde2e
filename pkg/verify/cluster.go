package verify

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/onsi/ginkgo"
	projectv1 "github.com/openshift/api/project/v1"
	image "github.com/openshift/client-go/image/clientset/versioned"
	project "github.com/openshift/client-go/project/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/openshift/osde2e/pkg/config"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func NewCluster(kubeconfig []byte) (*Cluster, error) {
	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	cluster := &Cluster{
		restConfig: restConfig,
	}
	return cluster, nil
}

func newCluster(kubeconfig []byte) *Cluster {
	cluster, err := NewCluster(config.Cfg.Kubeconfig)
	if err != nil {
		ginkgo.Fail("couldn't configure cluster client: " + err.Error())
	}
	return cluster
}

type Cluster struct {
	restConfig *rest.Config
	proj       string
}

func (c *Cluster) Setup() {
	// setup project to run tests
	prefix := randomStr(5)
	proj, err := c.createProject(prefix)
	if err != nil {
		ginkgo.Fail("could not create project: " + err.Error())
		return
	} else if proj == nil {
		ginkgo.Fail("could not create project")
		return
	}

	c.proj = proj.Name
}

func (c *Cluster) Cleanup() {
	err := c.cleanup(c.proj)
	if err != nil {
		msg := fmt.Sprintf("could not delete project '%s': %v", c.proj, err)
		ginkgo.Fail(msg)
		return
	}

	c.proj = ""
}

func (c *Cluster) Image() image.Interface {
	client, err := image.NewForConfig(c.restConfig)
	if err != nil {
		ginkgo.Fail("failed to create Image clientset: " + err.Error())
	}
	return client
}

func (c *Cluster) Project() project.Interface {
	client, err := project.NewForConfig(c.restConfig)
	if err != nil {
		ginkgo.Fail("failed to create Project clientset: " + err.Error())
	}
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
	return nil
}
