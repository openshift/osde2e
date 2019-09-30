package operators

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "github.com/openshift/api/project/v1"
	"github.com/openshift/osde2e/pkg/helper"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = ginkgo.Describe("[OSD] Dedicated Admin Operator", func() {
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
	checkClusterServiceVersion(h, operatorNamespace, operatorName)
	checkConfigMapLockfile(h, operatorNamespace, operatorLockFile)
	checkDeployment(h, operatorNamespace, operatorName, defaultDesiredReplicas)
	checkClusterRoles(h, clusterRoles)
	checkClusterRoleBindings(h, clusterRoleBindings)

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
