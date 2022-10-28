package workloads

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	v1 "github.com/openshift/api/route/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/cluster/healthchecks"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/util"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/common/helper"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	// Service name for the Redmine front-end
	redmineSvcName = "redmine-frontend"
	// Service port for the Redmine front-end
	redmineSvcPort = "3000"
)

// Specify where the YAML definitions are for the workloads.
var testDir = "workloads/e2e/redmine"

// Use the base folder name for the workload name. Make it easy!
var workloadName = filepath.Base(testDir)

var testName string = "[Suite: e2e] Workload (" + workloadName + ")"

func init() {
	alert.RegisterGinkgoAlert(testName, "SD-SREP", "Matt Bargenquast", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(testName, func() {
	defer ginkgo.GinkgoRecover()
	// setup helper
	h := helper.New()

	// used for verifying creation of workload pods
	podPrefixes := []string{"redmine"}

	redmineTimeoutInSeconds := 2500
	util.GinkgoIt("should get created in the cluster", func() {

		// Does this workload exist? If so, this must be a repeat run.
		// In this case we should assume the workload has had a valid run once already
		// And simply run another test validating the workload.
		h.SetServiceAccount("")
		if _, ok := h.GetWorkload(workloadName); ok {
			// Run the workload test
			doTest(h)

		} else {
			// Create all K8s objects that are within the testDir
			err := createWorkload(h)
			Expect(err).NotTo(HaveOccurred(), "couldn't create workload")

			// Give the cluster a second to churn before checking
			time.Sleep(3 * time.Second)

			// Wait for all pods to come up healthy
			err = wait.PollImmediate(15*time.Second, 10*time.Minute, func() (bool, error) {
				// This is pretty basic. Are all the pods up? Cool.
				if check, err := healthchecks.CheckPodHealth(h.Kube().CoreV1(), nil, h.CurrentProject(), podPrefixes...); !check || err != nil {
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

func createWorkload(h *helper.H) error {
	// Create all K8s objects that are within the testDir
	objects, err := helper.ApplyYamlInFolder(testDir, h.CurrentProject(), h.Kube())

	// Log how many objects have been created from the workload yaml
	log.Printf("%v objects created", len(objects))

	if err != nil {
		return fmt.Errorf("couldn't apply k8s yaml: %s", err.Error())
	}

	// Create an OpenShift route to go with it
	appRoute := &v1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name: "redmine",
		},
		Spec: v1.RouteSpec{
			To: v1.RouteTargetReference{
				Name: "redmine-frontend",
			},
			TLS: &v1.TLSConfig{Termination: "edge"},
		},
		Status: v1.RouteStatus{},
	}
	_, err = h.Route().RouteV1().Routes(h.CurrentProject()).Create(context.TODO(), appRoute, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("couldn't create application route: %v", err)
	}

	return nil
}

func doTest(h *helper.H) {

	// track if error occurs
	var err error

	// duration in seconds between polls
	interval := 5

	// convert time.Duration type
	timeoutDuration := time.Duration(viper.GetFloat64(config.Tests.PollingTimeout)) * time.Second
	intervalDuration := time.Duration(interval) * time.Second

	err = wait.PollImmediate(intervalDuration, timeoutDuration, func() (bool, error) {
		_, err = h.Kube().CoreV1().Services(h.CurrentProject()).ProxyGet("http", redmineSvcName, redmineSvcPort, "/", nil).DoRaw(context.TODO())
		if err == nil {
			return true, nil
		}
		return false, nil
	})
	Expect(err).NotTo(HaveOccurred(), "unable to access front end of app")
}
