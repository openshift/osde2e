package operators

import (
	"github.com/onsi/ginkgo"
	"github.com/openshift/osde2e/pkg/helper"
)

var _ = ginkgo.Describe("[OSD] Pagerduty Operator", func() {
	h := helper.New()
	var secrets = []string{
		"pd-secret",
	}
	var namespace string = "openshift-monitoring"
	// Check if pd-secret exists under openshift-monitoring ns
	checkSecrets(h, namespace, secrets)
})
