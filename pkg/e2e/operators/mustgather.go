package operators

import (
	"github.com/onsi/ginkgo"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
)

var mustGatherOperatorTest = "[Suite: operators] [OSD] Must Gather Operator"

func init() {
	alert.RegisterGinkgoAlert(mustGatherOperatorTest, "SD-SREP", "Arjun Naik", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(mustGatherOperatorTest, func() {
	var operatorName = "must-gather-operator"
	var operatorNamespace = "openshift-must-gather-operator"
	var operatorLockFile = "must-gather-operator-lock"
	var defaultDesiredReplicas int32 = 1

	var clusterRoles = []string{
		"must-gather-operator-admin",
		"must-gather-operator-edit",
		"must-gather-operator-view",
	}

	h := helper.New()
	checkClusterServiceVersion(h, operatorNamespace, operatorName)
	checkConfigMapLockfile(h, operatorNamespace, operatorLockFile)
	checkDeployment(h, operatorNamespace, operatorName, defaultDesiredReplicas)
	checkClusterRoles(h, clusterRoles, false)

	checkUpgrade(helper.New(), "openshift-must-gather-operator", "must-gather-operator",
		"must-gather-operator", "must-gather-operator-registry")
})
