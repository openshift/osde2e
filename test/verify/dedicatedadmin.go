package verify

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
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "github.com/openshift/api/project/v1"

	"github.com/openshift/osde2e/pkg/helper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const operatorNamespace string = "openshift-dedicated-admin"
const createdNamespace string = "dedicated-admin"
const operatorServiceAccount string = "dedicated-admin-operator"
const defaultDesiredReplicas int32 = 1
const operatorLockFile string = "dedicated-admin-operator-lock"
const testProjectPrefix string = "da-test-project"

var clusterRoles = [6]string{
	"dedicated-admin-operator",
	"dedicated-admin-operator-admin",
	"dedicated-admin-operator-edit",
	"dedicated-admin-operator-view",
	"dedicated-admins-cluster",
	"dedicated-admins-project",
}

var roleBindings = [2]string{
	"dedicated-admins-project-0",
	"dedicated-admins-project-1",
}

// Check that the operator deployment exists in the operator namespace
var _ = ginkgo.Describe("Operator Deployment", func() {
	h := helper.New()

	ginkgo.It("should exist", func() {
		// Get all deployments in the operator namespace
		deployments, err := h.Kube().AppsV1().Deployments(operatorNamespace).List(metav1.ListOptions{})
		Expect(err).To(Succeed(), "failed fetching deployments")
		Expect(deployments).NotTo(BeNil())

		// There should be 1 deployment only
		expectedDeployments := 1
		Expect(len(deployments.Items)).To(BeNumerically("==", expectedDeployments), "There should be 1 deployment.")

		for _, deployment := range deployments.Items {
			readyReplicas := deployment.Status.ReadyReplicas
			desiredReplicas := deployment.Status.Replicas

			// The desired replicas should match the default installed replica count
			Expect(desiredReplicas).To(BeNumerically("==", defaultDesiredReplicas), "The deployment desired replicas should not drift from the default 1.")

			// Desired replica count should match ready replica count
			Expect(readyReplicas).To(BeNumerically("==", desiredReplicas), "All desired replicas should be ready.")
		}
	})
})

// Check that the operator lock file exists
var _ = ginkgo.Describe("Operator lock file", func() {
	h := helper.New()

	ginkgo.It("should exist", func() {
		// The expected lock file should be there as a ConfigMap
		_, err := h.Kube().CoreV1().ConfigMaps(operatorNamespace).Get(operatorLockFile, metav1.GetOptions{})
		Expect(err).To(Succeed(), "failed fetching the lock file ConfigMap")
	})

})

// Check that the expected Cluster Roles exist
// TODO Should check ClusterRoleBindings, but not sure what is supposed to be there
var _ = ginkgo.Describe("Operator Cluster Roles", func() {
	h := helper.New()

	ginkgo.It("should exist", func() {
		// Get the cluster roles by name
		for _, clusterRoleName := range clusterRoles {
			_, err := h.Kube().RbacV1().ClusterRoles().Get(clusterRoleName, metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred(), "failed to get cluster role %d", clusterRoleName)
		}
	})
})

// TODO: Make sure dedicated-admins group is created

// Test the controller; make sure new rolebindings are created for new project
var _ = ginkgo.Describe("The Operator Controller", func() {
	h := helper.New()
	ginkgo.Context("when a new project is created", func() {
		ginkgo.It("should create the expected roleBindings", func() {
			projectRequest := v1.ProjectRequest{}
			projectRequest.Kind = "ProjectRequest"
			projectRequest.APIVersion = "project.openshift.io/v1"
			objectMeta := metav1.ObjectMeta{}
			objectMeta.Name = genSuffix(testProjectPrefix)
			projectRequest.ObjectMeta = objectMeta

			// Create a project; defer deletion of project
			project, err := h.Project().ProjectV1().ProjectRequests().Create(&projectRequest)
			defer h.Project().ProjectV1().Projects().Delete(project.Name, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())

			// Each Dedicated Admin roleBinding should be added to a newly created project
			for _, roleBindingName := range roleBindings {
				// TODO: Figure out how to use Eventually() for this poller
				// Eventually(getRoleBindingByName(roleBindingName, project.Name), 15, 1).Should(Succeed(), "roleBindings should eventually exist")

				// TODO: This would be better with a BeAssignableToTypeOf("whatever a rolebinding type is; not sure how to reference that")

				// Poll 10s to give controller time to create roleBindings
				for i := 0; i < 10; i++ {
					rb, err := h.Kube().RbacV1().RoleBindings(project.Name).Get(roleBindingName, metav1.GetOptions{})
					fmt.Printf("AT POLLER DIGIT %d:\n%v\n", i, rb)
					if err == nil {
						break
					}
					time.Sleep(1)
				}
			}
		})
	})
})

func getRoleBindingByName(name, nameSpace string) error {
	// getRoleBindingByName makes the call to OpenShift to get a roleBinding
	// and returns the error value
	//
	// This is a helper function to strip return results
	// so Gomega's 'Eventually()' can be used for polling
	//
	// TODO: How can we pass the helper from above to this function?
	// It doesn't like the pointer (nil pointer panics) if I pass *helper.H
	// And this instantiation bombs out with "you may only call BeforeEach in a context..." etc
	h := helper.New()
	_, err := h.Kube().RbacV1().RoleBindings(nameSpace).Get(name, metav1.GetOptions{})
	return err
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
