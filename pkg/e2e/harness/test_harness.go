package harness

import (
	"fmt"
	"log"
	"strings"

	"github.com/onsi/ginkgo"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/prow"
)

var _ = ginkgo.Describe("[Suite: harness] Test Harness", func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	harnessTimeoutInSeconds := float64(viper.GetFloat64(config.Tests.HarnessTimeout))
	log.Printf("harness timeout is %v", harnessTimeoutInSeconds)
	ginkgo.It("should run until completion", func() {
		harnesses := strings.Split(viper.GetString(config.Tests.TestHarnesses), ",")
		failed := h.RunTestHarness("test-harness", int(harnessTimeoutInSeconds), harnesses, []string{})
		if len(failed) > 0 {
			message := fmt.Sprintf("Test harness failed: %v", failed)
			if url, ok := prow.JobURL(); ok {
				message += "\n" + url
			}
			if viper.GetString(config.Addons.SlackChannel) != "" {
				if err := alert.SendSlackMessage(viper.GetString(config.Addons.SlackChannel), message); err != nil {
					log.Printf("Failed sending slack alert for addon failure: %v", err)
				}
			}
		}
	}, harnessTimeoutInSeconds+30)
})
