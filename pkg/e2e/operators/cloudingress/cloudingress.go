package cloudingress

import (
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/constants"
)

const (
	TestPrefix            = "CloudIngressOperator"
	OperatorName          = "cloud-ingress-operator"
	OperatorNamespace     = "openshift-cloud-ingress-operator"
	apiSchemeResourceName = "rh-api"
)

// utils
func init() {
	alert.RegisterGinkgoAlert(constants.SuiteInforming+TestPrefix, "SD-SRE", "@sd-sre-aurora-team", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
	alert.RegisterGinkgoAlert(constants.SuiteOperators+TestPrefix, "SD-SRE", "@sd-sre-aurora-team", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}
