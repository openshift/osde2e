package helper

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/gomega"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/runner"
)

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
		err = ioutil.WriteFile(dst, data, os.ModePerm)
		Expect(err).NotTo(HaveOccurred())
	}
}
