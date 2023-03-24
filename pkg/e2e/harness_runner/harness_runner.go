package harness_runner

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/openshift/osde2e/pkg/common/alert"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/prow"
)

var h *helper.H

var _ = ginkgo.Describe("[Suite: harnessess] Test Harness", func() {
	ginkgo.BeforeAll(func() {
		h = helper.New()
	})

	TimeoutInSeconds := viper.GetFloat64(config.Tests.PollingTimeout)
	ginkgo.It("should run until completion", func(ctx context.Context) {
		ginkgo.GinkgoWriter.Write([]byte("Test harness timeout is " + fmt.Sprintf("%v", TimeoutInSeconds)))
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
