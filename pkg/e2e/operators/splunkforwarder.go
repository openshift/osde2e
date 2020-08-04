package operators

import (
	"github.com/onsi/ginkgo"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
)

var splunkForwarderBlocking string = "[Suite: operators] [OSD] Splunk Forwarder Operator"
var splunkForwarderInforming string = "[Suite: informing] [OSD] Splunk Forwarder Operator"

func init() {
	alert.RegisterGinkgoAlert(splunkForwarderBlocking, "SD-SREP", "Matt Bargenquast", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
	alert.RegisterGinkgoAlert(splunkForwarderInforming, "SD-SREP", "Matt Bargenquast", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(splunkForwarderBlocking, func() {

	var operatorName = "splunk-forwarder-operator"
	var operatorNamespace string = "openshift-splunk-forwarder-operator"
	var operatorLockFile string = "splunk-forwarder-operator-lock"
	var defaultDesiredReplicas int32 = 1

	var clusterRoleBindings = []string{
		"splunk-forwarder-operator-clusterrolebinding",
	}

	var clusterRoles = []string{
		"splunk-forwarder-operator",
		"splunk-forwarder-operator-og-admin",
		"splunk-forwarder-operator-og-edit",
		"splunk-forwarder-operator-og-view",
	}

	h := helper.New()
	checkClusterServiceVersion(h, operatorNamespace, operatorName)
	checkConfigMapLockfile(h, operatorNamespace, operatorLockFile)
	checkDeployment(h, operatorNamespace, operatorName, defaultDesiredReplicas)
	checkClusterRoleBindings(h, clusterRoleBindings)
	checkClusterRoles(h, clusterRoles)
})

var _ = ginkgo.Describe(splunkForwarderInforming, func() {
	checkUpgrade(helper.New(), "openshift-splunk-forwarder-operator", "openshift-splunk-forwarder-operator",
		"splunk-forwarder-operator.v0.1.157-3dca592",
	)
})
