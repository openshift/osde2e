package cloudingress

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	"github.com/openshift/osde2e/pkg/common/providers/rosaprovider"
	"github.com/openshift/osde2e/pkg/e2e/operators"
)

// tests

var _ = ginkgo.Describe("[Suite: operators] "+TestPrefix, label.Operators, func() {
	ginkgo.BeforeEach(func() {
		if viper.GetBool(rosaprovider.STS) {
			ginkgo.Skip("STS does not support CIO")
		}
		if viper.GetBool(config.Hypershift) {
			ginkgo.Skip("Cloud Ingress Operator is not supported on HyperShift")
		}
	})

	var defaultDesiredReplicas int32 = 1

	h := helper.New()

	// Check that the operator deployment exists in the operator namespace
	ginkgo.Context("deployment", func() {
		ginkgo.It("should exist", func(ctx context.Context) {
			deployment, err := operators.PollDeployment(ctx, h, OperatorNamespace, OperatorName)
			Expect(err).ToNot(HaveOccurred(), "failed fetching deployment")
			Expect(deployment).NotTo(BeNil(), "deployment is nil")
		})

		ginkgo.It("should have all desired replicas ready", func(ctx context.Context) {
			deployment, err := operators.PollDeployment(ctx, h, OperatorNamespace, OperatorName)
			Expect(err).ToNot(HaveOccurred(), "failed fetching deployment")

			readyReplicas := deployment.Status.ReadyReplicas
			desiredReplicas := deployment.Status.Replicas

			// The desired replicas should match the default installed replica count
			Expect(desiredReplicas).To(BeNumerically("==", defaultDesiredReplicas), "The deployment desired replicas should not drift from the default 1.")

			// Desired replica count should match ready replica count
			Expect(readyReplicas).To(BeNumerically("==", desiredReplicas), "All desired replicas should be ready.")
		})
	})
})
