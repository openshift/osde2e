package helper

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/runner"
)

// Runner creates an extended test suite runner and configure RBAC for it.
func (h *H) Runner() *runner.Runner {
	h.GiveCurrentProjectClusterAdmin()
	r := *runner.DefaultRunner

	// setup clients
	r.Kube = h.Kube()
	r.Image = h.Image()

	// setup tests
	r.Namespace = h.CurrentProject()
	return &r
}

// WriteResults dumps runner results into the ReportDir.
func (h *H) WriteResults(results map[string][]byte) {
	for filename, data := range results {
		dst := filepath.Join(h.ReportDir, filename)
		err := ioutil.WriteFile(dst, data, os.ModePerm)
		Expect(err).NotTo(HaveOccurred())
	}
}
