// Package e2e launches an OSD cluster, performs tests on it, and destroys it.
package e2e

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/onsi/ginkgo"
	ginkgoConfig "github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/gomega"
	"github.com/spf13/viper"

	"github.com/openshift/osde2e/pkg/common/aws"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/events"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/metadata"
	"github.com/openshift/osde2e/pkg/common/phase"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/runner"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/upgrade"
	"github.com/openshift/osde2e/pkg/common/util"
	"github.com/openshift/osde2e/pkg/debug"
)

const (
	// hiveLog is the name of the hive log file.
	hiveLog string = "hive-log.txt"
)

// provisioner is used to deploy and manage clusters.
var provider spi.Provider

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

	dryRun := viper.GetBool(config.DryRun)

	ginkgoConfig.DefaultReporterConfig.NoisySkippings = !viper.GetBool(config.Tests.SuppressSkipNotifications)
	ginkgoConfig.GinkgoConfig.SkipString = viper.GetString(config.Tests.GinkgoSkip)
	ginkgoConfig.GinkgoConfig.FocusString = viper.GetString(config.Tests.GinkgoFocus)
	ginkgoConfig.GinkgoConfig.DryRun = dryRun

	// Get the cluster ID now to test against later
	clusterID := viper.GetString(config.Cluster.ID)

	// setup OSD unless Kubeconfig is present
	if len(viper.GetString(config.Kubeconfig.Path)) > 0 {
		log.Print("Found an existing Kubeconfig!")
	} else {
		if provider, err = providers.ClusterProvider(); err != nil {
			return fmt.Errorf("could not setup cluster provider: %v", err)
		}

		metadata.Instance.SetEnvironment(provider.Environment())

		// configure cluster and upgrade versions
		if err = ChooseVersions(); err != nil {
			return fmt.Errorf("failed to configure versions: %v", err)
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
		}

		// check that enough quota exists for this test if creating cluster
		if len(clusterID) == 0 {
			if dryRun {
				log.Printf("This is a dry run. Skipping quota check.")
			} else if enoughQuota, err := provider.CheckQuota(); err != nil {
				log.Printf("Failed to check if enough quota is available: %v", err)
			} else if !enoughQuota {
				return fmt.Errorf("currently not enough quota exists to run this test")
			}
		}
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

	// Update the metadata object to use the report directory.
	metadata.Instance.SetReportDir(reportDir)

	log.Println("Running e2e tests...")

	if viper.GetString(config.Suffix) == "" {
		viper.Set(config.Suffix, util.RandomStr(3))
	}

	testsPassed := runTestsInPhase(phase.InstallPhase, "OSD e2e suite")
	upgradeTestsPassed := true

	// upgrade cluster if requested
	if viper.GetString(config.Upgrade.Image) != "" || viper.GetString(config.Upgrade.ReleaseName) != "" {
		if len(viper.GetString(config.Kubeconfig.Contents)) > 0 {
			if err = upgrade.RunUpgrade(provider); err != nil {
				events.RecordEvent(events.UpgradeFailed)
				return fmt.Errorf("error performing upgrade: %v", err)
			}
			events.RecordEvent(events.UpgradeSuccessful)

			log.Println("Running e2e tests POST-UPGRADE...")
			upgradeTestsPassed = runTestsInPhase(phase.UpgradePhase, "OSD e2e suite post-upgrade")
		} else {
			log.Println("No Kubeconfig found from initial cluster setup. Unable to run upgrade.")
		}
	}

	if reportDir != "" {
		if err = metadata.Instance.WriteToJSON(reportDir); err != nil {
			return fmt.Errorf("error while writing the custom metadata: %v", err)
		}

		checkBeforeMetricsGeneration()

		newMetrics := NewMetrics()
		if newMetrics == nil {
			return fmt.Errorf("error getting new metrics provider")
		}
		prometheusFilename, err := newMetrics.WritePrometheusFile(reportDir)
		if err != nil {
			return fmt.Errorf("error while writing prometheus metrics: %v", err)
		}

		if viper.GetBool(config.Tests.UploadMetrics) {
			jobName := viper.GetString(config.JobName)
			if strings.HasPrefix(jobName, "rehearse-") {
				log.Printf("Job %s is a rehearsal, so metrics upload is being skipped.", jobName)
			} else {
				if err := uploadFileToMetricsBucket(filepath.Join(reportDir, prometheusFilename)); err != nil {
					return fmt.Errorf("error while uploading prometheus metrics: %v", err)
				}
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
			log.Printf("For debugging, please look for cluster ID %s in environment %s", clusterID, provider.Environment())
		}
	}

	if !dryRun {
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

func cleanupAfterE2E(h *helper.H) (errors []error) {
	var err error
	defer ginkgo.GinkgoRecover()

	if viper.GetBool(config.MustGather) {
		log.Print("Running Must Gather...")
		mustGatherTimeoutInSeconds := 900
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
	if !viper.GetBool(config.DryRun) {
		h.Cleanup()
	}

	return errors
}

func runTestsInPhase(phase string, description string) bool {
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
	phaseReporter := reporters.NewJUnitReporter(phaseReportPath)
	ginkgoPassed := false

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
				data, err := ioutil.ReadFile(filepath.Join(phaseDirectory, file.Name()))
				if err != nil {
					log.Printf("error opening junit file %s: %s", file.Name(), err.Error())
					return false
				}
				// Use Ginkgo's JUnitTestSuite to unmarshal the JUnit XML file
				var testSuite reporters.JUnitTestSuite

				if err = xml.Unmarshal(data, &testSuite); err != nil {
					log.Printf("error unmarshalling junit xml: %s", err.Error())
					return false
				}

				for i, testcase := range testSuite.TestCases {
					isSkipped := testcase.Skipped != nil
					isFail := testcase.FailureMessage != nil

					if !isSkipped {
						numTests++
					}
					if !isFail && !isSkipped {
						numPassingTests++
					}

					testSuite.TestCases[i].Name = fmt.Sprintf("[%s] %s", phase, testcase.Name)
				}

				data, err = xml.Marshal(&testSuite)

				err = ioutil.WriteFile(filepath.Join(phaseDirectory, file.Name()), data, 0644)
				if err != nil {
					log.Printf("error writing to junit file: %s", err.Error())
					return false
				}
			}
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

	if !viper.GetBool(config.DryRun) && fmt.Sprintf("%v", viper.Get(config.Cluster.State)) == string(spi.ClusterStateReady) {
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

			err := debug.GenerateDiff(viper.GetString(config.BaseJobURL), phase, dependencies, viper.GetString(config.JobName), viper.GetInt(config.JobID))
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

	aws.WriteToS3(aws.CreateS3URL(viper.GetString(config.Tests.MetricsBucket), "incoming", filepath.Base(filename)), data)
	return err
}
