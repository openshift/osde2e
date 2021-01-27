package addons

import (
	"fmt"
	"log"
	"strings"

	"github.com/onsi/ginkgo"
	"github.com/spf13/viper"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
)

var _ = ginkgo.Describe("[Suite: addons] Addon Test Harness", func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	addonTimeoutInSeconds := float64(viper.GetFloat64(config.Tests.PollingTimeout))
	log.Printf("addon timeout is %v", addonTimeoutInSeconds)
	if addonTimeoutInSeconds == 30 {
		// 30s is too short of a time for addons. So override the default with
		// a new default of 60m (3600s)
		addonTimeoutInSeconds = 3600
	}
	ginkgo.It("should run until completion", func() {
		h.SetServiceAccount(viper.GetString(config.Addons.TestUser))
		harnesses := strings.Split(viper.GetString(config.Addons.TestHarnesses), ",")
		failed := h.RunAddonTests("addon-tests", int(addonTimeoutInSeconds), harnesses, []string{})
		if len(failed) > 0 {
			// tests failed, notify
			alert.SendSlackMessage(viper.GetString(config.Addons.SlackChannel), fmt.Sprintf("Addon tests failed: %v", failed))
		}
	}, addonTimeoutInSeconds+30)
})
