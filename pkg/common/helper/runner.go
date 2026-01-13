package helper

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/aws"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/mainjobrunner"
)

// RunnerWithNoCommand creates an extended test suite runner and configure RBAC for it.
func (h *H) RunnerWithNoCommand() *mainjobrunner.MainJobRunner {
	r := mainjobrunner.DefaultMainJobRunner.DeepCopy()

	// setup clients
	r.Kube = h.Kube()
	r.Image = h.Image()

	// setup tests
	r.Namespace = h.CurrentProject()
	r.PodSpec.ServiceAccountName = h.GetNamespacedServiceAccount()
	return r
}

// SetRunnerProject sets namespace for runner pod
func (h *H) SetRunnerProject(project string, r *mainjobrunner.MainJobRunner) *mainjobrunner.MainJobRunner {
	r.Namespace = project
	return r
}

// SetRunnerCommand sets given command to a pod runner
func (h *H) SetRunnerCommand(cmd string, r *mainjobrunner.MainJobRunner) *mainjobrunner.MainJobRunner {
	r.Cmd = cmd
	return r
}

// Runner creates an extended test suite runner and configure RBAC for it and runs cmd in it.
func (h *H) Runner(cmd string) *mainjobrunner.MainJobRunner {
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

// UploadResultsToS3 dumps runner results into the s3 bucket in given aws session.
func (h *H) UploadResultsToS3(results map[string][]byte, key string) error {
	for filename, data := range results {
		session, err := aws.CcsAwsSession.GetSession()
		if err != nil {
			return fmt.Errorf("error getting aws session: %v", err)
		}
		aws.WriteToS3Session(session, aws.CreateS3URL(viper.GetString(config.Tests.LogBucket), key, filepath.Base(filename)), data)
	}
	return nil
}
