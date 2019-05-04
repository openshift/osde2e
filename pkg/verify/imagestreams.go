package verify

import (
	"fmt"

	"github.com/onsi/ginkgo"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/osde2e/pkg/config"
)

var _ = ginkgo.Describe("ImageStreams", func() {
	defer ginkgo.GinkgoRecover()

	var err error
	var cluster *Cluster
	ginkgo.BeforeEach(func() {
		cluster, err = NewCluster(config.Cfg.Kubeconfig)
		if err != nil {
			ginkgo.Fail("couldn't configure cluster client: " + err.Error())
		}
		cluster.Setup()
	})

	ginkgo.AfterEach(func() {
		cluster.Cleanup()
	})

	ginkgo.It("ImageStreams should exist in the cluster", func() {
		list, err := cluster.Image().ImageV1().ImageStreams(cluster.proj).List(metav1.ListOptions{})
		if err != nil {
			ginkgo.Fail("Couldn't list clusters: " + err.Error())
		}

		minImages := 20
		if len(list.Items) < minImages {
			msg := fmt.Sprintf("wanted at least '%d' images but have only '%d'", minImages, len(list.Items))
			ginkgo.Fail(msg)
		}
	})
})
