package operators

import (
	"context"
	"fmt"

	"github.com/onsi/ginkgo"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/runtime/schema"

	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/common/helper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	unstruct "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var operatorName = "rbac-permissions-operator"
var operatorNamespace = "openshift-rbac-permissions"

var _ = ginkgo.Describe("[Suite: operators] [OSD] RBAC Operator", func() {
	var operatorLockFile = "rbac-permissions-operator-lock"
	var defaultDesiredReplicas int32 = 1

	var clusterRoles = []string{
		"rbac-permissions-operator-admin",
		"rbac-permissions-operator-edit",
		"rbac-permissions-operator-view",
	}

	h := helper.New()
	checkClusterServiceVersion(h, operatorNamespace, operatorName)
	checkConfigMapLockfile(h, operatorNamespace, operatorLockFile)
	checkDeployment(h, operatorNamespace, operatorName, defaultDesiredReplicas)
	checkClusterRoles(h, clusterRoles)
})

var _ = ginkgo.Describe("[Suite: operators] [OSD] Dedicated Admins SubjectPermission", func() {
	h := helper.New()
	checkSubjectPermissions(h, "dedicated-admins")
})

var _ = ginkgo.Describe("[Suite: informing] [OSD] Upgrade RBAC Permissions Operator", func() {
	checkUpgrade(helper.New(), "openshift-rbac-permissions", "rbac-permissions-operator",
		"rbac-permissions-operator.v0.1.97-68cf185",
	)
})

func checkSubjectPermissions(h *helper.H, spName string) {
	ginkgo.Context("SubjectPermission", func() {
		ginkgo.It("should have the expected ClusterRoles, ClusterRoleBindings and RoleBindinsg", func() {
			clusterRoles, clusterRoleBindings, roleBindings, err := getSubjectPermissionRBACInfo(h, spName)
			Expect(err).NotTo(HaveOccurred())

			for _, clusterRoleName := range clusterRoles {
				_, err := h.Kube().RbacV1().ClusterRoles().Get(context.TODO(), clusterRoleName, metav1.GetOptions{})
				Expect(err).ToNot(HaveOccurred(), "failed to get clusterRole %v\n", clusterRoleName)
			}

			for _, clusterRoleBindingName := range clusterRoleBindings {
				err := pollClusterRoleBinding(h, clusterRoleBindingName)
				Expect(err).ToNot(HaveOccurred(), "failed to get clusterRoleBinding %v\n", clusterRoleBindingName)
			}
			for _, roleBindingName := range roleBindings {
				err := pollRoleBinding(h, h.CurrentProject(), roleBindingName)
				Expect(err).NotTo(HaveOccurred(), "failed to get roleBinding %v\n", roleBindingName)
			}

		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
}

func getSubjectPermissionRBACInfo(h *helper.H, spName string) ([]string, []string, []string, error) {
	us, err := h.Dynamic().Resource(schema.GroupVersionResource{
		Group:    "managed.openshift.io",
		Version:  "v1alpha1",
		Resource: "subjectpermissions"}).Namespace(operatorNamespace).Get(context.TODO(), spName, metav1.GetOptions{})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error getting %s SubjectPermission", spName)
	}

	subjectName, _, err := unstruct.NestedString(us.Object, "spec", "subjectName")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error getting subjectName")
	}

	clusterRoles, _, err := unstruct.NestedStringSlice(us.Object, "spec", "clusterPermissions")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error getting clusterPermissions")
	}

	var clusterRoleBindings = []string{}
	for _, crName := range clusterRoles {
		clusterRoleBindings = append(clusterRoleBindings, crName+"-"+subjectName)
	}

	permissions, _, err := unstruct.NestedSlice(us.Object, "spec", "permissions")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error getting permissions")
	}

	var roleBindings = []string{}
	for _, p := range permissions {
		perm := p.(map[string]interface{})
		cpName, _, err := unstruct.NestedString(perm, "clusterRoleName")
		if err != nil {
			return nil, nil, nil, fmt.Errorf("Error getting permission clusterRoleName")
		}
		clusterRoles = append(clusterRoles, cpName)
		roleBindings = append(roleBindings, cpName+"-"+subjectName)
	}

	return clusterRoles, clusterRoleBindings, roleBindings, nil
}
