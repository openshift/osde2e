package workloads

import (
	"context"
	"log"
	"path/filepath"
	"time"

	v1 "github.com/openshift/api/route/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/cluster/healthchecks"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/util"
	"k8s.io/apimachinery/pkg/util/wait"
)

// Specify where the YAML definitions are for the workloads.
var testDir = "/assets/workloads/e2e/guestbook"

// Use the base folder name for the workload name. Make it easy!
var workloadName = filepath.Base(testDir)

var testName string = "[Suite: e2e] Workload (" + workloadName + ")"

func init() {
	alert.RegisterGinkgoAlert(testName, "SD-CICD", "Jeffrey Sica", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(testName, func() {
	defer ginkgo.GinkgoRecover()
	// setup helper
	h := helper.New()

	// used for verifying creation of workload pods
	podPrefixes := []string{"frontend", "redis-master", "redis-slave"}

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
			objects, err := helper.ApplyYamlInFolder(testDir, h.CurrentProject(), h.Kube())
			Expect(err).NotTo(HaveOccurred(), "couldn't apply k8s yaml")

			// Log how many objects have been created
			log.Printf("%v objects created", len(objects))

			// Create an OpenShift route to go with it
			appRoute := &v1.Route{
				ObjectMeta: metav1.ObjectMeta{
					Name: "guestbook",
				},
				Spec: v1.RouteSpec{
					To: v1.RouteTargetReference{
						Name: "frontend",
					},
					TLS: &v1.TLSConfig{Termination: "edge"},
				},
				Status: v1.RouteStatus{},
			}
			_, err = h.Route().RouteV1().Routes(h.CurrentProject()).Create(context.TODO(), appRoute, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred(), "couldn't create application route")

			// Give the cluster a second to churn before checking
			time.Sleep(3 * time.Second)

			// Wait for all pods to come up healthy
			err = wait.PollImmediate(15*time.Second, 5*time.Minute, func() (bool, error) {
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

	})
})

func doTest(h *helper.H) {
	_, err := h.Kube().CoreV1().Services(h.CurrentProject()).ProxyGet("http", "frontend", "3000", "/", nil).DoRaw(context.TODO())
	Expect(err).NotTo(HaveOccurred(), "unable to access front end of app")
}
