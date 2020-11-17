package verify

import (
	"context"
	"github.com/onsi/ginkgo"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/gomega"
)

var dedicatedAdminSccTestName = "[Suite: e2e] [OSD] RBAC Dedicated Admins SCC permissions"

func init() {
	alert.RegisterGinkgoAlert(dedicatedAdminSccTestName, "SD-CICD", "Matt Bargenquast", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(dedicatedAdminSccTestName, func() {
	h := helper.New()

	workloadDir := "/assets/workloads/e2e/scc"

	ginkgo.Context("Dedicated Admin permissions", func() {

		ginkgo.It("should include anyuid", func() {
			checkSccPermissions(h, "dedicated-admins-cluster", "anyuid")
		})

		ginkgo.It("should include nonroot", func() {
			checkSccPermissions(h, "dedicated-admins-cluster", "nonroot")
		})

		ginkgo.It("can create pods with SCCs", func() {
			_, err := helper.ApplyYamlInFolder(workloadDir, h.CurrentProject(), h.Kube())
			Expect(err).NotTo(HaveOccurred(), "couldn't apply workload yaml")
		})
	})
})

func checkSccPermissions(h *helper.H, clusterRole string, scc string) {

	// Get the cluster role containing the definition
	cr, err := h.Kube().RbacV1().ClusterRoles().Get(context.TODO(), clusterRole, metav1.GetOptions{})
	Expect(err).ToNot(HaveOccurred(), "failed to get clusterRole %s\n", clusterRole)

	foundRule := false
	for _, rule := range cr.Rules {

		// Find rules relating to SCCs
		isSccRule := false
		for _, resource := range rule.Resources {
			if resource == "securitycontextconstraints" {
				isSccRule = true
			}
		}
		if !isSccRule {
			continue
		}

		// check for 'use' verb
		for _, verb := range rule.Verbs {
			if verb == "use" {
				foundRule = true
				break
			}
		}
	}
	Expect(foundRule).To(BeTrue())
}
