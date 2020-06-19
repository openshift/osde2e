// Package openshift runs the OpenShift extended test suite.
package openshift

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/runner"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BuildE2EConfig is the base configuration for the E2E run.
var BuildE2EConfig = E2EConfig{
	TestCmd: "run",
	Suite:   "openshift/image-ecosystem",
	Env: []string{
		"DELETE_NAMESPACE=false",
	},
	Flags: []string{
		"--include-success",
		"--junit-dir=" + runner.DefaultRunner.OutputDir,
	},
}

var testApplications = []string{
	"django-psql",
	"rails-postgresql",
	// TODO: The following applications rely on an imagestream not present until at least v4.3.5
	// "cakephp-mysql",
	// "nodejs-mongodb",
}

var _ = ginkgo.Describe("[Suite: app-builds] OpenShift Application Build E2E", func() {
	defer ginkgo.GinkgoRecover()

	h := helper.New()

	e2eTimeoutInSeconds := 3600
	ginkgo.It("should get created in the cluster", func() {

		namespacesExist := false
		for _, application := range testApplications {
			_, err := findAppNamespace(h, application)
			if err == nil {
				namespacesExist = true
				break
			}
		}

		if namespacesExist {

			// The job namespaces are present, indicating that the test has run once
			// In this case, just verify the healthy state of the applications

			log.Printf("Existing applications detected, will verify rather than build.")
			for _, applicationName := range testApplications {
				appNamespace, err := findAppNamespace(h, applicationName)
				Expect(err).NotTo(HaveOccurred())

				list, err := h.Kube().CoreV1().Pods(appNamespace).List(context.TODO(), metav1.ListOptions{
					FieldSelector: fmt.Sprintf("status.phase=%s", kubev1.PodFailed),
				})
				Expect(err).NotTo(HaveOccurred(), "couldn't list Pods")
				Expect(list).NotTo(BeNil())
				Expect(list.Items).Should(HaveLen(0), "'%d' Pods are 'Failed'", len(list.Items))

			}

		} else {

			// The applications do not exist, so test the successful build of them.
			// configure tests
			cfg := BuildE2EConfig
			// Add run flags for the testing apps
			cfg.Flags = append(cfg.Flags, "--run \"Building ("+strings.Join(testApplications, "|")+") app\"")

			cmd := cfg.Cmd()

			// setup runner
			r := h.Runner(cmd)

			r.Name = "openshift-tests"

			// run tests
			stopCh := make(chan struct{})
			err := r.Run(e2eTimeoutInSeconds, stopCh)
			Expect(err).NotTo(HaveOccurred())

			// get results
			results, err := r.RetrieveResults()
			Expect(err).NotTo(HaveOccurred())

			// write results
			h.WriteResults(results)
		}

	}, float64(e2eTimeoutInSeconds+30))

})

func findAppNamespace(h *helper.H, appName string) (string, error) {

	namespaceRegex := regexp.MustCompile("e2e-test-" + appName + "-repo-test-\\w+")
	namespaceList, err := h.Project().ProjectV1().Projects().List(context.TODO(), metav1.ListOptions{})
	//h.Kube().CoreV1().Namespaces().List()
	if err != nil {
		err = fmt.Errorf("failed to fetch namespaces")
	}

	foundNamespace := ""
	for _, namespace := range namespaceList.Items {
		if namespaceRegex.MatchString(namespace.Name) {
			foundNamespace = namespace.Name
		}
	}
	if foundNamespace == "" {
		err = fmt.Errorf("no matching namespace found for %s", appName)
	}

	return foundNamespace, err
}
