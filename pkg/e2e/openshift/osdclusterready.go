// Package openshift contains the OpenShift extended test suite.
package openshift

import (
	"context"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e-common/pkg/clients/openshift"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/label"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	suiteName = "OSD Cluster Ready"
	timeout   = 60 * time.Minute
)

// Assert osd-cluster-ready status
var _ = ginkgo.Describe(suiteName, ginkgo.Ordered, label.OCPNightlyBlocking, func() {
	var k8s *openshift.Client

	ginkgo.BeforeAll(func(ctx context.Context) {
		log.SetLogger(ginkgo.GinkgoLogr)
		var err error
		k8s, err = openshift.NewFromKubeconfig(viper.GetString(config.Kubeconfig.Path), ginkgo.GinkgoLogr)
		Expect(err).ShouldNot(HaveOccurred(), "Unable to setup k8s client")
	})

	ginkgo.It("should verify cluster is ready", func(ctx context.Context) {
		joberr := k8s.OSDClusterHealthy(ctx, viper.GetString(config.ReportDir), timeout)
		Expect(joberr).ShouldNot(HaveOccurred(), "osd-cluster-ready job did not succeed")
	})
})
