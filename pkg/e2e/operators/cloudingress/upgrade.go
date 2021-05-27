package cloudingress

import (
	"github.com/onsi/ginkgo"
	"github.com/openshift/osde2e/pkg/common/constants"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/e2e/operators"
)

var _ = ginkgo.Describe(constants.SuiteInforming+TestPrefix, func() {
	h := helper.New()
	operators.CheckUpgrade(h, "openshift-cloud-ingress-operator", "cloud-ingress-operator", "cloud-ingress-operator", "cloud-ingress-operator-registry")
	operators.CheckPod(h, "openshift-cloud-ingress-operator", "cloud-ingress-operator", 200, 0)
})
