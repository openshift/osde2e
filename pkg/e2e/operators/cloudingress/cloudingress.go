package cloudingress

import (
	"github.com/openshift/osde2e/pkg/common/alert"
)

const (
	CloudIngressNamespace         = "openshift-cloud-ingress-operator"
	CloudIngressTestName          = "[Suite: operators] [OSD] Cloud Ingress Operator"
	CloudIngressInformingTestName = "[Suite: informing] [OSD] Cloud Ingress Operator"
	OperatorName                  = "cloud-ingress-operator"
)

// utils
func init() {
	alert.RegisterGinkgoAlert(CloudIngressTestName, "SD-SRE", "Alex Chvatal", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
	alert.RegisterGinkgoAlert(CloudIngressInformingTestName, "SD-SRE", "Alex Chvatal", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}
