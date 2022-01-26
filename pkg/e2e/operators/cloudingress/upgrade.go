package cloudingress

import (
	"github.com/onsi/ginkgo/v2"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/constants"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/e2e/operators"
)

var _ = ginkgo.Describe(constants.SuiteInforming+TestPrefix, func() {
	ginkgo.BeforeEach(func() {
		if viper.GetBool("rosa.STS") {
			ginkgo.Skip("STS does not support MVO")
		}
	})

	h := helper.New()
	operators.CheckUpgrade(h, "openshift-cloud-ingress-operator", "cloud-ingress-operator", "cloud-ingress-operator", "cloud-ingress-operator-registry")
	operators.CheckPod(h, "openshift-cloud-ingress-operator", "cloud-ingress-operator", 200, 0)
})
