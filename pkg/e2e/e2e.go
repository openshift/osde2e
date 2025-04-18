// Package e2e launches an OSD cluster, performs tests on it, and destroys it.
package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	clusterutil "github.com/openshift/osde2e/pkg/common/cluster"
	"github.com/openshift/osde2e/pkg/common/clusterproperties"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/events"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/phase"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/runner"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/upgrade"
	"github.com/openshift/osde2e/pkg/common/util"
	"github.com/openshift/osde2e/pkg/common/versions"
	"github.com/openshift/osde2e/pkg/debug"
	"github.com/openshift/osde2e/pkg/e2e/routemonitors"
	vegeta "github.com/tsenart/vegeta/lib"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrlog "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// hiveLog is the name of the hive log file.
	hiveLog string = "hive-log.txt"

	// buildLog is the name of the build log file.
	buildLog string = "test_output.log"

	Success = 0
	Failure = 1
	Aborted = 130
)

// provisioner is used to deploy and manage clusters.
var provider spi.Provider

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
		status, err := configureVersion()
		if status != Success {
			log.Printf("Failed configure cluster version: %v", err)
			return false
		}
		cluster, err = clusterutil.ProvisionCluster(nil)

		events.HandleErrorWithEvents(err, events.InstallSuccessful, events.InstallFailed)
		if err != nil {
			log.Printf("Failed to set up or retrieve cluster: %v", err)
			getLogs()
			return false
		}

		viper.Set(config.Cluster.ID, cluster.ID())
		log.Printf("CLUSTER_ID set to %s from OCM.", viper.GetString(config.Cluster.ID))
		_, err = clusterutil.WaitForOCMProvisioning(provider, viper.GetString(config.Cluster.ID), nil, false)
		if err != nil {
			log.Printf("Cluster never became ready: %v", err)
			getLogs()
			return false
		}
		log.Printf("Cluster status is ready")

		if viper.GetString(config.SharedDir) != "" {
			if err = os.WriteFile(fmt.Sprintf("%s/cluster-id", viper.GetString(config.SharedDir)), []byte(cluster.ID()), 0o644); err != nil {
				log.Printf("Error writing cluster ID to shared directory: %v", err)
			} else {
				log.Printf("Wrote cluster ID to shared dir: %v", cluster.ID())
			}
		} else {
			log.Printf("No shared directory provided, skip writing cluster ID")
		}

		if (!viper.GetBool(config.Addons.SkipAddonList) || viper.GetString(config.Provider) != "mock") && len(cluster.Addons()) > 0 {
			log.Printf("Found addons: %s", strings.Join(cluster.Addons(), ","))
		}

		if err = provider.AddProperty(cluster, "UpgradeVersion", viper.GetString(config.Upgrade.ReleaseName)); err != nil {
			log.Printf("Error while adding upgrade version property to cluster via OCM: %v", err)
		}

		if !viper.GetBool(config.Tests.SkipClusterHealthChecks) {
			if viper.GetBool(config.Cluster.Reused) {
				// We should manually run all our health checks if the cluster is waking up
				err = clusterutil.WaitForClusterReadyPostWake(cluster.ID(), nil)
			} else {
				// This is a new cluster and we should check the OSD Ready job
				err = clusterutil.WaitForClusterReadyPostInstall(cluster.ID(), nil)
			}
			if err != nil {
				log.Println("*******************")
				log.Printf("Cluster failed health check: %v", err)
				log.Println("*******************")
				getLogs()
			} else {
				log.Println("Cluster is healthy and ready for testing")
			}
		} else {
			log.Println("Skipping health checks as requested")
		}

		var kubeconfigBytes []byte
		clusterConfigerr := wait.PollUntilContextTimeout(context.Background(), 2*time.Second, 5*time.Minute, false, func(ctx context.Context) (bool, error) {
			kubeconfigBytes, err = provider.ClusterKubeconfig(viper.GetString(config.Cluster.ID))
			if err != nil {
				log.Printf("Failed to retrieve kubeconfig: %v\nWaiting two seconds before retrying", err)
				return false, nil
			} else {
				log.Printf("Successfully retrieved kubeconfig from OCM.")
				viper.Set(config.Kubeconfig.Contents, string(kubeconfigBytes))
				return true, nil
			}
		})

		if clusterConfigerr != nil {
			events.HandleErrorWithEvents(err, events.InstallKubeconfigRetrievalSuccess, events.InstallKubeconfigRetrievalFailure)
			log.Printf("Failed retrieving kubeconfig: %v", clusterConfigerr)
			getLogs()
			return false
		}

		if viper.GetString(config.SharedDir) != "" {
			if err = os.WriteFile(fmt.Sprintf("%s/kubeconfig", viper.GetString(config.SharedDir)), kubeconfigBytes, 0o644); err != nil {
				log.Printf("Error writing cluster kubeconfig to shared directory: %v", err)
			} else {
				log.Printf("Passed kubeconfig to prow steps.")
			}
		}

		getLogs()

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
			events.HandleErrorWithEvents(err, events.InstallAddonsSuccessful, events.InstallAddonsFailed)
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

func configureVersion() (int, error) {
	// configure cluster and upgrade versions
	versionSelector := versions.VersionSelector{Provider: provider}
	if err := versionSelector.SelectClusterVersions(); err != nil {
		// If we can't find a version to use, exit with an error code.
		return Failure, err
	}

	switch {
	case !viper.GetBool(config.Cluster.EnoughVersionsForOldestOrMiddleTest):
		return Aborted, fmt.Errorf("there were not enough available cluster image sets to choose and oldest or middle cluster image set to test against -- skipping tests")
	case !viper.GetBool(config.Cluster.PreviousVersionFromDefaultFound):
		return Aborted, fmt.Errorf("no previous version from default found with the given arguments")
	case viper.GetBool(config.Upgrade.UpgradeVersionEqualToInstallVersion):
		return Aborted, fmt.Errorf("install version and upgrade version are the same -- skipping tests")
	case viper.GetString(config.Upgrade.ReleaseName) == util.NoVersionFound:
		return Aborted, fmt.Errorf("no valid upgrade versions were found. Skipping tests")
	case viper.GetString(config.Cluster.Version) == "":
		returnState := Aborted
		if viper.GetBool(config.Cluster.LatestYReleaseAfterProdDefault) || viper.GetBool(config.Cluster.LatestZReleaseAfterProdDefault) {
			log.Println("At the latest available version with no newer targets. Exiting...")
			returnState = Success
		}
		return returnState, fmt.Errorf("no valid install version found")
	}
	return Success, nil
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
func RunTests() int {
	var err error
	var exitCode int

	testing.Init()

	exitCode, err = runGinkgoTests()
	if err != nil {
		log.Printf("OSDE2E failed: %v", err)
	}

	return exitCode
}

// runGinkgoTests runs the osde2e test suite using Ginkgo.
// nolint:gocyclo
func runGinkgoTests() (int, error) {
	var err error

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
		// Flag to delete sice these Print statements are duplicated, all we really are doing is setting an array to be passed to the Ginkgo suite.
		log.Printf("%v", testsToRun)
		suiteConfig.FocusStrings = testsToRun
		log.Printf("%v", suiteConfig.FocusStrings)
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

	reportDir := viper.GetString(config.ReportDir)
	sharedDir := viper.GetString(config.SharedDir)
	runtimeDir := fmt.Sprintf("%s/osde2e-%s", os.TempDir(), util.RandomStr(10))

	if reportDir == "" {
		reportDir = runtimeDir
		viper.Set(config.ReportDir, reportDir)
	}

	log.Printf("Writing files to report directory: %s", reportDir)
	if err = os.MkdirAll(reportDir, os.ModePerm); err != nil {
		log.Printf("Could not create report directory: %s, %v", reportDir, err)
	}

	if sharedDir != "" {
		log.Printf("Writing shared files to directory: %s", sharedDir)
		if err = os.MkdirAll(sharedDir, os.ModePerm); err != nil {
			log.Printf("Could not create shared directory: %s, %v", sharedDir, err)
		}
	}

	// Redirect stdout to where we want it to go
	buildLogPath := filepath.Join(reportDir, buildLog)
	buildLogWriter, err := os.Create(buildLogPath)
	if err != nil {
		return Failure, fmt.Errorf("unable to create build log in report directory: %v", err)
	}

	mw := io.MultiWriter(os.Stdout, buildLogWriter)
	log.SetOutput(mw)

	log.Printf("Outputting log to build log at %s", buildLogPath)

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
					return Failure, fmt.Errorf("error occurred during beforeSuite function")
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
	if viper.GetBool(config.Cluster.ProvisionOnly) {
		log.Println("Provision only execution finished, exiting.")
		return Success, nil
	}
	upgradeTestsPassed := true

	// upgrade cluster if requested
	if upgradeCluster {
		if len(viper.GetString(config.Kubeconfig.Contents)) > 0 {
			// setup helper
			h, err := helper.NewOutsideGinkgo()
			if h == nil || err != nil {
				return Failure, fmt.Errorf("unable to generate helper outside ginkgo: %v", err)
			}

			// create route monitors for the upgrade
			var routeMonitorChan chan struct{}
			closeMonitorChan := make(chan struct{})
			if viper.GetBool(config.Upgrade.MonitorRoutesDuringUpgrade) && !suiteConfig.DryRun {
				routeMonitorChan = setupRouteMonitors(context.TODO(), h, closeMonitorChan)
				log.Println("Route Monitors created.")
			}

			// run the upgrade
			if err = upgrade.RunUpgrade(h); err != nil {
				events.RecordEvent(events.UpgradeFailed)
				return Failure, fmt.Errorf("error performing upgrade: %v", err)
			}
			events.RecordEvent(events.UpgradeSuccessful)

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

			// close route monitors
			if viper.GetBool(config.Upgrade.MonitorRoutesDuringUpgrade) && !suiteConfig.DryRun {
				close(routeMonitorChan)
				<-closeMonitorChan
				log.Println("Route monitors reconciled")
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
			return Failure, fmt.Errorf("unable to generate helper object for cleanup: %v", err)
		}

		cleanupAfterE2E(context.TODO(), h)

		if err = deleteCluster(provider); err != nil {
			return Failure, err
		}
	}

	if !testsPassed || !upgradeTestsPassed {
		viper.Set(config.Cluster.Passing, false)
		return Failure, fmt.Errorf("tests failed, please inspect logs for more details")
	}

	return Success, nil
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

// ManyGroupedFailureName is the incident title assigned to incidents reperesenting a large
// cluster of test failures.
const ManyGroupedFailureName = "A lot of tests failed together"

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

	// We need a provider to hibernate
	// We need a cluster to hibernate
	// We need to check that the test run wants to hibernate after this run
	if provider != nil && viper.GetString(config.Cluster.ID) != "" && viper.GetBool(config.Cluster.SkipDestroyCluster) {
		// Current default expiration is 6 hours.
		// If this cluster has addons, we don't want to extend the expiration

		if !viper.GetBool(config.Cluster.Reused) && clusterStatus != clusterproperties.StatusCompletedError && viper.GetString(config.Addons.IDs) == "" {
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
		if viper.GetBool(config.Cluster.ProvisionOnly) {
			return true
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

// checkBeforeMetricsGeneration runs a variety of checks before generating metrics.
func checkBeforeMetricsGeneration() error {
	// Check for hive-log.txt
	if _, err := os.Stat(filepath.Join(viper.GetString(config.ReportDir), hiveLog)); os.IsNotExist(err) {
		events.RecordEvent(events.NoHiveLogs)
	}

	return nil
}

// setupRouteMonitors initializes performance+availability monitoring of cluster routes,
// returning a channel which can be used to terminate the monitoring.
func setupRouteMonitors(ctx context.Context, h *helper.H, closeChannel chan struct{}) chan struct{} {
	routeMonitorChan := make(chan struct{})
	go func() {
		// Set up the route monitors
		routeMonitors, err := routemonitors.Create(ctx, h)
		if err != nil {
			log.Printf("Error creating route monitors: %v\n", err)
			close(closeChannel)
			return
		}

		// Set the route monitors to become active
		routeMonitors.Start()

		// Set up ongoing monitoring of metric gathering from the monitors
		go func() {
			// Create an aggregate channel of all individual metric channels
			agg := make(chan *vegeta.Result)
			for _, ch := range routeMonitors.Monitors {
				go func(c <-chan *vegeta.Result) {
					for msg := range c {
						agg <- msg
					}
				}(ch)
			}
			for {
				select {
				// A metric is waiting for storage
				case msg := <-agg:
					routeMonitors.Metrics[msg.Attack].Add(msg)
					_ = routeMonitors.Plots[msg.Attack].Add(msg)
				}
			}
		}()

		// Close down route monitoring when signalled to
		for {
			select {
			case <-routeMonitorChan:
				log.Println("Closing route monitors...")
				routeMonitors.End()
				_ = routeMonitors.SaveReports(viper.GetString(config.ReportDir))
				_ = routeMonitors.SavePlots(viper.GetString(config.ReportDir))
				_ = routeMonitors.ExtractData(viper.GetString(config.ReportDir))
				close(closeChannel)
				return
			}
		}
	}()
	return routeMonitorChan
}
