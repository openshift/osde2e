package cloudingress

import (
	"github.com/onsi/ginkgo/v2"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/constants"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/providers/rosaprovider"
	"github.com/openshift/osde2e/pkg/e2e/operators"
)

var _ = ginkgo.Describe(constants.SuiteInforming+TestPrefix, func() {
	ginkgo.BeforeEach(func() {
		if viper.GetBool(rosaprovider.STS) {
			ginkgo.Skip("STS does not support CIO")
		}
	})

	h := helper.New()
	operators.CheckUpgrade(h, OperatorNamespace, OperatorName, OperatorName, "cloud-ingress-operator-registry")
	operators.CheckPod(h, OperatorNamespace, OperatorName, 200, 0)
})
