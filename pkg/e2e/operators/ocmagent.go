package operators

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
)

var (
	ocmAgentTestPrefix = "[Suite: informing] [OSD] OCM Agent Operator"
	ocmAgentBasicTest  = ocmAgentTestPrefix + " Basic Test"
)

func init() {
	alert.RegisterGinkgoAlert(ocmAgentBasicTest, "SD_SREP", "@sre-platform-team-v1alpha1", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(ocmAgentBasicTest, func() {
	var (
		operatorNamespace = "openshift-ocm-agent-operator"
		operatorName      = "ocm-agent-operator"
		clusterRoles      = []string{
			"ocm-agent-operator",
		}
		clusterRoleBindings = []string{
			"ocm-agent-operator",
		}
		// servicePort = 8081
	)
	h := helper.New()
	checkClusterServiceVersion(h, operatorNamespace, operatorName)
	checkDeployment(h, operatorNamespace, operatorName, 1)
	checkClusterRoles(h, clusterRoles, true)
	checkClusterRoleBindings(h, clusterRoleBindings, true)
	// checkService(h, operatorNamespace, operatorName, servicePort)
	// checkUpgrade(helper.New(), operatorNamespace, operatorName, operatorName, "ocm-agent-operator-registry")
})
