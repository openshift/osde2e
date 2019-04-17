package verify

import (
	image "github.com/openshift/client-go/image/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Cluster) checkImageStreams() (interface{}, error) {
	proj, err := c.createProject("imagestreams")
	if err != nil {
		return nil, err
	}
	defer c.cleanup(proj.Name)

	client, err := image.NewForConfig(c.cfg)
	if err != nil {
		return nil, err
	}

	list, err := client.ImageV1().ImageStreams(proj.Name).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	if len(list.Items) > 20 {
		return true, nil
	} else {
		return false, nil
	}
}
