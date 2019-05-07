package verify

import (
	"log"

	"github.com/onsi/ginkgo"
)

var _ = ginkgo.Describe("Pods", func() {
	defer ginkgo.GinkgoRecover()
	_, cluster := NewCluster()

	ginkgo.It("should be mostly running", func() {
		log.Println(cluster.proj)
	})
})
