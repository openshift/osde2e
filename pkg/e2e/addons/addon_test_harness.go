package addons

import (
	"github.com/onsi/ginkgo"
	"github.com/spf13/viper"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
)

var _ = ginkgo.Describe("[Suite: addons] Addon Test Harness", func() {
	defer ginkgo.GinkgoRecover()
	h := helper.New()

	addonTimeoutInSeconds := 3600
	ginkgo.It("should run until completion", func() {
		h.SetServiceAccount(viper.GetString(config.Addons.TestUser))
		h.RunAddonTests("addon-tests", []string{})
	}, float64(addonTimeoutInSeconds+30))
})
