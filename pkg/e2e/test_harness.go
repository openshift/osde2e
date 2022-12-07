package e2e

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

//Generic harness suite: Essentially replicates addon harness suite, with generalized harness suite
var _ = ginkgo.Describe("[Suite: harnesses] Test Harness", func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	testTimeoutInSeconds := float64(viper.GetFloat64(config.Harness.PollingTimeout))
	util.GinkgoIt("should run until completion", func(ctx context.Context) {
		log.Printf(" timeout is %v", testTimeoutInSeconds)
		h.SetServiceAccount(ctx, viper.GetString(config.Harness.TestUser))
		harnesses := strings.Split(viper.GetString(config.Harness.Images), ",")
		failed := h.RunTestHarness(ctx, "test-harness", int(testTimeoutInSeconds), harnesses, []string{})
		if len(failed) > 0 {
			message := fmt.Sprintf("Tests failed: %v", failed)
			if url, ok := prow.JobURL(); ok {
				message += "\n" + url
			}
			if err := alert.SendSlackMessage(viper.GetString(config.Harness.SlackChannel), message); err != nil {
				log.Printf("Failed sending slack alert for test failure: %v", err)
			}
		}
	}, testTimeoutInSeconds+30)
})
