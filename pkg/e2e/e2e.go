// Package e2e launches an OSD cluster, performs tests on it, and destroys it.
package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/reporters"
	"github.com/onsi/ginkgo/v2/types"
	"github.com/onsi/gomega"
	"github.com/openshift/osde2e/internal/analysisengine"
	"github.com/openshift/osde2e/internal/reporter"
	clusterutil "github.com/openshift/osde2e/pkg/common/cluster"
	"github.com/openshift/osde2e/pkg/common/clusterproperties"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/phase"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/runner"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/upgrade"
	"github.com/openshift/osde2e/pkg/common/util"
	"github.com/openshift/osde2e/pkg/debug"
	ctrlog "sigs.k8s.io/controller-runtime/pkg/log"
)

// provisioner is used to deploy and manage clusters.
var provider spi.Provider

// runLogAnalysis performs log analysis powered failure analysis if enabled
func runLogAnalysis(ctx context.Context, err error) {
	log.Println("Running Log analysis")

	reportDir := viper.GetString(config.ReportDir)
	if reportDir == "" {
		log.Println("No report directory available for Log analysis")
		return
	}

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

	// Add Slack reporter if enabled
	enableSlackNotify := viper.GetBool(config.Tests.EnableSlackNotify)
	slackWebhook := viper.GetString(config.LogAnalysis.SlackWebhook)
	defaultChannel := viper.GetString(config.LogAnalysis.SlackChannel)
	if enableSlackNotify && slackWebhook != "" && defaultChannel != "" {
		slackConfig := reporter.SlackReporterConfig(slackWebhook, true)
		slackConfig.Settings["channel"] = defaultChannel
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
		ArtifactsDir:       reportDir,
		PromptTemplate:     "default",
		APIKey:             viper.GetString(config.LogAnalysis.APIKey),
		FailureContext:     err.Error(),
		ClusterInfo:        clusterInfo,
		NotificationConfig: notificationConfig,
	}

	engine, err := analysisengine.New(ctx, engineConfig)
	if err != nil {
		log.Printf("Unable to create analysis engine: %v", err)
		return
	}

	result, runErr := engine.Run(ctx)
	if runErr != nil {
		log.Printf("Log analysis failed: %v", runErr)
		return
	}

	log.Printf("Log analysis completed successfully. Results written to %s/%s/", reportDir, analysisengine.AnalysisDirName)
	log.Printf("=== Log Analysis Result ===\n%s", result.Content)
}

// beforeSuite attempts to populate several required cluster fields (either by provisioning a new cluster, or re-using an existing one)
// If there is an issue with provisioning, retrieving, or getting the kubeconfig, this will return `false`.
func beforeSuite() bool {
	ctrlog.SetLogger(ginkgo.GinkgoLogr)
	// Skip provisioning if we already have a kubeconfig
	var err error

	// We can capture this error if TEST_KUBECONFIG is set, but we can't use it to skip provisioning
	if err := config.LoadKubeconfig(); err != nil {
		log.Printf("Not loading kubeconfig: %v", err)
	}

	// populate viper clusterID if shared dir contains one.
	// Important to do this beforeSuite for multi step jobs.
	if err := config.LoadClusterId(); err != nil {
		log.Printf("Not loading cluster id: %v", err)
		return false
	}
	var cluster *spi.Cluster
	if provider, err = providers.ClusterProvider(); err != nil {
		log.Printf("Error getting cluster provider: %s", err.Error())
		return false
	}
	if viper.GetString(config.Kubeconfig.Contents) == "" {
		cluster, err = clusterutil.Provision(provider)
		getLogs()
		if err != nil {
			return false
		}
	} else {
		log.Println("Using provided kubeconfig")
		cluster, err = provider.GetCluster(viper.GetString(config.Cluster.ID))
		if err != nil {
			log.Printf("Failed to get cluster %s: %s", viper.GetString(config.Cluster.ID), err)
			return false
		}
	}
	clusterutil.SetClusterIntoViperConfig(cluster)

	if len(viper.GetString(config.Addons.IDs)) > 0 {
		if viper.GetString(config.Provider) != "mock" {
			err = installAddons()
			if err != nil {
				log.Printf("Cluster failed installing addons: %v", err)
				getLogs()
				return false
			}
		} else {
			log.Println("Skipping addon installation due to mock provider.")
			log.Println("If you are running local addon tests, please ensure the addon components are already installed.")
		}
	}

	return true
}

func getLogs() {
	clusterID := viper.GetString(config.Cluster.ID)
	if provider == nil {
		log.Println("OSD was not configured. Skipping log collection...")
	} else if clusterID == "" {
		log.Println("CLUSTER_ID is not set, likely due to a setup failure. Skipping log collection...")
	} else {
		logs, err := provider.Logs(clusterID)
		if err != nil {
			log.Printf("Error collecting cluster logs: %s", err.Error())
		} else {
			writeLogs(logs)
		}
	}
}

func writeLogs(m map[string][]byte) {
	for k, v := range m {
		name := k + "-log.txt"
		filePath := filepath.Join(viper.GetString(config.ReportDir), name)
		err := os.WriteFile(filePath, v, os.ModePerm)
		if err != nil {
			log.Printf("Error writing log %s: %s", filePath, err.Error())
		}
	}
}

// installAddons installs addons onto the cluster
func installAddons() (err error) {
	clusterID := viper.GetString(config.Cluster.ID)
	params := make(map[string]map[string]string)
	strParams := viper.GetString(config.Addons.Parameters)
	if err := json.Unmarshal([]byte(strParams), &params); err != nil {
		return fmt.Errorf("failed unmarshalling addon parameters %s: %w", strParams, err)
	}
	num, err := provider.InstallAddons(clusterID, strings.Split(viper.GetString(config.Addons.IDs), ","), params)
	if err != nil {
		return fmt.Errorf("could not install addons: %s", err.Error())
	}
	if num > 0 {
		if err = clusterutil.WaitForClusterReadyPostInstall(clusterID, nil); err != nil {
			return fmt.Errorf("failed waiting for cluster ready: %v", err)
		}
	}

	return nil
}

// -- END Ginkgo setup

// RunTests initializes Ginkgo and runs the osde2e test suite.
// Deprecated: This function is maintained for backward compatibility but will be removed.
// Use the new orchestrator-based architecture instead: e2e.NewFactory().NewOrchestrator().Run(ctx)
func RunTests(ctx context.Context) int {
	var err error
	var exitCode int

	testing.Init()

	exitCode, err = runGinkgoTests()
	if err != nil {
		log.Printf("OSDE2E failed: %v", err)
		if viper.GetBool(config.LogAnalysis.EnableAnalysis) {
			runLogAnalysis(ctx, err)
		}
	}

	return exitCode
}

// runGinkgoTests runs the osde2e test suite using Ginkgo.
// nolint:gocyclo
func runGinkgoTests() (int, error) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	viper.Set(config.Cluster.Passing, false)
	suiteConfig, reporterConfig := ginkgo.GinkgoConfiguration()
	suiteConfig.Timeout = time.Hour * time.Duration(viper.GetInt(config.Tests.SuiteTimeout))

	if skip := viper.GetString(config.Tests.GinkgoSkip); skip != "" {
		suiteConfig.SkipStrings = append(suiteConfig.SkipStrings, skip)
	}

	if labels := viper.GetString(config.Tests.GinkgoLabelFilter); labels != "" {
		suiteConfig.LabelFilter = labels
	}

	if testsToRun := viper.GetStringSlice(config.Tests.TestsToRun); len(testsToRun) > 0 {
		suiteConfig.FocusStrings = testsToRun
	}

	if focus := viper.GetString(config.Tests.GinkgoFocus); focus != "" {
		suiteConfig.FocusStrings = append(suiteConfig.FocusStrings, focus)
	}
	suiteConfig.DryRun = viper.GetBool(config.DryRun)

	if suiteConfig.DryRun {
		// Draw attention to DRYRUN as it can exist in ENV.
		log.Println(string("\x1b[33m"), "WARNING! This is a DRY RUN. Review this state if outcome is unexpected.", string("\033[0m"))
	}

	logLevel := viper.GetString(config.Tests.GinkgoLogLevel)
	switch logLevel {
	case "v":
		reporterConfig.Verbose = true
	case "vv":
		reporterConfig.VeryVerbose = true
	case "succinct":
		fallthrough
	default:
		reporterConfig.Succinct = true
	}

	// Suppress color output
	// https://onsi.github.io/ginkgo/#other-settings
	reporterConfig.NoColor = true

	if viper.GetString(config.Suffix) == "" {
		viper.Set(config.Suffix, util.RandomStr(5))
	}

	runInstallTests := true
	upgradeCluster := false
	if viper.GetString(config.Upgrade.Image) != "" || viper.GetString(config.Upgrade.ReleaseName) != "" {
		upgradeCluster = true
		if runInstallTests = viper.GetBool(config.Upgrade.RunPreUpgradeTests); !runInstallTests {
			if !suiteConfig.DryRun {
				if !beforeSuite() {
					return config.Failure, fmt.Errorf("error occurred during beforeSuite function")
				}
			}
		}
	}

	testsPassed := true
	if runInstallTests {
		log.Println("Running e2e tests...")
		testsPassed = runTestsInPhase(phase.InstallPhase, "OSD e2e suite", suiteConfig, reporterConfig)
		getLogs()
		viper.Set(config.Cluster.Passing, testsPassed)
	}

	upgradeTestsPassed := true

	// upgrade cluster if requested
	if upgradeCluster {
		if len(viper.GetString(config.Kubeconfig.Contents)) > 0 {
			// setup helper
			h, err := helper.NewOutsideGinkgo()
			if h == nil || err != nil {
				return config.Failure, fmt.Errorf("unable to generate helper outside ginkgo: %v", err)
			}

			// run the upgrade
			if err = upgrade.RunUpgrade(h); err != nil {
				return config.Failure, fmt.Errorf("error performing upgrade: %v", err)
			}

			if viper.GetBool(config.Upgrade.RunPostUpgradeTests) {
				log.Println("Running e2e tests POST-UPGRADE...")
				viper.Set(config.Cluster.Passing, false)
				upgradeTestsPassed = runTestsInPhase(
					phase.UpgradePhase,
					"OSD e2e suite post-upgrade",
					suiteConfig,
					reporterConfig,
				)
				viper.Set(config.Cluster.Passing, upgradeTestsPassed)
			}
		} else {
			log.Println("Unable to perform cluster upgrade, no kubeconfig found.")
		}
	}

	// Cleanup
	if !suiteConfig.DryRun {
		getLogs()

		h, err := helper.NewOutsideGinkgo()
		if h == nil {
			log.Printf("Failed to generate helper object to perform cleanup operations, deleting cluster: %t", !viper.GetBool(config.Cluster.SkipDestroyCluster))
			// Ignoring the error to return actual error which caused runtime to abort
			_ = deleteCluster(provider)
			return config.Failure, fmt.Errorf("unable to generate helper object for cleanup: %v", err)
		}

		cleanupAfterE2E(context.TODO(), h)

		if err = deleteCluster(provider); err != nil {
			return config.Failure, err
		}
	}

	if !testsPassed || !upgradeTestsPassed {
		viper.Set(config.Cluster.Passing, false)
		return config.Failure, fmt.Errorf("tests failed, please inspect logs for more details")
	}

	return config.Success, nil
}

// deleteCluster destroys the cluster based on defined settings
func deleteCluster(provider spi.Provider) error {
	clusterID := viper.GetString(config.Cluster.ID)

	if clusterID == "" {
		log.Printf("Cluster ID is empty, unable to destroy cluster")
		return nil
	}

	if !viper.GetBool(config.Cluster.SkipDestroyCluster) {
		log.Printf("Destroying cluster '%s'...", clusterID)

		if err := provider.DeleteCluster(clusterID); err != nil {
			return fmt.Errorf("error deleting cluster: %s", err.Error())
		}
	} else {
		if provider != nil {
			log.Printf("For debugging, please look for cluster ID %s in environment %s", clusterID, provider.Environment())
		}
	}

	return nil
}

func cleanupAfterE2E(ctx context.Context, h *helper.H) (errors []error) {
	clusterStatus := clusterproperties.StatusCompletedFailing
	defer ginkgo.GinkgoRecover()

	if !viper.GetBool(config.SkipMustGather) {
		log.Print("Running Must Gather...")
		mustGatherTimeoutInSeconds := 1800
		h.SetServiceAccount(ctx, "system:serviceaccount:%s:cluster-admin")
		r := h.Runner(fmt.Sprintf("oc adm must-gather --dest-dir=%v", runner.DefaultRunner.OutputDir))
		r.Name = "must-gather"
		r.Tarball = true
		stopCh := make(chan struct{})
		err := r.Run(mustGatherTimeoutInSeconds, stopCh)

		if err != nil {
			log.Printf("Error running must-gather: %s", err.Error())
			clusterStatus = clusterproperties.StatusCompletedError
		} else {
			gatherResults, err := r.RetrieveResults()
			if err != nil {
				log.Printf("Error retrieving must-gather results: %s", err.Error())
				clusterStatus = clusterproperties.StatusCompletedError
			} else {
				h.WriteResults(gatherResults)
			}
		}

		log.Print("Gathering Project States...")
		h.InspectState(ctx)

		log.Print("Gathering OLM State...")
		if err = h.InspectOLM(ctx); err != nil {
			errors = append(errors, err)
		}
	} else {
		log.Print("Skipping must-gather as requested")
	}

	clusterID := viper.GetString(config.Cluster.ID)
	if len(clusterID) > 0 {

		// Get state from Provisioner
		log.Printf("Gathering cluster state from %s", provider.Type())

		cluster, err := provider.GetCluster(clusterID)
		if err != nil {
			log.Printf("error getting Cluster state: %s", err.Error())
			clusterStatus = clusterproperties.StatusCompletedError
		} else {
			defer func() {
				// set the completed property right before this function returns, which should be after
				// all cleanup is finished.
				if viper.GetBool(config.Cluster.Passing) {
					clusterStatus = clusterproperties.StatusCompletedPassing
				}

				err = provider.AddProperty(cluster, clusterproperties.Status, clusterStatus)
				err = provider.AddProperty(cluster, clusterproperties.JobID, "")
				err = provider.AddProperty(cluster, clusterproperties.JobName, "")
				err = provider.AddProperty(cluster, clusterproperties.Availability, clusterproperties.Used)
				if err != nil {
					log.Printf("Failed setting completed status: %v", err)
				}
			}()
			log.Printf("Cluster addons: %v", cluster.Addons())
			log.Printf("Cluster cloud provider: %v", cluster.CloudProvider())
			log.Printf("Cluster expiration: %v", cluster.ExpirationTimestamp())
			log.Printf("Cluster flavor: %s", cluster.Flavour())
			log.Printf("Cluster state: %v", cluster.State())
		}

	} else {
		log.Print("No cluster ID set. Skipping OCM Queries.")
	}

	// We need to clean up our helper tests manually.
	h.Cleanup(ctx)

	// If this is a nightly test, we don't want to expire this immediately
	if viper.GetString(config.Cluster.InstallSpecificNightly) != "" || viper.GetString(config.Cluster.ReleaseImageLatest) != "" {
		if viper.GetString(config.Cluster.ID) != "" {
			if err := provider.Expire(viper.GetString(config.Cluster.ID), 30*time.Minute); err != nil {
				errors = append(errors, err)
			}
		}
	}

	if provider != nil && viper.GetString(config.Cluster.ID) != "" && viper.GetBool(config.Cluster.SkipDestroyCluster) {
		// Current default expiration is 6 hours.
		// If this cluster has addons, we don't want to extend the expiration

		if !viper.GetBool(config.Cluster.ClaimedFromReserve) && clusterStatus != clusterproperties.StatusCompletedError && viper.GetString(config.Addons.IDs) == "" {
			cluster, err := provider.GetCluster(viper.GetString(config.Cluster.ID))
			if err != nil {
				log.Printf("Error getting cluster from provider: %s", err.Error())
			}
			if !cluster.ExpirationTimestamp().Add(6 * time.Hour).After(cluster.CreationTimestamp().Add(24 * time.Hour)) {
				if err := provider.ExtendExpiry(viper.GetString(config.Cluster.ID), 6, 0, 0); err != nil {
					log.Printf("Error extending cluster expiration: %s", err.Error())
				}
			}
		}
	}
	return errors
}

// nolint:gocyclo
func runTestsInPhase(
	phase string,
	description string,
	suiteConfig types.SuiteConfig,
	reporterConfig types.ReporterConfig,
) bool {
	viper.Set(config.Phase, phase)
	reportDir := viper.GetString(config.ReportDir)
	phaseDirectory := filepath.Join(reportDir, phase)
	if _, err := os.Stat(phaseDirectory); os.IsNotExist(err) {
		if err := os.Mkdir(phaseDirectory, os.FileMode(0o755)); err != nil {
			log.Printf("error while creating phase directory %s", phaseDirectory)
			return false
		}
	}
	suffix := viper.GetString(config.Suffix)
	ginkgoPassed := false

	if !suiteConfig.DryRun {
		if !beforeSuite() {
			return false
		}
	}

	// Generate JUnit report once all tests have finished with customized settings
	_ = ginkgo.ReportAfterSuite("OSDE2E", func(report ginkgo.Report) {
		err := reporters.GenerateJUnitReportWithConfig(
			report,
			filepath.Join(phaseDirectory, fmt.Sprintf("junit_%v.xml", suffix)),
			reporters.JunitReportConfig{OmitSpecLabels: true, OmitLeafNodeType: true},
		)
		if err != nil {
			log.Printf("error creating junit report file %s", err.Error())
		}
	})

	// https://github.com/konflux-ci/architecture/blob/cd41772b27bb89cd061e85cdaa7488afc4e29a2e/ADR/0030-tekton-results-naming-convention.md
	if konfluxTestOutputFile := viper.GetString(config.KonfluxTestOutputFile); konfluxTestOutputFile != "" {
		ginkgo.ReportAfterSuite("OSDE2E konflux results", func(report ginkgo.Report) {
			konfluxResults := map[string]any{
				"result":    "FAILURE",
				"timestamp": report.EndTime.Format(time.RFC3339),
				// TODO: do something with warnings
				"warnings": 0,
			}

			if report.SuiteSucceeded {
				konfluxResults["result"] = "SUCCESS"
			}

			var successes, failures int
			for _, report := range report.SpecReports {
				switch report.State {
				case types.SpecStatePassed:
					successes++
				case types.SpecStateFailed:
					failures++
				}
			}
			konfluxResults["successes"] = successes
			konfluxResults["failures"] = failures

			bits, err := json.Marshal(konfluxResults)
			if err != nil {
				log.Printf("unable to marshal konflux results: %s", err)
			}
			if err = os.WriteFile(konfluxTestOutputFile, bits, os.ModePerm); err != nil {
				log.Printf("failed to write konflux results to %s: %s", "", err)
			}
		})
	}

	// We need this anonymous function to make sure GinkgoRecover runs where we want it to
	// and will still execute the rest of the function regardless whether the tests pass or fail.
	func() {
		defer ginkgo.GinkgoRecover()

		ginkgoPassed = ginkgo.RunSpecs(ginkgo.GinkgoT(), description, suiteConfig, reporterConfig)
	}()

	clusterID := viper.GetString(config.Cluster.ID)

	clusterState := spi.ClusterStateUnknown

	if clusterID != "" {
		cluster, err := provider.GetCluster(clusterID)
		if err != nil {
			log.Printf("error getting cluster state after a test run: %v", err)
			return false
		}
		clusterState = cluster.State()
	}
	if !suiteConfig.DryRun && clusterState == spi.ClusterStateReady && viper.GetString(config.JobName) != "" && viper.GetString(config.JobType) == "periodic" {
		h, err := helper.NewOutsideGinkgo()
		if h == nil {
			log.Println("Unable to generate helper outside of ginkgo: %w", err)
			return ginkgoPassed
		}
		dependencies, err := debug.GenerateDependencies(h.Kube())
		if err != nil {
			log.Printf("Error generating dependencies: %s", err.Error())
		} else {
			if err = os.WriteFile(filepath.Join(phaseDirectory, "dependencies.txt"), []byte(dependencies), 0o644); err != nil {
				log.Printf("Error writing dependencies.txt: %s", err.Error())
			}

			err := debug.GenerateDiff(phase, dependencies)
			if err != nil {
				log.Printf("Error generating diff: %s", err.Error())
			}

		}
	}
	return ginkgoPassed
}
