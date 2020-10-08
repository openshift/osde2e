package cloudingress

import (
	"github.com/openshift/osde2e/pkg/common/alert"
)

const (
	CloudIngressNamespace = "openshift-ingress-operator"
	CloudIngressTestName  = "[Suite: informing] CloudIngressOperator"
)

// utils
func init() {
	alert.RegisterGinkgoAlert(CloudIngressTestName, "SD-SRE", "Alex Chvatal", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}
