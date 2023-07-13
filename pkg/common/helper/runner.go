package helper

import (
	"os"
	"path/filepath"

	. "github.com/onsi/gomega"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/templates"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/runner"
)

// RunnerWithTemplateCommand creates an extended test suite runner for templated test harness pods, configures RBAC for it.
func (h *H) RunnerWithTemplateCommand(timeout float64, harness string, suffix string, jobName string, serviceAccountDir string) *runner.Runner {
	r := h.RunnerWithNoCommand()
	r.PodSpec.ServiceAccountName = h.GetNamespacedServiceAccount()
	latestImageStream, err := r.GetLatestImageStreamTag()
	Expect(err).NotTo(HaveOccurred(), "Could not get latest imagestream tag")
	cmd, err := getCommandString(h, timeout, latestImageStream, harness, suffix, jobName, serviceAccountDir)
	Expect(err).NotTo(HaveOccurred(), "Could not create pod command from template")
	r.Cmd = cmd
	return r
}

// RunnerWithNoCommand creates an extended test suite runner and configure RBAC for it.
func (h *H) RunnerWithNoCommand() *runner.Runner {
	r := runner.DefaultRunner.DeepCopy()

	// setup clients
	r.Kube = h.Kube()
	r.Image = h.Image()

	// setup tests
	r.Namespace = h.CurrentProject()
	r.PodSpec.ServiceAccountName = h.GetNamespacedServiceAccount()
	return r
}

// Runner creates an extended test suite runner and configure RBAC for it and runs cmd in it.
func (h *H) Runner(cmd string) *runner.Runner {
	r := h.RunnerWithNoCommand()
	r.PodSpec.ServiceAccountName = h.GetNamespacedServiceAccount()
	r.Cmd = cmd
	return r
}

// WriteResults dumps runner results into the ReportDir.
func (h *H) WriteResults(results map[string][]byte) {
	for filename, data := range results {
		dst := filepath.Join(viper.GetString(config.ReportDir), viper.GetString(config.Phase), filename)
		err := os.MkdirAll(filepath.Dir(dst), os.FileMode(0o755))
		Expect(err).NotTo(HaveOccurred())
		err = os.WriteFile(dst, data, os.ModePerm)
		Expect(err).NotTo(HaveOccurred())
	}
}

// Generates templated command string to provide to test harness container
func getCommandString(h *H, timeout float64, latestImageStream string, harness string, suffix string, jobName string, serviceAccountDir string) (string, error) {
	values := struct {
		Name                 string
		JobName              string
		Arguments            string
		Timeout              int
		Image                string
		OutputDir            string
		ServiceAccount       string
		PushResultsContainer string
		Suffix               string
		Server               string
		CA                   string
		TokenFile            string
		EnvironmentVariables []struct {
			Name  string
			Value string
		}
	}{
		Name:                 jobName,
		JobName:              jobName,
		Timeout:              int(timeout),
		Image:                harness,
		OutputDir:            runner.DefaultRunner.OutputDir,
		ServiceAccount:       h.GetNamespacedServiceAccount(),
		PushResultsContainer: latestImageStream,
		Suffix:               suffix,
		Server:               "https://kubernetes.default",
		CA:                   serviceAccountDir + "/ca.crt",
		TokenFile:            serviceAccountDir + "/token",
		EnvironmentVariables: []struct {
			Name  string
			Value string
		}{
			{
				Name:  "OCM_CLUSTER_ID",
				Value: viper.GetString(config.Cluster.ID),
			},
		},
	}
	testTemplate, err := templates.LoadTemplate("tests/tests-runner.template")
	Expect(err).NotTo(HaveOccurred(), "Could not load pod template")
	return h.ConvertTemplateToString(testTemplate, values)
}
