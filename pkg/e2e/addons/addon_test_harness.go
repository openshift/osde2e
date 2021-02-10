package addons

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/onsi/ginkgo"
	"github.com/spf13/viper"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
)

// jobURL infers the URL of this job using environment variables
// provided by Prow. It is not foolproof, and the URLs generated
// are only valid for "JOB_TYPE=periodic" jobs.
func jobURL() (url string, ok bool) {
	if os.Getenv("JOB_TYPE") != "periodic" {
		return
	}
	var jobID, jobName string
	if jobID, ok = os.LookupEnv("BUILD_ID"); !ok {
		return
	}
	if jobName, ok = os.LookupEnv("JOB_NAME"); !ok {
		return
	}
	return fmt.Sprintf("https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/%s/%s", jobName, jobID), true
}

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
			message := fmt.Sprintf("Addon tests failed: %v", failed)
			if url, ok := jobURL(); ok {
				message += "\n" + url
			}
			if err := alert.SendSlackMessage(viper.GetString(config.Addons.SlackChannel), message); err != nil {
				log.Printf("Failed sending slack alert for addon failure: %v", err)
			}
		}
	}, addonTimeoutInSeconds+30)
})
