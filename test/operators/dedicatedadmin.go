package operators

// This is a test of the Dedicated Admin Operator
// This test checks:
// Operator Namespace exists
// Operator Deployment exists
// Replicas counts match, & pods are running
// ServiceAccount exists
// ConfigMaps exist
// Created Namespace exists
// When a new project is created; that the roleBindings are created in the project
// TODO: any SyncSets exist

import (
	"errors"
	"math/rand"
	"strings"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "github.com/openshift/api/project/v1"
	"github.com/openshift/osde2e/pkg/helper"

	operatorv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Timeout is the duration in seconds that the polling should last
// Because ginkgo requires a float64 in seconds, we do the math
// up here. Below, for the time.Durations, we use seconds because
// we've calculated 30min in seconds up here already.
const globalPollingTimeout int = 30 * 60

const operatorName = "dedicated-admin-operator"
const operatorNamespace string = "openshift-dedicated-admin"

// These variables are not used yet.
// const createdNamespace string = "dedicated-admin"
// const operatorServiceAccount string = "dedicated-admin-operator"
const defaultDesiredReplicas int32 = 1
const operatorLockFile string = "dedicated-admin-operator-lock"
const testProjectPrefix string = "da-test-project"

var clusterRoles = []string{
	"dedicated-admin-operator",
	"dedicated-admin-operator-admin",
	"dedicated-admin-operator-edit",
	"dedicated-admin-operator-view",
	"dedicated-admins-cluster",
	"dedicated-admins-project",
}

var clusterRoleBindings = []string{
	"dedicated-admin-operator-admin",
}

var roleBindings = []string{
	"dedicated-admins-project-0",
	"dedicated-admins-project-1",
}

var _ = ginkgo.Describe("[OSD] Dedicated Admin Operator", func() {
	h := helper.New()
	// Check that the operator clusterServiceVersion exists
	ginkgo.Context("clusterServiceVersion", func() {
		ginkgo.It("should exist", func() {
			csvs, err := pollCsvList(h)
			Expect(err).ToNot(HaveOccurred(), "failed fetching the clusterServiceVersions")
			Expect(csvs).NotTo(BeNil())
			Expect(csvDisplayNameMatch(operatorName, csvs)).Should(BeTrue(),
				"no clusterServiceVersions with .spec.displayName '%v'", operatorName)
		}, float64(globalPollingTimeout))
	})

	// Check that the operator configmap has been deployed
	ginkgo.Context("configmaps", func() {
		ginkgo.It("should exist", func() {
			// Wait for lockfile to signal operator is active
			err := pollLockFile(h)
			Expect(err).ToNot(HaveOccurred(), "failed fetching the configMap lockfile")
		}, float64(globalPollingTimeout))
	})

	// Check that the operator deployment exists in the operator namespace
	ginkgo.Context("deployments", func() {
		ginkgo.It("should exist", func() {
			deployments, err := pollDeploymentList(h)
			Expect(err).ToNot(HaveOccurred(), "failed fetching deployments")
			Expect(deployments).NotTo(BeNil(), "deployment list is nil")
			Expect(deploymentNameMatch(operatorName, deployments)).Should(BeTrue(),
				"no deployments with name '%v'", operatorName)
		}, float64(globalPollingTimeout))
		ginkgo.It("should only be 1", func() {
			expectedDeployments := 1
			deployments, err := pollDeploymentList(h)
			Expect(err).ToNot(HaveOccurred(), "failed fetching deployments")
			Expect(len(deployments.Items)).To(BeNumerically("==", expectedDeployments), "There should be 1 deployment.")
		}, float64(globalPollingTimeout))
		ginkgo.It("should have all desired replicas ready", func() {
			deployments, err := pollDeploymentList(h)
			Expect(err).ToNot(HaveOccurred(), "failed fetching deployments")

			for _, deployment := range deployments.Items {
				readyReplicas := deployment.Status.ReadyReplicas
				desiredReplicas := deployment.Status.Replicas

				// The desired replicas should match the default installed replica count
				Expect(desiredReplicas).To(BeNumerically("==", defaultDesiredReplicas), "The deployment desired replicas should not drift from the default 1.")

				// Desired replica count should match ready replica count
				Expect(readyReplicas).To(BeNumerically("==", desiredReplicas), "All desired replicas should be ready.")
			}
		}, float64(globalPollingTimeout))
	})

	// Check that the clusterRoles exist
	ginkgo.Context("clusterRoles", func() {
		ginkgo.It("should exist", func() {
			for _, clusterRoleName := range clusterRoles {
				_, err := h.Kube().RbacV1().ClusterRoles().Get(clusterRoleName, metav1.GetOptions{})
				Expect(err).ToNot(HaveOccurred(), "failed to get cluster role %v\n", clusterRoleName)
			}
		}, float64(globalPollingTimeout))
	})

	// Check that the clusterRoleBindings exist
	ginkgo.Context("clusterRoleBindings", func() {
		ginkgo.It("should exist", func() {
			for _, clusterRoleBindingName := range clusterRoleBindings {
				_, err := h.Kube().RbacV1().ClusterRoleBindings().Get(clusterRoleBindingName, metav1.GetOptions{})
				Expect(err).ToNot(HaveOccurred(), "failed to get cluster role binding %v\n", clusterRoleBindingName)
			}
		}, float64(globalPollingTimeout))
	})

	// Test the controller; make sure new rolebindings are created for new project
	ginkgo.Context("controller", func() {
		ginkgo.It("should create the expected roleBindings", func() {
			projectRequest := v1.ProjectRequest{}
			projectRequest.Kind = "ProjectRequest"
			projectRequest.APIVersion = "project.openshift.io/v1"
			objectMeta := metav1.ObjectMeta{}
			objectMeta.Name = genSuffix(testProjectPrefix)
			projectRequest.ObjectMeta = objectMeta

			// Create a project; defer deletion of project
			project, err := h.Project().ProjectV1().ProjectRequests().Create(&projectRequest)
			defer func() {
				err := h.Project().ProjectV1().Projects().Delete(project.Name, &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred())
			}()
			Expect(err).NotTo(HaveOccurred())

			// Each Dedicated Admin roleBinding should be added to a newly created project
			for _, roleBindingName := range roleBindings {
				// TODO: Figure out how to use Eventually() for this poller
				// Eventually(pollRoleBinding(h, roleBindingName), 5, 1).Should(Succeed(), "roleBindings should eventually exist")
				// TODO: This would be better with a BeAssignableToTypeOf("whatever a rolebinding type is; not sure how to reference that")

				err := pollRoleBinding(h, project.Name, roleBindingName)
				Expect(err).NotTo(HaveOccurred())
			}
		}, float64(globalPollingTimeout))
	})
})

func pollRoleBinding(h *helper.H, projectName string, roleBindingName string) error {
	// pollRoleBinding will check for the existence of a roleBinding
	// in the specified project, and wait for it to exist, until a timeout

	var err error
	// interval is the duration in seconds between polls
	// values here for humans

	interval := 5

	// convert time.Duration type
	timeoutDuration := time.Duration(globalPollingTimeout) * time.Second
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
				time.Sleep(intervalDuration)
			} else {
				err = errors.New("Failed to get rolebinding %v before timeout, roleBindingName")
				break Loop
			}
		}
	}

	return err
}

func pollLockFile(h *helper.H) error {
	// GetConfigMap polls for a configMap with a timeout
	// to handle the case when a new cluster is up but the OLM has not yet
	// finished deploying the operator

	var err error

	// interval is the duration in seconds between polls
	// values here for humans
	interval := 30

	// convert time.Duration type
	timeoutDuration := time.Duration(globalPollingTimeout) * time.Second
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
				time.Sleep(intervalDuration)
			} else {
				err = errors.New("Failed to get configmap before timeout")
				break Loop
			}
		}
	}

	return err
}

func pollDeploymentList(h *helper.H) (*appsv1.DeploymentList, error) {
	// pollDeploymentList polls for deployments with a timeout
	// to handle the case when a new cluster is up but the OLM has not yet
	// finished deploying the operator

	var err error
	var deploymentList *appsv1.DeploymentList

	// interval is the duration in seconds between polls
	// values here for humans
	interval := 5

	// convert time.Duration type
	timeoutDuration := time.Duration(globalPollingTimeout) * time.Second
	intervalDuration := time.Duration(interval) * time.Second

	start := time.Now()

Loop:
	for {
		deploymentList, err = h.Kube().AppsV1().Deployments(operatorNamespace).List(metav1.ListOptions{})
		elapsed := time.Since(start)

		switch {
		case err == nil:
			// Success
			break Loop
		default:
			if elapsed < timeoutDuration {
				time.Sleep(intervalDuration)
			} else {
				deploymentList = nil
				err = errors.New("Failed to get Deployments before timeout")
				break Loop
			}
		}
	}

	return deploymentList, err
}

func pollCsvList(h *helper.H) (*operatorv1.ClusterServiceVersionList, error) {
	// pollCsvList polls for clusterServiceVersions with a timeout
	// to handle the case when a new cluster is up but the OLM has not yet
	// finished deploying the operator

	var err error
	var csvList *operatorv1.ClusterServiceVersionList

	// interval is the duration in seconds between polls
	// values here for humans
	interval := 5

	// convert time.Duration type
	timeoutDuration := time.Duration(globalPollingTimeout) * time.Second
	intervalDuration := time.Duration(interval) * time.Second

	start := time.Now()

Loop:
	for {
		csvList, err = h.Operator().OperatorsV1alpha1().ClusterServiceVersions(operatorNamespace).List(metav1.ListOptions{})
		elapsed := time.Since(start)

		switch {
		case err == nil:
			// Success
			break Loop
		default:
			if elapsed < timeoutDuration {
				time.Sleep(intervalDuration)
			} else {
				csvList = nil
				err = errors.New("Failed to get clusterServiceVersions before timeout")
				break Loop
			}
		}
	}

	return csvList, err
}

func csvDisplayNameMatch(expected string, csvs *operatorv1.ClusterServiceVersionList) bool {
	// csvDisplayNameMatch iterates a ClusterServiceVersionList
	// and looks for an expected string in the .spec.displayName

	for _, csv := range csvs.Items {
		if expected == csv.Spec.DisplayName {
			return true
		}
	}
	return false
}

func deploymentNameMatch(expected string, deployments *appsv1.DeploymentList) bool {
	// deploymentNameMatch iterates a DeploymentList
	// and looks for an expected string in the .metadata.name

	for _, deployment := range deployments.Items {
		if expected == deployment.GetName() {
			return true
		}
	}
	return false
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
