package verify

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	projectv1 "github.com/openshift/api/project/v1"
	project "github.com/openshift/client-go/project/clientset/versioned"
)

func NewCluster(kubeconfig []byte) (*Cluster, error) {
	cfg, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	return &Cluster{
		cfg: cfg,
	}, nil
}

type Cluster struct {
	cfg *rest.Config
}

func (c *Cluster) createProject(prefix string) (*projectv1.Project, error) {
	client, err := project.NewForConfig(c.cfg)
	if err != nil {
		return nil, err
	}

	proj := &projectv1.Project{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: prefix,
		},
	}
	return client.ProjectV1().Projects().Create(proj)
}

func (c *Cluster) cleanup(projectName string) error {
	return nil
}
