package helper

import (
	"context"
	"fmt"

	projectv1 "github.com/openshift/api/project/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (h *H) createProject(suffix string) (*projectv1.Project, error) {
	proj := &projectv1.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: "osde2e-" + suffix,
		},
	}
	return h.Project().ProjectV1().Projects().Create(context.TODO(), proj, metav1.CreateOptions{})
}

func (h *H) cleanup(projectName string) error {
	err := h.Project().ProjectV1().Projects().Delete(context.TODO(), projectName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to cleanup project '%s': %v", projectName, err)
	}

	err = h.Kube().CoreV1().ServiceAccounts("dedicated-admin").Delete(context.TODO(), h.CurrentProject(), metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to cleanup sa '%s': %v", projectName, err)
	}

	return nil
}
