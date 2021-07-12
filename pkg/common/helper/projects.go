package helper

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/openshift/osde2e/pkg/common/runner"

	projectv1 "github.com/openshift/api/project/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func (h *H) createProject(suffix string) (*projectv1.Project, error) {
	proj := &projectv1.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: "osde2e-" + suffix,
		},
	}

	project, err := h.Project().ProjectV1().Projects().Create(context.TODO(), proj, metav1.CreateOptions{})

	if err != nil {
		return project, err
	}

	wait.PollImmediate(5*time.Second, 60*time.Second, func() (done bool, err error) {
		ns, err := h.Kube().CoreV1().Namespaces().Get(context.TODO(), project.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		if ns != nil {
			return true, nil
		}

		return false, nil
	})

	return project, err
}

func (h *H) inspect(projects []string) error {
	inspectTimeoutInSeconds := 200
	h.SetServiceAccount("system:serviceaccount:%s:cluster-admin")

	// add "ns/" prefix to each project
	for i, project := range projects {
		projects[i] = "ns/" + project
	}
	projectsArg := strings.Join(projects, " ")
	r := h.Runner(fmt.Sprintf("oc adm inspect %v --dest-dir=%v", projectsArg, runner.DefaultRunner.OutputDir))
	r.Name = "must-gather-additional-projects"
	r.Tarball = true
	stopCh := make(chan struct{})

	err := r.Run(inspectTimeoutInSeconds, stopCh)
	if err != nil {
		return fmt.Errorf("Error running project inspection: %s", err.Error())
	}

	gatherResults, err := r.RetrieveResults()
	if err != nil {
		return fmt.Errorf("Error retrieving project inspection results: %s", err.Error())
	}

	h.WriteResults(gatherResults)
	return nil
}
