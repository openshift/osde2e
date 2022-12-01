package cloudingress

import (
	"github.com/openshift/osde2e/pkg/common/alert"
)

const (
	TestPrefix            = "CloudIngressOperator"
	OperatorName          = "cloud-ingress-operator"
	OperatorNamespace     = "openshift-cloud-ingress-operator"
	apiSchemeResourceName = "rh-api"
)

// utils
func init() {
	alert.RegisterGinkgoAlert("[Suite: informing] "+TestPrefix, "SD-SRE", "@sd-sre-aurora-team", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
	alert.RegisterGinkgoAlert("[Suite: operators] "+TestPrefix, "SD-SRE", "@sd-sre-aurora-team", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}
