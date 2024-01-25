// Package openshift contains the OpenShift extended test suite.
package openshift

import (
	"context"
	"fmt"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e-common/pkg/clients/openshift"
	"github.com/openshift/osde2e/pkg/common/label"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	suiteName = "OSD Cluster Ready"
	timeout   = 30 * time.Minute
	namespace = "openshift-monitoring"
	jobname   = "osd-cluster-ready"
)

// Assert osd-cluster-ready status
var _ = ginkgo.Describe(suiteName, ginkgo.Ordered, label.OCPNightlyBlocking, func() {
	var k8s *openshift.Client

	ginkgo.BeforeAll(func(ctx context.Context) {
		log.SetLogger(ginkgo.GinkgoLogr)
		var err error
		k8s, err = openshift.New(ginkgo.GinkgoLogr)
		Expect(err).ShouldNot(HaveOccurred(), "Unable to setup k8s client")
	})

	ginkgo.It("should verify cluster is ready", func(ctx context.Context) {
		joberr := k8s.WatchJob(ctx, namespace, jobname)
		if joberr != nil {
			logs, err := k8s.GetJobLogs(ctx, jobname, namespace)
			if err != nil {
				ginkgo.GinkgoLogr.Error(fmt.Errorf("could not get osd-cluster-ready logs"), err.Error())
			} else {
				ginkgo.GinkgoLogr.Info("job log:", logs)
			}
		}
		Expect(joberr).ShouldNot(HaveOccurred(), "osd-cluster-ready job did not succeed")
	}, ginkgo.SpecTimeout(timeout))
})
