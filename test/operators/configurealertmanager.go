package operators

// This is a test of the Configure Alertmanager Operator
// Currently, this test just checks for the existence of the deployment

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "github.com/openshift/api/project/v1"
	"github.com/openshift/osde2e/pkg/helper"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// timeout is the duration in minutes that the polling should last
const globalPollingTimeout int = 30

// operator-specific values
const operatorNamespace = "openshift-monitoring"
const operatorDeploymentName = "configure-alertmanager-operator"

var _ = ginkgo.FDescribe("[OSD] configure-alertmanager-operator", func() {
	h := helper.New()

	// Check that the operator deployment exists in the operator namespace
	ginkgo.Context("deployment", func() {
		ginkgo.It("should exist", func() {
			fmt.Printf("CONFIG ###################\n")
			fmt.Printf("%v\n", cfg)
			fmt.Printf("CONFIG ###################\n")
			deployments, err := pollDeployment(h)

			Expect(err).ToNot(HaveOccurred(), "failed fetching deployment")
		})
	})

func pollDeployment(h *helper.H) (*appsv1.Deployment, error) {
	// pollDeployment polls for the operator deployment until a timeout
	// to handle the case when a new cluster is up but the OLM has not yet
	// finished deploying the operator

	var err error
	var deployment *appsv1.Deployment

	// interval is the duration in seconds between polls
	// values here for humans
	interval := 5

	// convert time.Duration type
	timeoutDuration := time.Duration(globalPollingTimeout) * time.Minute
	intervalDuration := time.Duration(interval) * time.Second

	start := time.Now()

Loop:
	for {
		deployment, err = h.Kube().AppsV1().Deployments(operatorNamespace).Get(operatorDeploymentName, metav1.GetOptions)
		elapsed := time.Now().Sub(start)

		switch {
		case err == nil:
			// Success
			break Loop
		default:
			if elapsed < timeoutDuration {
				timeTilTimeout := timeoutDuration - elapsed
				log.Printf("Failed to get %v deployment, will retry (timeout in: %v", operatorDeploymentName, timeTilTimeout)
				time.Sleep(intervalDuration)
			} else {
				log.Printf("Failed to get %v deployments before timeout, failing", operatorDeploymentName)
				break Loop
			}
		}
	}

	return deployment, err
}