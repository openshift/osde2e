package operators

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "github.com/openshift/api/project/v1"
	"github.com/openshift/osde2e/pkg/helper"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = ginkgo.FDescribe("[OSD] Dedicated Admin Operator", func() {
	const operatorName = "dedicated-admin-operator"
	const operatorNamespace string = "openshift-dedicated-admin"
	const operatorLockFile string = "dedicated-admin-operator-lock"
	const defaultDesiredReplicas int32 = 1
	const operatorServiceAccount string = "dedicated-admin-operator"

	const createdNamespace string = "dedicated-admin"
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

	h := helper.New()
	// Check that the operator clusterServiceVersion exists
	ginkgo.Context("clusterServiceVersion", func() {
		ginkgo.It("should exist", func() {
			csvs, err := pollCsvList(h, operatorNamespace)
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
			err := pollLockFile(h, operatorNamespace, operatorLockFile)
			Expect(err).ToNot(HaveOccurred(), "failed fetching the configMap lockfile")
		}, float64(globalPollingTimeout))
	})

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
			defer h.Project().ProjectV1().Projects().Delete(project.Name, &metav1.DeleteOptions{})
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
