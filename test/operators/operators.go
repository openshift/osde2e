package operators

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/helper"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
)

const globalPollingTimeout = 30 * 60

func checkClusterServiceVersion(h *helper.H, operatorNamespace, operatorName string) {
	// Check that the operator clusterServiceVersion exists
	ginkgo.Context("clusterServiceVersion", func() {
		ginkgo.It("should exist", func() {
			csvs, err := pollCsvList(h, operatorNamespace, operatorName)
			Expect(err).ToNot(HaveOccurred(), "failed fetching the clusterServiceVersions")
			Expect(csvs).NotTo(BeNil())
		}, float64(globalPollingTimeout))
	})
}

func checkConfigMapLockfile(h *helper.H, operatorNamespace, operatorLockFile string) {
	// Check that the operator configmap has been deployed
	ginkgo.Context("configmaps", func() {
		ginkgo.It("should exist", func() {
			// Wait for lockfile to signal operator is active
			err := pollLockFile(h, operatorNamespace, operatorLockFile)
			Expect(err).ToNot(HaveOccurred(), "failed fetching the configMap lockfile")
		}, float64(globalPollingTimeout))
	})
}

func checkDeployment(h *helper.H, operatorNamespace string, operatorName string, defaultDesiredReplicas int32) {
	// Check that the operator deployment exists in the operator namespace
	ginkgo.Context("deployment", func() {
		ginkgo.It("should exist", func() {
			deployment, err := pollDeployment(h, operatorNamespace, operatorName)
			Expect(err).ToNot(HaveOccurred(), "failed fetching deployment")
			Expect(deployment).NotTo(BeNil(), "deployment is nil")
		}, float64(globalPollingTimeout))
		ginkgo.It("should have all desired replicas ready", func() {
			deployment, err := pollDeployment(h, operatorNamespace, operatorName)
			Expect(err).ToNot(HaveOccurred(), "failed fetching deployment")

			readyReplicas := deployment.Status.ReadyReplicas
			desiredReplicas := deployment.Status.Replicas

			// The desired replicas should match the default installed replica count
			Expect(desiredReplicas).To(BeNumerically("==", defaultDesiredReplicas), "The deployment desired replicas should not drift from the default 1.")

			// Desired replica count should match ready replica count
			Expect(readyReplicas).To(BeNumerically("==", desiredReplicas), "All desired replicas should be ready.")
		}, float64(globalPollingTimeout))
	})
}

func checkClusterRoles(h *helper.H, clusterRoles []string) {
	// Check that the clusterRoles exist
	ginkgo.Context("clusterRoles", func() {
		ginkgo.It("should exist", func() {
			for _, clusterRoleName := range clusterRoles {
				_, err := h.Kube().RbacV1().ClusterRoles().Get(clusterRoleName, metav1.GetOptions{})
				Expect(err).ToNot(HaveOccurred(), "failed to get cluster role %v\n", clusterRoleName)
			}
		}, float64(globalPollingTimeout))
	})
}

func checkClusterRoleBindings(h *helper.H, clusterRoleBindings []string) {
	// Check that the clusterRoleBindings exist
	ginkgo.Context("clusterRoleBindings", func() {
		ginkgo.It("should exist", func() {
			for _, clusterRoleBindingName := range clusterRoleBindings {
				_, err := h.Kube().RbacV1().ClusterRoleBindings().Get(clusterRoleBindingName, metav1.GetOptions{})
				Expect(err).ToNot(HaveOccurred(), "failed to get cluster role binding %v\n", clusterRoleBindingName)
			}
		}, float64(globalPollingTimeout))
	})
}

func checkRoleBindings(h *helper.H, operatorNamespace string, roleBindings []string) {
	// Check that deployed rolebindings exist
	ginkgo.Context("roleBindings", func() {
		ginkgo.It("should exist", func() {
			for _, roleBindingName := range roleBindings {
				err := pollRoleBinding(h, operatorNamespace, roleBindingName)
				Expect(err).NotTo(HaveOccurred(), "failed to get roleBinding %v\n", roleBindingName)
			}
		}, float64(globalPollingTimeout))
	})

}

func pollRoleBinding(h *helper.H, projectName string, roleBindingName string) error {
	// pollRoleBinding will check for the existence of a roleBinding
	// in the specified project, and wait for it to exist, until a timeout

	var err error
	// interval is the duration in seconds between polls
	// values here for humans

	interval := 5

	// convert time.Duration type
	timeoutDuration := time.Duration(globalPollingTimeout*60) * time.Minute
	intervalDuration := time.Duration(interval) * time.Second

	start := time.Now()

Loop:
	for {
		_, err = h.Kube().RbacV1().RoleBindings(projectName).Get(roleBindingName, metav1.GetOptions{})
		elapsed := time.Since(start)

		switch {
		case err == nil:
			// Success
			break Loop
		default:
			if elapsed < timeoutDuration {
				log.Printf("Waiting %v for %s roleBinding to exist", (timeoutDuration - elapsed), roleBindingName)
				time.Sleep(intervalDuration)
			} else {
				err = fmt.Errorf("Failed to get rolebinding %s before timeout", roleBindingName)
				break Loop
			}
		}
	}

	return err
}

func pollLockFile(h *helper.H, operatorNamespace, operatorLockFile string) error {
	// GetConfigMap polls for a configMap with a timeout
	// to handle the case when a new cluster is up but the OLM has not yet
	// finished deploying the operator

	var err error

	// interval is the duration in seconds between polls
	// values here for humans
	interval := 30

	// convert time.Duration type
	timeoutDuration := time.Duration(globalPollingTimeout) * time.Minute
	intervalDuration := time.Duration(interval) * time.Second

	start := time.Now()

Loop:
	for {
		_, err = h.Kube().CoreV1().ConfigMaps(operatorNamespace).Get(operatorLockFile, metav1.GetOptions{})
		elapsed := time.Since(start)

		switch {
		case err == nil:
			// Success
			break Loop
		default:
			if elapsed < timeoutDuration {
				log.Printf("Waiting %v for %s configMap to exist", (timeoutDuration - elapsed), operatorLockFile)
				time.Sleep(intervalDuration)
			} else {
				err = fmt.Errorf("Failed to get configMap %s before timeout", operatorLockFile)
				break Loop
			}
		}
	}

	return err
}

func pollDeployment(h *helper.H, operatorNamespace, deploymentName string) (*appsv1.Deployment, error) {
	// pollDeployment polls for a deployment with a timeout
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
		deployment, err = h.Kube().AppsV1().Deployments(operatorNamespace).Get(deploymentName, metav1.GetOptions{})
		elapsed := time.Since(start)

		switch {
		case err == nil:
			// Success
			break Loop
		default:
			if elapsed < timeoutDuration {
				log.Printf("Waiting %v for %s deployment to exist", (timeoutDuration - elapsed), deploymentName)
				time.Sleep(intervalDuration)
			} else {
				deployment = nil
				err = fmt.Errorf("Failed to get %s Deployment before timeout", deploymentName)
				break Loop
			}
		}
	}

	return deployment, err
}

func pollCsvList(h *helper.H, operatorNamespace, csvDisplayName string) (*operatorv1.ClusterServiceVersionList, error) {
	// pollCsvList polls for clusterServiceVersions with a timeout
	// to handle the case when a new cluster is up but the OLM has not yet
	// finished deploying the operator

	var err error
	var csvList *operatorv1.ClusterServiceVersionList

	// interval is the duration in seconds between polls
	// values here for humans
	interval := 5

	// convert time.Duration type
	timeoutDuration := time.Duration(globalPollingTimeout) * time.Minute
	intervalDuration := time.Duration(interval) * time.Second

	start := time.Now()

Loop:
	for {
		csvList, err = h.Operator().OperatorsV1alpha1().ClusterServiceVersions(operatorNamespace).List(metav1.ListOptions{})
		for _, csv := range csvList.Items {
			switch {
			case csvDisplayName == csv.Spec.DisplayName:
				// Success
				err = nil
			default:
				err = fmt.Errorf("No matching clusterServiceVersion in CSV List")
			}
		}
		elapsed := time.Since(start)

		switch {
		case err == nil:
			// Success
			break Loop
		default:
			if elapsed < timeoutDuration {
				log.Printf("Waiting %v for %s clusterServiceVersion to exist", (timeoutDuration - elapsed), csvDisplayName)
				time.Sleep(intervalDuration)
			} else {
				csvList = nil
				err = fmt.Errorf("Failed to get %s clusterServiceVersion before timeout", csvDisplayName)
				break Loop
			}
		}
	}

	return csvList, err
}

func genSuffix(prefix string) string {
	// genSuffix creates a random 8 character string to append to object
	// names when creating Kubernetes objects so there aren't any
	// accidental name collisions

	// Seed rand so there's actual randomness
	// otherwise the string is always the same
	rand.Seed(time.Now().UnixNano())

	bytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		bytes[i] = byte(65 + rand.Intn(25))
	}
	return prefix + "-" + strings.ToLower(string(bytes))
}
