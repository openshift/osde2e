package helper

import (
	"fmt"
	"math/rand"

	. "github.com/onsi/gomega"

	projectv1 "github.com/openshift/api/project/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GiveCurrentProjectClusterAdmin to default service account and ensure its removed after project deletion.
func (h *H) GiveCurrentProjectClusterAdmin() {
	// use OwnerReference of project to ensure deletion
	Expect(h.proj).NotTo(BeNil())
	gvk := schema.FromAPIVersionAndKind("project.openshift.io/v1", "Project")
	projRef := *metav1.NewControllerRef(h.proj, gvk)

	// create binding with OwnerReference
	_, err := h.Kube().RbacV1().ClusterRoleBindings().Create(&rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "osde2e-test-access-",
			OwnerReferences: []metav1.OwnerReference{
				projRef,
			},
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      "default",
				Namespace: h.CurrentProject(),
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     "cluster-admin",
		},
	})
	Expect(err).NotTo(HaveOccurred(), "couldn't set correct permissions for OpenShift E2E")
}

func (h *H) createProject(suffix string) (*projectv1.Project, error) {
	proj := &projectv1.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: "osde2e-" + suffix,
		},
	}
	return h.Project().ProjectV1().Projects().Create(proj)
}

func (h *H) cleanup(projectName string) error {
	err := h.Project().ProjectV1().Projects().Delete(projectName, &metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to cleanup project '%s': %v", projectName, err)
	}
	return nil
}

func RandomStr(length int) (str string) {
	chars := "0123456789abcdefghijklmnopqrstuvwxyz"
	for i := 0; i < length; i++ {
		c := string(chars[rand.Intn(len(chars))])
		str += c
	}
	return
}
