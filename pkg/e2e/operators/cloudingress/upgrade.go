package cloudingress

import (
	"github.com/onsi/ginkgo/v2"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/label"
	"github.com/openshift/osde2e/pkg/common/providers/rosaprovider"
	"github.com/openshift/osde2e/pkg/e2e/operators"
)

var _ = ginkgo.Describe("[Suite: informing] "+TestPrefix, label.Informing, func() {
	ginkgo.BeforeEach(func() {
		if viper.GetBool(rosaprovider.STS) {
			ginkgo.Skip("STS does not support CIO")
		}
		if viper.GetBool(config.Hypershift) {
			ginkgo.Skip("Cloud Ingress Operator is not supported on HyperShift")
		}
	})

	h := helper.New()
	operators.CheckUpgrade(h, OperatorNamespace, OperatorName, OperatorName, "cloud-ingress-operator-registry")
	operators.CheckPod(h, OperatorNamespace, OperatorName, 200, 0)
})
