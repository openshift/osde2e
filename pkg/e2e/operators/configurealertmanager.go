package operators

import (
	"github.com/onsi/ginkgo"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
)

var configureAlertManagerOperators string = "[Suite: operators] [OSD] Configure AlertManager Operator"

func init() {
	alert.RegisterGinkgoAlert(configureAlertManagerOperators, "SD-SREP", "Christopher Collins", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(configureAlertManagerOperators, func() {
	var operatorName = "configure-alertmanager-operator"
	var operatorNamespace string = "openshift-monitoring"
	var operatorLockFile string = "configure-alertmanager-operator-lock"
	var defaultDesiredReplicas int32 = 1

	var clusterRoles = []string{
		"configure-alertmanager-operator",
	}

	var clusterRoleBindings = []string{}

	var roleBindings = []string{
		"configure-alertmanager-operator",
	}

	var roles = []string{
		"configure-alertmanager-operator",
	}

	h := helper.New()
	checkClusterServiceVersion(h, operatorNamespace, operatorName)
	checkConfigMapLockfile(h, operatorNamespace, operatorLockFile)
	checkDeployment(h, operatorNamespace, operatorName, defaultDesiredReplicas)
	checkClusterRoles(h, clusterRoles, true)
	checkClusterRoleBindings(h, clusterRoleBindings, true)
	checkRole(h, operatorNamespace, roles)
	checkRoleBindings(h, operatorNamespace, roleBindings)
	checkUpgrade(helper.New(), "openshift-monitoring", "configure-alertmanager-operator",
		"configure-alertmanager-operator", "configure-alertmanager-operator-registry")
})
