package scale

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/util"
	kubev1 "k8s.io/api/core/v1"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/cluster"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
)

const (
	numNodesToScaleTo = 12
)

var masterVerticalTestName string = "[Suite: scale-mastervertical] Scaling"

func init() {
	alert.RegisterGinkgoAlert(masterVerticalTestName, "SD-CICD", "Diego Santamaria", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(masterVerticalTestName, func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	masterVerticalTimeoutInSeconds := 7200
	util.GinkgoIt("should be tested with MasterVertical", func(ctx context.Context) {
		var err error
		// Before we do anything, scale the cluster.
		err = cluster.ScaleCluster(viper.GetString(config.Cluster.ID), numNodesToScaleTo)
		Expect(err).NotTo(HaveOccurred())

		h.SetServiceAccount(ctx, "system:serviceaccount:%s:cluster-admin")
		// setup runner
		scaleCfg := scaleRunnerConfig{
			Name:         "master-vertical",
			PlaybookPath: "workloads/mastervertical.yml",
		}
		r := scaleCfg.Runner(h)

		// only test on 3 nodes
		r.PodSpec.Containers[0].Env = append(r.PodSpec.Containers[0].Env, kubev1.EnvVar{
			Name:  "MASTERVERTICAL_PROJECTS",
			Value: "100",
		})
		// run tests
		stopCh := make(chan struct{})
		err = r.Run(masterVerticalTimeoutInSeconds, stopCh)
		Expect(err).NotTo(HaveOccurred())
	}, float64(masterVerticalTimeoutInSeconds))
})
