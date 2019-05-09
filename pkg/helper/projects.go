package helper

import (
	"fmt"
	"math/rand"

	projectv1 "github.com/openshift/api/project/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

func randomStr(length int) (str string) {
	chars := "0123456789abcdefghijklmnopqrstuvwxyz"
	for i := 0; i < length; i++ {
		c := string(chars[rand.Intn(len(chars))])
		str += c
	}
	return
}
