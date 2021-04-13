// Package e2e launches an OSD cluster, performs tests on it, and destroys it.
package e2e

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	junit "github.com/joshdk/go-junit"
	vegeta "github.com/tsenart/vegeta/lib"

	pd "github.com/PagerDuty/go-pagerduty"
	"github.com/onsi/ginkgo"
	ginkgoConfig "github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/gomega"
	"github.com/spf13/viper"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/aws"
	"github.com/openshift/osde2e/pkg/common/cluster"
	clusterutil "github.com/openshift/osde2e/pkg/common/cluster"
	"github.com/openshift/osde2e/pkg/common/clusterproperties"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/events"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/metadata"
	"github.com/openshift/osde2e/pkg/common/pagerduty"
	"github.com/openshift/osde2e/pkg/common/phase"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/prow"
	"github.com/openshift/osde2e/pkg/common/runner"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/upgrade"
	"github.com/openshift/osde2e/pkg/common/util"
	"github.com/openshift/osde2e/pkg/debug"
	"github.com/openshift/osde2e/pkg/e2e/routemonitors"
	"github.com/openshift/osde2e/pkg/reporting/ginkgorep"
)

const (
	// hiveLog is the name of the hive log file.
	hiveLog string = "hive-log.txt"

	// buildLog is the name of the build log file.
	buildLog string = "test_output.log"
)

// provisioner is used to deploy and manage clusters.
var provider spi.Provider

// --- BEGIN Ginkgo setup
// Check if the test should run
var _ = ginkgo.BeforeEach(func() {
	testText := ginkgo.CurrentGinkgoTestDescription().TestText
	testContext := strings.TrimSpace(strings.TrimSuffix(ginkgo.CurrentGinkgoTestDescription().FullTestText, testText))

	shouldRun := false
	testsToRun := viper.GetStringSlice(config.Tests.TestsToRun)
	for _, testToRun := range testsToRun {
		if strings.HasPrefix(testContext, testToRun) {
			shouldRun = true
			break
		}
	}

	if !shouldRun {
		ginkgo.Skip(fmt.Sprintf("test %s will not be run as its context (%s) is not specified as part of the tests to run", ginkgo.CurrentGinkgoTestDescription().FullTestText, testContext))
	}
})

// beforeSuite attempts to populate several required cluster fields (either by provisioning a new cluster, or re-using an existing one)
// If there is an issue with provisioning, retrieving, or getting the kubeconfig, this will return `false`.
func beforeSuite() bool {
	// Skip provisioning if we already have a kubeconfig
	if viper.GetString(config.Kubeconfig.Contents) == "" {

		cluster, err := clusterutil.ProvisionCluster(nil)
		events.HandleErrorWithEvents(err, events.InstallSuccessful, events.InstallFailed)
		if err != nil {
			log.Printf("Failed to set up or retrieve cluster: %v", err)
			getLogs()
			return false
		}

		viper.Set(config.Cluster.ID, cluster.ID())
		log.Printf("CLUSTER_ID set to %s from OCM.", viper.GetString(config.Cluster.ID))

		viper.Set(config.Cluster.Name, cluster.Name())
		log.Printf("CLUSTER_NAME set to %s from OCM.", viper.GetString(config.Cluster.Name))

		viper.Set(config.Cluster.Version, cluster.Version())
		log.Printf("CLUSTER_VERSION set to %s from OCM.", viper.GetString(config.Cluster.Version))

		viper.Set(config.CloudProvider.CloudProviderID, cluster.CloudProvider())
		log.Printf("CLOUD_PROVIDER_ID set to %s from OCM.", viper.GetString(config.CloudProvider.CloudProviderID))

		viper.Set(config.CloudProvider.Region, cluster.Region())
		log.Printf("CLOUD_PROVIDER_REGION set to %s from OCM.", viper.GetString(config.CloudProvider.Region))

		if !viper.GetBool(config.Addons.SkipAddonList) || viper.GetString(config.Provider) != "mock" {
			log.Printf("Found addons: %s", strings.Join(cluster.Addons(), ","))
		}

		metadata.Instance.SetClusterName(cluster.Name())
		metadata.Instance.SetClusterID(cluster.ID())
		metadata.Instance.SetRegion(cluster.Region())

		if err = provider.AddProperty(cluster, "UpgradeVersion", viper.GetString(config.Upgrade.ReleaseName)); err != nil {
			log.Printf("Error while adding upgrade version property to cluster via OCM: %v", err)
		}

		if viper.GetString(config.Tests.SkipClusterHealthChecks) != "true" {
			if err := clusterutil.WaitForClusterReadyPostInstall(cluster.ID(), nil); err != nil {
				log.Printf("Cluster failed health check: %v", err)
				getLogs()
				return false
			}
			log.Println("Cluster is healthy and ready for testing")
		} else {
			log.Println("Skipping health checks as requested")
		}
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

		var kubeconfigBytes []byte
		if kubeconfigBytes, err = provider.ClusterKubeconfig(cluster.ID()); err != nil {
			events.HandleErrorWithEvents(err, events.InstallKubeconfigRetrievalSuccess, events.InstallKubeconfigRetrievalFailure)
			log.Printf("Failed retrieving kubeconfig: %v", err)
			getLogs()
			return false
		}
		viper.Set(config.Kubeconfig.Contents, string(kubeconfigBytes))

		if len(viper.GetString(config.Kubeconfig.Contents)) == 0 {
			// Give the cluster some breathing room.
			log.Println("OSD cluster installed. Sleeping for 600s.")
			time.Sleep(600 * time.Second)
		} else {
			log.Printf("No kubeconfig contents found, but there should be some by now.")
		}
		getLogs()
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
		err := ioutil.WriteFile(filePath, v, os.ModePerm)
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
		if err = cluster.WaitForClusterReadyPostInstall(clusterID, nil); err != nil {
			return fmt.Errorf("failed waiting for cluster ready: %v", err)
		}
	}

	return nil
}

// -- END Ginkgo setup

// RunTests initializes Ginkgo and runs the osde2e test suite.
func RunTests() bool {
	testing.Init()

	if err := runGinkgoTests(); err != nil {
		log.Printf("Tests failed: %v", err)
		return false
	}

	return true
}

// runGinkgoTests runs the osde2e test suite using Ginkgo.
// nolint:gocyclo
func runGinkgoTests() error {
	var err error

	gomega.RegisterFailHandler(ginkgo.Fail)

	ginkgoConfig.DefaultReporterConfig.NoisySkippings = !viper.GetBool(config.Tests.SuppressSkipNotifications)
	ginkgoConfig.GinkgoConfig.SkipString = viper.GetString(config.Tests.GinkgoSkip)
	ginkgoConfig.GinkgoConfig.FocusString = viper.GetString(config.Tests.GinkgoFocus)
	ginkgoConfig.GinkgoConfig.DryRun = viper.GetBool(config.DryRun)

	if ginkgoConfig.GinkgoConfig.DryRun {
		// Draw attention to DRYRUN as it can exist in ENV.
		log.Println(string("\x1b[33m"), "WARNING! This is a DRY RUN. Review this state if outcome is unexpected.", string("\033[0m"))
	}

	// setup reporter
	reportDir := viper.GetString(config.ReportDir)
	if reportDir == "" {
		reportDir, err = ioutil.TempDir("", "")

		if err != nil {
			return fmt.Errorf("error creating temporary directory: %v", err)
		}

		log.Printf("Writing files to temporary directory %s", reportDir)
		viper.Set(config.ReportDir, reportDir)
	} else if err = os.Mkdir(reportDir, os.ModePerm); err != nil {
		log.Printf("Could not create reporter directory: %v", err)
	}

	// Redirect stdout to where we want it to go
	buildLogPath := filepath.Join(reportDir, buildLog)
	buildLogWriter, err := os.Create(buildLogPath)

	if err != nil {
		return fmt.Errorf("unable to create build log in report directory: %v", err)
	}

	mw := io.MultiWriter(os.Stdout, buildLogWriter)
	log.SetOutput(mw)

	log.Printf("Outputting log to build log at %s", buildLogPath)

	// Get the cluster ID now to test against later
	clusterID := viper.GetString(config.Cluster.ID)
	providerCfg := viper.GetString(config.Provider)
	// setup OSD unless Kubeconfig is present
	if len(viper.GetString(config.Kubeconfig.Path)) > 0 && providerCfg == "mock" {
		log.Print("Found an existing Kubeconfig!")
		if provider, err = providers.ClusterProvider(); err != nil {
			return fmt.Errorf("could not setup cluster provider: %v", err)
		}
		metadata.Instance.SetEnvironment(provider.Environment())
	} else {
		if provider, err = providers.ClusterProvider(); err != nil {
			return fmt.Errorf("could not setup cluster provider: %v", err)
		}

		metadata.Instance.SetEnvironment(provider.Environment())

		// configure cluster and upgrade versions
		if err = ChooseVersions(); err != nil {
			log.Printf("failed to configure versions: %v", err)
			return nil
		}

		switch {
		case !viper.GetBool(config.Cluster.EnoughVersionsForOldestOrMiddleTest):
			log.Printf("There were not enough available cluster image sets to choose and oldest or middle cluster image set to test against. Skipping tests.")
			return nil
		case !viper.GetBool(config.Cluster.PreviousVersionFromDefaultFound):
			log.Printf("No previous version from default found with the given arguments.")
			return nil
		case viper.GetBool(config.Upgrade.UpgradeVersionEqualToInstallVersion):
			log.Printf("Install version and upgrade version are the same. Skipping tests.")
			return nil
		case viper.GetString(config.Upgrade.ReleaseName) == util.NoVersionFound:
			log.Printf("No valid upgrade versions were found. Skipping tests.")
			return nil
		case viper.GetString(config.Upgrade.Image) != "" && viper.GetBool(config.Upgrade.ManagedUpgrade):
			log.Printf("image-based managed upgrades are unsupported: %s", viper.GetString(config.Upgrade.Image))
			return nil
		}
	}

	// Update the metadata object to use the report directory.
	metadata.Instance.SetReportDir(reportDir)

	log.Println("Running e2e tests...")

	if viper.GetString(config.Suffix) == "" {
		viper.Set(config.Suffix, util.RandomStr(5))
	}

	testsPassed := runTestsInPhase(phase.InstallPhase, "OSD e2e suite", ginkgoConfig.GinkgoConfig.DryRun)
	getLogs()
	upgradeTestsPassed := true

	// upgrade cluster if requested
	if viper.GetString(config.Upgrade.Image) != "" || viper.GetString(config.Upgrade.ReleaseName) != "" {

		if len(viper.GetString(config.Kubeconfig.Contents)) > 0 {
			// create route monitors for the upgrade
			var routeMonitorChan chan struct{}
			closeMonitorChan := make(chan struct{})
			if viper.GetBool(config.Upgrade.MonitorRoutesDuringUpgrade) && !ginkgoConfig.GinkgoConfig.DryRun {
				routeMonitorChan = setupRouteMonitors(closeMonitorChan)
				log.Println("Route Monitors created.")
			}

			// run the upgrade
			if err = upgrade.RunUpgrade(); err != nil {
				events.RecordEvent(events.UpgradeFailed)
				return fmt.Errorf("error performing upgrade: %v", err)
			}
			events.RecordEvent(events.UpgradeSuccessful)

			// test upgrade rescheduling if desired
			if !viper.GetBool(config.Upgrade.ManagedUpgradeRescheduled) {
				log.Println("Running e2e tests POST-UPGRADE...")
				upgradeTestsPassed = runTestsInPhase(phase.UpgradePhase, "OSD e2e suite post-upgrade", ginkgoConfig.GinkgoConfig.DryRun)
			}
			log.Println("Upgrade rescheduled, skip the POST-UPGRADE testing")

			// close route monitors
			if viper.GetBool(config.Upgrade.MonitorRoutesDuringUpgrade) && !ginkgoConfig.GinkgoConfig.DryRun {
				close(routeMonitorChan)
				_ = <-closeMonitorChan
				log.Println("Route monitors reconciled")
			}

		} else {
			log.Println("No Kubeconfig found from initial cluster setup. Unable to run upgrade.")
		}
	}

	if reportDir != "" {
		if err = metadata.Instance.WriteToJSON(reportDir); err != nil {
			return fmt.Errorf("error while writing the custom metadata: %v", err)
		}

		// TODO: SDA-2594 Hotfix
		//checkBeforeMetricsGeneration()

		newMetrics := NewMetrics()
		if newMetrics == nil {
			return fmt.Errorf("error getting new metrics provider")
		}
		prometheusFilename, err := newMetrics.WritePrometheusFile(reportDir)
		if err != nil {
			return fmt.Errorf("error while writing prometheus metrics: %v", err)
		}

		jobName := viper.GetString(config.JobName)
		if jobName == "" {
			log.Printf("Skipping metrics upload for local osde2e run.")
		} else if strings.HasPrefix(jobName, "rehearse-") {
			log.Printf("Job %s is a rehearsal, so metrics upload is being skipped.", jobName)
		} else {
			if err := uploadFileToMetricsBucket(filepath.Join(reportDir, prometheusFilename)); err != nil {
				return fmt.Errorf("error while uploading prometheus metrics: %v", err)
			}
		}
	}

	if viper.GetBool(config.Cluster.DestroyAfterTest) {
		log.Printf("Destroying cluster '%s'...", clusterID)

		if err = provider.DeleteCluster(clusterID); err != nil {
			return fmt.Errorf("error deleting cluster: %s", err.Error())
		}
	} else {
		// When using a local kubeconfig, provider might not be set
		if provider != nil {
			log.Printf("For debugging, please look for cluster ID %s in environment %s", viper.GetString(config.Cluster.ID), provider.Environment())
		}
	}

	if !ginkgoConfig.GinkgoConfig.DryRun {
		getLogs()

		h := helper.NewOutsideGinkgo()

		if h == nil {
			return fmt.Errorf("Unable to generate helper object for cleanup")
		}

		cleanupAfterE2E(h)

	}

	if !testsPassed || !upgradeTestsPassed {
		return fmt.Errorf("please inspect logs for more details")
	}

	return nil
}

func openPDAlerts(suites []junit.Suite, jobName, jobURL string) {
	pdc := pagerduty.Config{
		IntegrationKey: viper.GetString(config.Alert.PagerDutyAPIToken),
	}
	failingTests := []string{}
	for _, suite := range suites {
	inner:
		for _, testcase := range suite.Tests {
			if testcase.Status != junit.StatusFailed {
				continue inner
			}
			failingTests = append(failingTests, testcase.Name)
		}
	}
	jobDetails := map[string]string{
		"details":        jobURL,
		"clusterID":      viper.GetString(config.Cluster.ID),
		"clusterName":    viper.GetString(config.Cluster.Name),
		"clusterVersion": viper.GetString(config.Cluster.Version),
		"expiration":     "clusters expire 6 hours after creation",
	}
	// if too many things failed, open a single alert that isn't grouped with the others.
	if len(failingTests) > 10 {
		jobDetails["help"] = "This is likely a more complex problem, like a test harness or infrastructure issue. The test harness will attempt to notify #sd-cicd"
		if event, err := pdc.FireAlert(pd.V2Payload{
			Summary:  "A lot of tests failed together",
			Severity: "info",
			Source:   jobName,
			Group:    "", // do not group
			Details:  jobDetails,
		}); err != nil {
			log.Printf("Failed creating pagerduty incident for failure: %v", err)
		} else {
			if err := alert.SendSlackMessage("sd-cicd", fmt.Sprintf(`@osde2e A bunch of tests failed at once:
pipeline: %s
URL: %s
PD info: %v`, jobName, jobURL, event)); err != nil {
				log.Printf("Failed sending slack message to CICD team: %v", err)
			}
		}
		return
	}
	// open an alert for each failing test
	for _, name := range failingTests {
		if _, err := pdc.FireAlert(pd.V2Payload{
			Summary:  name + " failed",
			Severity: "info",
			Source:   jobName,
			Group:    name, // group by test case
			Details:  jobDetails,
		}); err != nil {
			log.Printf("Failed creating pagerduty incident for failure: %v", err)
		}
	}
	return
}

func cleanupAfterE2E(h *helper.H) (errors []error) {
	var err error
	defer ginkgo.GinkgoRecover()

	if viper.GetBool(config.MustGather) {
		log.Print("Running Must Gather...")
		mustGatherTimeoutInSeconds := 1800
		h.SetServiceAccount("system:serviceaccount:%s:cluster-admin")
		r := h.Runner(fmt.Sprintf("oc adm must-gather --dest-dir=%v", runner.DefaultRunner.OutputDir))
		r.Name = "must-gather"
		r.Tarball = true
		stopCh := make(chan struct{})
		err := r.Run(mustGatherTimeoutInSeconds, stopCh)

		if err != nil {
			log.Printf("Error running must-gather: %s", err.Error())
		} else {
			gatherResults, err := r.RetrieveResults()
			if err != nil {
				log.Printf("Error retrieving must-gather results: %s", err.Error())
			} else {
				h.WriteResults(gatherResults)
			}
		}
	}

	log.Print("Gathering Test Project State...")
	h.InspectState()

	log.Print("Gathering OLM State...")
	h.InspectOLM()

	log.Print("Gathering Cluster State...")
	clusterState := h.GetClusterState()
	stateResults := make(map[string][]byte, len(clusterState))
	for resource, list := range clusterState {
		data, err := json.MarshalIndent(list, "", "    ")
		if err != nil {
			log.Printf("error marshalling JSON for %s/%s/%s", resource.Group, resource.Version, resource.Resource)
		} else {
			var gbuf bytes.Buffer
			zw := gzip.NewWriter(&gbuf)
			_, err = zw.Write(data)
			if err != nil {
				log.Print("Error writing data to buffer")
			}
			err = zw.Close()
			if err != nil {
				log.Print("Error closing writer to buffer")
			}
			// include gzip in filename to mark compressed data
			filename := fmt.Sprintf("%s-%s-%s.json.gzip", resource.Group, resource.Version, resource.Resource)
			stateResults[filename] = gbuf.Bytes()
		}
	}

	// write results to disk
	log.Println("Writing cluster state results")
	h.WriteResults(stateResults)

	clusterID := viper.GetString(config.Cluster.ID)
	if len(clusterID) > 0 {
		if provider, err = providers.ClusterProvider(); err != nil {
			log.Printf("Error getting cluster provider: %s", err.Error())
		}

		// Get state from Provisioner
		log.Printf("Gathering cluster state from %s", provider.Type())

		cluster, err := provider.GetCluster(clusterID)
		if err != nil {
			log.Printf("error getting Cluster state: %s", err.Error())
		} else {
			defer func() {
				// set the completed property right before this function returns, which should be after
				// all cleanup is finished.
				err = provider.AddProperty(cluster, clusterproperties.Status, clusterproperties.StatusCompleted)
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

	// Do any addon cleanup if configured
	log.Printf("Addon cleanup: %v", viper.GetBool(config.Addons.RunCleanup))
	if viper.GetBool(config.Addons.RunCleanup) {
		// By default, use the existing test harnesses for cleanup
		harnesses := strings.Split(viper.GetString(config.Addons.TestHarnesses), ",")
		arguments := []string{"cleanup"}

		// Check if cleanup harnesses exist and if so, use those instead
		cleanupHarnesses := viper.GetString(config.Addons.CleanupHarnesses)
		if len(cleanupHarnesses) > 0 {
			harnesses = strings.Split(cleanupHarnesses, ",")
			arguments = []string{}
		}
		log.Println("Running addon cleanup...")
		h.RunAddonTests("addon-cleanup", 300, harnesses, arguments)
	}

	// We need to clean up our helper tests manually.
	h.Cleanup()

	// We need a provider to hibernate
	// We need a cluster to hibernate
	// We need to check that the test run wants to hibernate after this run
	/*
		if provider != nil && viper.GetString(config.Cluster.ID) != "" && viper.GetBool(config.Cluster.HibernateAfterUse) {
			msg := "Unable to hibernate %s"
			if provider.Hibernate(viper.GetString(config.Cluster.ID)) {
				msg = "Hibernating %s"
			}
			log.Printf(msg, viper.GetString(config.Cluster.ID))
		}
	*/
	return errors
}

// nolint:gocyclo
func runTestsInPhase(phase string, description string, dryrun bool) bool {
	viper.Set(config.Phase, phase)
	reportDir := viper.GetString(config.ReportDir)
	phaseDirectory := filepath.Join(reportDir, phase)
	if _, err := os.Stat(phaseDirectory); os.IsNotExist(err) {
		if err := os.Mkdir(phaseDirectory, os.FileMode(0755)); err != nil {
			log.Printf("error while creating phase directory %s", phaseDirectory)
			return false
		}
	}
	suffix := viper.GetString(config.Suffix)
	phaseReportPath := filepath.Join(phaseDirectory, fmt.Sprintf("junit_%v.xml", suffix))
	phaseReporter := ginkgorep.NewPhaseReporter(phase, phaseReportPath)
	ginkgoPassed := false

	if !dryrun || !ginkgoConfig.GinkgoConfig.DryRun {
		if !beforeSuite() {
			log.Println("Error getting kubeconfig from beforeSuite function")
			return false
		}
	}

	// We need this anonymous function to make sure GinkgoRecover runs where we want it to
	// and will still execute the rest of the function regardless whether the tests pass or fail.
	func() {
		defer ginkgo.GinkgoRecover()
		ginkgoPassed = ginkgo.RunSpecsWithDefaultAndCustomReporters(ginkgo.GinkgoT(), description, []ginkgo.Reporter{phaseReporter})
	}()

	files, err := ioutil.ReadDir(phaseDirectory)
	if err != nil {
		log.Printf("error reading phase directory: %s", err.Error())
		return false
	}

	numTests := 0
	numPassingTests := 0

	for _, file := range files {
		if file != nil {
			// Process the jUnit XML result files
			if junitFileRegex.MatchString(file.Name()) {
				suites, err := junit.IngestFile(filepath.Join(phaseDirectory, file.Name()))
				if err != nil {
					log.Printf("error reading junit xml file %s: %s", file.Name(), err.Error())
					return false
				}

				for _, testSuite := range suites {
					for _, testcase := range testSuite.Tests {
						isSkipped := testcase.Status == junit.StatusSkipped
						isFail := testcase.Status == junit.StatusFailed

						if !isSkipped {
							numTests++
						}
						if !isFail && !isSkipped {
							numPassingTests++
						}
					}
				}

				// fire PD incident if JOB_TYPE==periodic
				if os.Getenv("JOB_TYPE") == "periodic" {
					url, _ := prow.JobURL()
					jobName := os.Getenv("JOB_NAME")
					openPDAlerts(suites, jobName, url)
				}
			}
		}
	}
	// If we could have opened new alerts, consolidate them
	if os.Getenv("JOB_TYPE") == "periodic" {
		err := pagerduty.MergeCICDIncidents(pd.NewClient(viper.GetString(config.Alert.PagerDutyUserToken)))
		if err != nil {
			log.Printf("Failed merging PD incidents: %v", err)
		}
	}

	passRate := float64(numPassingTests) / float64(numTests)

	if math.IsNaN(passRate) {
		log.Printf("Pass rate is NaN: numPassingTests = %d, numTests = %d", numPassingTests, numTests)
	} else {
		metadata.Instance.SetPassRate(phase, passRate)
	}

	files, err = ioutil.ReadDir(reportDir)
	if err != nil {
		log.Printf("error reading phase directory: %s", err.Error())
		return false
	}

	// Ensure all log metrics are zeroed out before running again
	metadata.Instance.ResetLogMetrics()

	//Ensure all before suite metrics are zeroed out before running again
	metadata.Instance.ResetBeforeSuiteMetrics()

	for _, file := range files {
		if logFileRegex.MatchString(file.Name()) {
			data, err := ioutil.ReadFile(filepath.Join(reportDir, file.Name()))
			if err != nil {
				log.Printf("error opening log file %s: %s", file.Name(), err.Error())
				return false
			}
			for _, metric := range config.GetLogMetrics() {
				metadata.Instance.IncrementLogMetric(metric.Name, metric.HasMatches(data))
			}
			for _, metric := range config.GetBeforeSuiteMetrics() {
				metadata.Instance.IncrementBeforeSuiteMetric(metric.Name, metric.HasMatches(data))
			}
		}
	}

	logMetricTestSuite := reporters.JUnitTestSuite{
		Name: "Log Metrics",
	}

	for name, value := range metadata.Instance.LogMetrics {
		testCase := reporters.JUnitTestCase{
			ClassName: "Log Metrics",
			Name:      fmt.Sprintf("[Log Metrics] %s", name),
			Time:      float64(value),
		}

		if config.GetLogMetrics().GetMetricByName(name).IsPassing(value) {
			testCase.PassedMessage = &reporters.JUnitPassedMessage{
				Message: fmt.Sprintf("Passed with %d matches", value),
			}
		} else {
			testCase.FailureMessage = &reporters.JUnitFailureMessage{
				Message: fmt.Sprintf("Failed with %d matches", value),
			}
			logMetricTestSuite.Failures++
		}
		logMetricTestSuite.Tests++

		logMetricTestSuite.TestCases = append(logMetricTestSuite.TestCases, testCase)
	}

	data, err := xml.Marshal(&logMetricTestSuite)

	err = ioutil.WriteFile(filepath.Join(phaseDirectory, "junit_logmetrics.xml"), data, 0644)
	if err != nil {
		log.Printf("error writing to junit file: %s", err.Error())
		return false
	}

	beforeSuiteMetricTestSuite := reporters.JUnitTestSuite{
		Name: "Before Suite Metrics",
	}

	for name, value := range metadata.Instance.BeforeSuiteMetrics {
		testCase := reporters.JUnitTestCase{
			ClassName: "Before Suite Metrics",
			Name:      fmt.Sprintf("[BeforeSuite] %s", name),
			Time:      float64(value),
		}

		if config.GetBeforeSuiteMetrics().GetMetricByName(name).IsPassing(value) {
			testCase.PassedMessage = &reporters.JUnitPassedMessage{
				Message: fmt.Sprintf("Passed with %d matches", value),
			}
		} else {
			testCase.FailureMessage = &reporters.JUnitFailureMessage{
				Message: fmt.Sprintf("Failed with %d matches", value),
			}
			beforeSuiteMetricTestSuite.Failures++
		}
		beforeSuiteMetricTestSuite.Tests++

		beforeSuiteMetricTestSuite.TestCases = append(beforeSuiteMetricTestSuite.TestCases, testCase)
	}

	newdata, err := xml.Marshal(&beforeSuiteMetricTestSuite)

	err = ioutil.WriteFile(filepath.Join(phaseDirectory, "junit_beforesuite.xml"), newdata, 0644)
	if err != nil {
		log.Printf("error writing to junit file: %s", err.Error())
		return false
	}

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
	if !dryrun && clusterState == spi.ClusterStateReady {
		h := helper.NewOutsideGinkgo()
		if h == nil {
			log.Println("Unable to generate helper outside of ginkgo")
			return ginkgoPassed
		}
		dependencies, err := debug.GenerateDependencies(h.Kube())
		if err != nil {
			log.Printf("Error generating dependencies: %s", err.Error())
		} else {
			if err = ioutil.WriteFile(filepath.Join(phaseDirectory, "dependencies.txt"), []byte(dependencies), 0644); err != nil {
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

// uploadFileToMetricsBucket uploads the given file (with absolute path) to the metrics S3 bucket "incoming" directory.
func uploadFileToMetricsBucket(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	return aws.WriteToS3(aws.CreateS3URL(viper.GetString(config.Tests.MetricsBucket), "incoming", filepath.Base(filename)), data)
}

// setupRouteMonitors initializes performance+availability monitoring of cluster routes,
// returning a channel which can be used to terminate the monitoring.
func setupRouteMonitors(closeChannel chan struct{}) chan struct{} {
	routeMonitorChan := make(chan struct{})
	go func() {
		// Set up the route monitors
		routeMonitors, err := routemonitors.Create()
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
					routeMonitors.Plots[msg.Attack].Add(msg)
				}
			}
		}()

		// Close down route monitoring when signalled to
		for {
			select {
			case <-routeMonitorChan:
				log.Println("Closing route monitors...")
				routeMonitors.End()
				routeMonitors.SaveReports(viper.GetString(config.ReportDir))
				routeMonitors.SavePlots(viper.GetString(config.ReportDir))
				routeMonitors.ExtractData(viper.GetString(config.ReportDir))
				routeMonitors.StoreMetadata()
				close(closeChannel)
				return
			}
		}
	}()
	return routeMonitorChan
}
