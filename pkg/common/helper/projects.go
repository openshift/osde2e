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

func (h *H) createProject(ctx context.Context, suffix string) (*projectv1.Project, error) {
	proj := &projectv1.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: "osde2e-" + suffix,
		},
	}

	project, err := h.Project().ProjectV1().Projects().Create(ctx, proj, metav1.CreateOptions{})
	if err != nil {
		return project, err
	}

	wait.PollImmediate(5*time.Second, 60*time.Second, func() (done bool, err error) {
		ns, err := h.Kube().CoreV1().Namespaces().Get(ctx, project.Name, metav1.GetOptions{})
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

func (h *H) inspect(ctx context.Context, projects []string) error {
	inspectTimeoutInSeconds := 200
	h.SetServiceAccount(ctx, "system:serviceaccount:%s:cluster-admin")

	// Get a list of projects from the cluster
	clusterProjects, err := h.Project().ProjectV1().Projects().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	// from our list of projects to inspect, build a list of ones which exist on the cluster
	inspectProjects := make([]string, 0)
	for _, clusterProject := range clusterProjects.Items {
		for _, p := range projects {
			if clusterProject.Name == p {
				// Add an 'ns/' prefix for the inspect call
				inspectProjects = append(inspectProjects, "ns/"+clusterProject.Name)
			}
		}
	}

	projectsArg := strings.Join(inspectProjects, " ")
	r := h.Runner(fmt.Sprintf("oc adm inspect %v --dest-dir=%v", projectsArg, runner.DefaultRunner.OutputDir))
	r.Name = "must-gather-additional-projects"
	r.Tarball = true
	stopCh := make(chan struct{})

	err = r.Run(inspectTimeoutInSeconds, stopCh)
	if err != nil {
		return fmt.Errorf("error running project inspection: %s", err.Error())
	}

	gatherResults, err := r.RetrieveResults()
	if err != nil {
		return fmt.Errorf("error retrieving project inspection results: %s", err.Error())
	}

	h.WriteResults(gatherResults)
	return nil
}
