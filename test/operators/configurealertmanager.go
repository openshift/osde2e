package operators

import (
	"github.com/onsi/ginkgo"
	"github.com/openshift/osde2e/pkg/helper"
)

var _ = ginkgo.FDescribe("[OSD] Configure AlertManager Operator", func() {
	const operatorName = "configure-alertmanager-operator"
	const operatorNamespace string = "openshift-monitoring"
	const operatorLockFile string = "configure-alertmanager-operator-lock"
	const defaultDesiredReplicas int32 = 1

	var clusterRoles = []string{
	}

	var clusterRoleBindings = []string{
	}

	h := helper.New()
	checkClusterServiceVersion(h, operatorNamespace, operatorName)
	checkConfigMapLockfile(h, operatorNamespace, operatorLockFile)
	checkDeployment(h, operatorNamespace, operatorName,  defaultDesiredReplicas)
	checkClusterRoles(h, clusterRoles)
	checkClusterRoleBindings(h, clusterRoleBindings)
})
