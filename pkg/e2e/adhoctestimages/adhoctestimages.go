package adhoctestimages

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/go-logr/logr"
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e-common/pkg/clients/ocm"
	"github.com/openshift/osde2e/internal/analysisengine"
	"github.com/openshift/osde2e/internal/reporter"
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

			// Defer the Expect calls to ensure they always run and get logged
			defer func() {
				Expect(err).NotTo(HaveOccurred(), "failed to run test suite")
				if results != nil {
					for _, suite := range results.Suites {
						for _, test := range suite.Tests {
							Expect(test.Error).To(BeNil(), fmt.Sprintf("failed test case: %q", test.Name))
						}
					}
				}
			}()

			// Collect failures in single loop
			var allFailures []string

			// Check for execution failure
			if err != nil {
				allFailures = append(allFailures, fmt.Sprintf("execution failure: %v", err))
			}

			// Collect test case failures in single loop
			if results != nil {
				for _, suite := range results.Suites {
					for _, test := range suite.Tests {
						if test.Error != nil {
							allFailures = append(allFailures, fmt.Sprintf("test case failure: %q - %v", test.Name, test.Error))
						}
					}
				}
			}

			if len(allFailures) > 0 && viper.GetBool(config.LogAnalysis.EnableAnalysis) {
				combinedErr := fmt.Errorf("failures in %s: %s", testImage, strings.Join(allFailures, "; "))
				runLogAnalysisForAdHocTestImage(ctx, logger, testImageConfig, combinedErr, exeConfig.OutputDir)
			}
		},
		testImageEntries)
})

// runLogAnalysisForAdHocTestImage performs log analysis powered failure analysis for a specific test image
func runLogAnalysisForAdHocTestImage(ctx context.Context, logger logr.Logger, testImageConfig config.AdHocTestImage, err error, artifactsDir string) {
	logger.Info("Running Log analysis for test image", "image", testImageConfig.Image, "slackChannel", testImageConfig.SlackChannel)

	clusterInfo := &analysisengine.ClusterInfo{
		ID:            viper.GetString(config.Cluster.ID),
		Name:          viper.GetString(config.Cluster.Name),
		Provider:      viper.GetString(config.Provider),
		Region:        viper.GetString(config.CloudProvider.Region),
		CloudProvider: viper.GetString(config.CloudProvider.CloudProviderID),
		Version:       viper.GetString(config.Cluster.Version),
	}

	// Setup notification config - composable approach for multiple reporters
	var notificationConfig *reporter.NotificationConfig
	var reporters []reporter.ReporterConfig

	// Get the global main slack workflow webhook
	slackWebhook := viper.GetString(config.LogAnalysis.SlackWebhook)
	enableSlackNotify := viper.GetBool(config.LogAnalysis.EnableSlackNotify)

	// Add Slack reporter if enabled, webhook exists, and channel is specified
	if enableSlackNotify && slackWebhook != "" && testImageConfig.SlackChannel != "" {
		slackConfig := reporter.SlackReporterConfig(slackWebhook, true)
		slackConfig.Settings["channel"] = testImageConfig.SlackChannel
		reporters = append(reporters, slackConfig)
	}

	// Create notification config if we have any reporters
	if len(reporters) > 0 {
		notificationConfig = &reporter.NotificationConfig{
			Enabled:   true,
			Reporters: reporters,
		}
	}

	engineConfig := &analysisengine.Config{
		ArtifactsDir:       artifactsDir,
		PromptTemplate:     "default",
		APIKey:             viper.GetString(config.LogAnalysis.APIKey),
		FailureContext:     err.Error(),
		ClusterInfo:        clusterInfo,
		NotificationConfig: notificationConfig,
	}

	engine, err := analysisengine.New(ctx, engineConfig)
	if err != nil {
		logger.Error(err, "Unable to create analysis engine for image", "image", testImageConfig.Image)
		return
	}

	result, runErr := engine.Run(ctx)
	if runErr != nil {
		logger.Error(runErr, "Log analysis failed for image", "image", testImageConfig.Image)
		return
	}

	logger.Info("Log analysis completed successfully", "image", testImageConfig.Image, "resultsDir", fmt.Sprintf("%s/%s/", artifactsDir, analysisengine.AnalysisDirName))
	log.Printf("=== Log Analysis Result for %s ===\n%s", testImageConfig.Image, result.Content)
}
