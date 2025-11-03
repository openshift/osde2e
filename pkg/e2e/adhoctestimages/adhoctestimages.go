package adhoctestimages

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e-common/pkg/clients/ocm"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/executor"
	"github.com/openshift/osde2e/pkg/common/label"
	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
	"k8s.io/client-go/tools/clientcmd"
)

var _ = ginkgo.Describe("Ad Hoc Test Images", ginkgo.Ordered, ginkgo.ContinueOnFailure, label.AdHocTestImages, func() {
	var (
		logger           = ginkgo.GinkgoLogr
		testImageEntries = []ginkgo.TableEntry{}
		testImages       []config.AdHocTestImage
		exeConfig        = &executor.Config{
			CloudProviderID:     viper.GetString(config.CloudProvider.CloudProviderID),
			CloudProviderRegion: viper.GetString(config.CloudProvider.Region),
			ClusterID:           viper.GetString(config.Cluster.ID),
			Environment:         ocm.Environment(viper.GetString(ocmprovider.Env)),
			PassthruSecrets:     viper.GetStringMapString(config.NonOSDe2eSecrets),
			SkipCleanup:         viper.GetBool(config.Cluster.SkipDestroyCluster),
			Timeout:             viper.GetDuration(config.Tests.AdHocTestContainerTimeout),
		}
		exe *executor.Executor
	)

	// Get test images using the new structured format
	var err error
	testImages, err = config.GetAdHocTestImages()
	if err != nil {
		ginkgo.Fail(fmt.Sprintf("Failed to get AdHocTestImages configuration: %v", err))
	}

	for _, testImage := range testImages {
		testImageEntries = append(testImageEntries, ginkgo.Entry(testImage.Image+" should pass", testImage))
	}

	ginkgo.BeforeAll(func(ctx context.Context) {
		var err error
		exeConfig.RestConfig, err = clientcmd.RESTConfigFromKubeConfig([]byte(viper.GetString(config.Kubeconfig.Contents)))
		Expect(err).NotTo(HaveOccurred())

		exe, err = executor.New(logger, exeConfig)
		Expect(err).NotTo(HaveOccurred())

		// Log test images with their slack channels
		imageNames := make([]string, len(testImages))
		for i, img := range testImages {
			imageNames[i] = img.Image
		}
		logger.Info("executing test suites", "suites", imageNames)
	})

	ginkgo.DescribeTable("execution",
		func(ctx context.Context, testImageConfig config.AdHocTestImage) {
			testImage := testImageConfig.Image
			baseImageName := strings.Split(testImage[strings.LastIndex(testImage, "/")+1:], ":")[0]
			exeConfig.OutputDir = filepath.Join(viper.GetString(config.ReportDir), viper.GetString(config.Phase), baseImageName)

			logger.Info("running test suite", "suite", testImage, "slackChannel", testImageConfig.SlackChannel, "timeout", exeConfig.Timeout)
			results, err := exe.Execute(ctx, testImage)
			Expect(err).NotTo(HaveOccurred(), "failed to run test suite")

			for _, suite := range results.Suites {
				for _, test := range suite.Tests {
					Expect(test.Error).To(BeNil(), fmt.Sprintf("failed test case: %q", test.Name))
				}
			}
		},
		testImageEntries)
})
