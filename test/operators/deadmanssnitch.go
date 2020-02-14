package operators

import (
	"github.com/onsi/ginkgo"
	"github.com/openshift/osde2e/pkg/helper"
)

var _ = ginkgo.Describe("[Suite: operators] [OSD] DeadMansSnitch Operator", func() {

	var namespace = "openshift-monitoring"
	var secrets = []string{
		"dms-secret",
	}

	h := helper.New()
	checkSecrets(h, namespace, secrets)
})
