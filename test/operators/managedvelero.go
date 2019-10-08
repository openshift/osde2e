package operators

import (
	"github.com/onsi/ginkgo"
	"github.com/openshift/osde2e/pkg/helper"
)

var _ = ginkgo.Describe("[OSD] Managed Velero Operator", func() {
	var operatorName = "managed-velero-operator"
	var operatorNamespace string = "openshift-velero"
	var operatorLockFile string = "managed-velero-operator-lock"
	var defaultDesiredReplicas int32 = 1

	var clusterRoles = []string{
		"managed-velero-operator",
	}

	var clusterRoleBindings = []string{
		"managed-velero-operator",
		"velero",
	}

	h := helper.New()
	checkConfigMapLockfile(h, operatorNamespace, operatorLockFile)
	checkDeployment(h, operatorNamespace, operatorName, defaultDesiredReplicas)
	checkDeployment(h, operatorNamespace, "velero", defaultDesiredReplicas)
	checkClusterRoles(h, clusterRoles)
	checkClusterRoleBindings(h, clusterRoleBindings)
	checkRoleBindings(h,
		operatorNamespace,
		[]string{"managed-velero-operator"})
	checkRoleBindings(h,
		"kube-system",
		[]string{"managed-velero-operator-cluster-config-v1-reader"})
})
