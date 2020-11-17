package operators

import (
	"context"
	"fmt"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	api "github.com/openshift/rbac-permissions-operator/pkg/apis/managed/v1alpha1"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	unstruct "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var operatorName = "rbac-permissions-operator"
var operatorNamespace = "openshift-rbac-permissions"
var rbacOperatorBlocking string = "[Suite: operators] [OSD] RBAC Operator"
var subjectPermissionsTestName string = "[Suite: operators] [OSD] RBAC Dedicated Admins SubjectPermission"

func init() {
	alert.RegisterGinkgoAlert(rbacOperatorBlocking, "SD-SREP", "Matt Bargenquast", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
	alert.RegisterGinkgoAlert(subjectPermissionsTestName, "SD-SREP", "Matt Bargenquast", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(rbacOperatorBlocking, func() {
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
	checkClusterRoles(h, clusterRoles, false)

	checkUpgrade(helper.New(), "openshift-rbac-permissions", "rbac-permissions-operator",
		"rbac-permissions-operator", "rbac-permissions-operator-registry")
})

var _ = ginkgo.Describe(subjectPermissionsTestName, func() {
	h := helper.New()
	checkSubjectPermissions(h, "dedicated-admins")

	ginkgo.Context("dedicated-admin-subjectpermission", func() {
		ginkgo.It("For dedicated admin access should be forbidden to create Subjectpermissions", func() {
			h.SetServiceAccount("system:serviceaccount:%s:dedicated-admin-project")
			sp := makeSubjectPermission("osde2e-dedicated-admins-cluster")
			err := createSubjectPermission(sp, operatorNamespace, h)
			Expect(err).To(HaveOccurred())
		})
	})

	ginkgo.Context("dedicated-admin-subjectpermission", func() {
		ginkgo.It("For cluster admin access should be allowed to create Subjectpermissions", func() {
			sp := makeSubjectPermission("osde2e-cluster-admins")
			err := createSubjectPermission(sp, operatorNamespace, h)
			Expect(err).NotTo(HaveOccurred())

		})
	})
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

func makeSubjectPermission(name string) api.SubjectPermission {
	sp := api.SubjectPermission{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SubjectPermission",
			APIVersion: "managed.openshift.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: operatorNamespace,
		},
		Spec: api.SubjectPermissionSpec{
			SubjectName:        name,
			ClusterPermissions: []string{"sre-admins-cluster"},
			Permissions: []api.Permission{
				{
					ClusterRoleName:        "sre-admins-project",
					NamespacesAllowedRegex: "^(default|openshift.*|kube.*)$",
					AllowFirst:             true,
				},
			},
		},
	}

	return sp
}

func createSubjectPermission(sp api.SubjectPermission, operatorNamespace string, h *helper.H) error {
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(sp.DeepCopy())
	if err != nil {
		return err
	}
	unstructuredObj := unstructured.Unstructured{obj}
	_, err = h.Dynamic().Resource(schema.GroupVersionResource{
		Group:    "managed.openshift.io",
		Version:  "v1alpha1",
		Resource: "subjectpermissions"}).Namespace(operatorNamespace).Create(context.TODO(), &unstructuredObj, metav1.CreateOptions{})
	return err
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
