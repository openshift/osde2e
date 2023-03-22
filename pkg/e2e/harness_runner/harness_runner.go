package harness_runner

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/onsi/ginkgo/v2"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/util"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/prow"
)

var _ = ginkgo.Describe("[Suite: Tests] Test Harness", func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	TimeoutInSeconds := float64(viper.GetFloat64(config.Tests.PollingTimeout))
	util.GinkgoIt("should run until completion", func(ctx context.Context) {
		log.Printf("Test harness timeout is %v", TimeoutInSeconds)
		h.SetServiceAccount(ctx, viper.GetString(config.Tests.TestUser))
		harnesses := strings.Split(viper.GetString(config.Tests.TestHarnesses), ",")
		failed := h.RunTests(ctx, "test-harness", int(TimeoutInSeconds), harnesses, []string{})
		if len(failed) > 0 {
			message := fmt.Sprintf("Tests failed: %v", failed)
			if url, ok := prow.JobURL(); ok {
				message += "\n" + url
			}
			if err := alert.SendSlackMessage(viper.GetString(config.Tests.SlackChannel), message); err != nil {
				log.Printf("Failed sending slack alert for test failure: %v", err)
			}
		}
	}, TimeoutInSeconds+30)
})
