package operators

import (
	"github.com/onsi/ginkgo"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
)

func init() {
	ma := alert.GetMetricAlerts()
	testAlert = alert.MetricAlert{
		Name:             "[Suite: operators] [OSD] Splunk Forwarder Operator",
		TeamOwner:        "SD-SREP",
		PrimaryContact:   "Matt Bargenquast",
		SlackChannel:     "sd-cicd-alerts",
		Email:            "sd-cicd@redhat.com",
		FailureThreshold: 4,
	}
	ma.AddAlert(testAlert)
}

var _ = ginkgo.Describe(testAlert.Name, func() {

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

var _ = ginkgo.Describe(testAlert.Name+" Upgrade", func() {
	checkUpgrade(helper.New(), "openshift-splunk-forwarder-operator", "openshift-splunk-forwarder-operator",
		"splunk-forwarder-operator.v0.1.157-3dca592",
	)
})
