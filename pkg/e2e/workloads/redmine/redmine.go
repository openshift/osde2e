package workloads

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/osd/healthchecks"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/common/helper"
	"k8s.io/apimachinery/pkg/util/wait"
)

// Specify where the YAML definitions are for the workloads.
var testDir = "/assets/workloads/e2e/redmine"

// Use the base folder name for the workload name. Make it easy!
var workloadName = filepath.Base(testDir)

var _ = ginkgo.Describe("[Suite: e2e] Workload ("+workloadName+")", func() {
	defer ginkgo.GinkgoRecover()
	// setup helper
	h := helper.New()

	redmineTimeoutInSeconds := 900
	ginkgo.It("should get created in the cluster", func() {
		// Does this workload exist? If so, this must be a repeat run.
		// In this case we should assume the workload has had a valid run once already
		// And simply run another test validating the workload.
		if _, ok := h.GetWorkload(workloadName); ok {
			// Run the workload test
			doTest(h)

		} else {
			// Create all K8s objects that are within the testDir
			objects, err := helper.ApplyYamlInFolder(testDir, h.CurrentProject(), h.Kube())
			Expect(err).NotTo(HaveOccurred(), "couldn't apply k8s yaml")

			// Log how many objects have been created
			log.Printf("%v objects created", len(objects))

			// Give the cluster a second to churn before checking
			time.Sleep(3 * time.Second)

			// Wait for all pods to come up healthy
			err = wait.PollImmediate(15*time.Second, 10*time.Minute, func() (bool, error) {
				// This is pretty basic. Are all the pods up? Cool.
				if check, err := healthchecks.CheckPodHealth(h.Kube().CoreV1()); !check || err != nil {
					return false, nil
				}
				return true, nil
			})
			Expect(err).NotTo(HaveOccurred(), "objects not created in a timely manner")
			// Run the test
			doTest(h)

			// If success, add the workload to the list of installed workloads
			h.AddWorkload(workloadName, h.CurrentProject())
		}

	}, float64(redmineTimeoutInSeconds))
})

func doTest(h *helper.H) {

	// track if error occurs
	var err error

	// duration in seconds between polls
	interval := 5

	// convert time.Duration type
	timeoutDuration := time.Duration(config.Instance.Tests.PollingTimeout) * time.Minute
	intervalDuration := time.Duration(interval) * time.Second

	start := time.Now()

Loop:
	for {
		_, err = h.Kube().CoreV1().Services(h.CurrentProject()).ProxyGet("http", "redmine-frontend", "3000", "/", nil).DoRaw()
		elapsed := time.Since(start)

		switch {
		case err == nil:
			// Success
			break Loop
		default:
			if elapsed < timeoutDuration {
				log.Printf("Waiting %v for application to load", (timeoutDuration - elapsed))
				time.Sleep(intervalDuration)
			} else {
				err = fmt.Errorf("failed to check service before timeout")
				break Loop
			}
		}
	}

	Expect(err).NotTo(HaveOccurred(), "unable to access front end of app")
}
