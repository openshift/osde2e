package verify

import (
	"fmt"
	"github.com/onsi/ginkgo"
	"math/rand"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	projectv1 "github.com/openshift/api/project/v1"
	project "github.com/openshift/client-go/project/clientset/versioned"

	image "github.com/openshift/client-go/image/clientset/versioned"
)

const TestKubeconfigEnv = "TEST_KUBECONFIG"

func init() {
	rand.Seed(time.Now().Unix())
}

func NewCluster(kubeconfig []byte) (*Cluster, error) {
	if kubeconfig == nil {
		kubeconfigStr, ok := os.LookupEnv(TestKubeconfigEnv)
		if !ok {
			return nil, fmt.Errorf("kubeconfig not provided and couldn't be loaded from '%s'", TestKubeconfigEnv)
		}
		kubeconfig = []byte(kubeconfigStr)
	}

	cfg, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	cluster := &Cluster{
		cfg: cfg,
	}

	ginkgo.BeforeEach(cluster.BeforeEach)
	ginkgo.AfterEach(cluster.AfterEach)
	return cluster, nil
}

type Cluster struct {
	cfg  *rest.Config
	proj string
}

func (c *Cluster) BeforeEach() {
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

func (c *Cluster) AfterEach() {
	err := c.cleanup(c.proj)
	if err != nil {
		msg := fmt.Sprintf("could not delete project '%s': %v", c.proj, err)
		ginkgo.Fail(msg)
		return
	}

	c.proj = ""
}

func (c *Cluster) Image() image.Interface {
	client, err := image.NewForConfig(c.cfg)
	if err != nil {
		ginkgo.Fail("failed to create Image clientset: " + err.Error())
	}
	return client
}

func (c *Cluster) Project() project.Interface {
	client, err := project.NewForConfig(c.cfg)
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

func (c *Cluster) createProject(prefix string) (*projectv1.Project, error) {
	proj := &projectv1.Project{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: prefix,
		},
	}
	return c.Project().ProjectV1().Projects().Create(proj)
}

func (c *Cluster) cleanup(projectName string) error {
	return nil
}
