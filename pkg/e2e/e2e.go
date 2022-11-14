// Package e2e launches an OSD cluster, performs tests on it, and destroys it.
package e2e

import (
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgtype"
	junit "github.com/joshdk/go-junit"
	vegeta "github.com/tsenart/vegeta/lib"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	pd "github.com/PagerDuty/go-pagerduty"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	"github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/db"

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
	// Skip provisioning if we already have a kubeconfig
	var err error

	// We can capture this error if TEST_KUBECONFIG is set, but we can't use it to skip provisioning
	config.LoadKubeconfig()

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
		if viper.Get(config.Addons.IDs) != nil {
			passthruSecrets := viper.GetStringMapString(config.NonOSDe2eSecrets)
			passthruSecrets["CLUSTER_ID"] = viper.GetString(config.Cluster.ID)
			viper.Set(config.NonOSDe2eSecrets, passthruSecrets)
		}

		viper.Set(config.Cluster.Name, cluster.Name())
		log.Printf("CLUSTER_NAME set to %s from OCM.", viper.GetString(config.Cluster.Name))

		viper.Set(config.Cluster.Version, cluster.Version())
		log.Printf("CLUSTER_VERSION set to %s from OCM.", viper.GetString(config.Cluster.Version))

		viper.Set(config.CloudProvider.CloudProviderID, cluster.CloudProvider())
		log.Printf("CLOUD_PROVIDER_ID set to %s from OCM.", viper.GetString(config.CloudProvider.CloudProviderID))

		viper.Set(config.CloudProvider.Region, cluster.Region())
		log.Printf("CLOUD_PROVIDER_REGION set to %s from OCM.", viper.GetString(config.CloudProvider.Region))

		if (!viper.GetBool(config.Addons.SkipAddonList) || viper.GetString(config.Provider) != "mock") && len(cluster.Addons()) > 0 {
			log.Printf("Found addons: %s", strings.Join(cluster.Addons(), ","))
		}

		metadata.Instance.SetClusterName(cluster.Name())
		metadata.Instance.SetClusterID(cluster.ID())
		metadata.Instance.SetRegion(cluster.Region())

		if err = provider.AddProperty(cluster, "UpgradeVersion", viper.GetString(config.Upgrade.ReleaseName)); err != nil {
			log.Printf("Error while adding upgrade version property to cluster via OCM: %v", err)
		}

		if viper.GetString(config.Tests.SkipClusterHealthChecks) != "true" {
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
		}

		var kubeconfigBytes []byte
		clusterConfigerr := wait.PollImmediate(2*time.Second, 5*time.Minute, func() (bool, error) {
			kubeconfigBytes, err = provider.ClusterKubeconfig(viper.GetString(config.Cluster.ID))
			if err != nil {
				log.Printf("Failed to get kubeconfig from OCM: %v\nWaiting two seconds before retrying", err)
				return false, err
			} else {
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

		getLogs()

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

	// If there are test harnesses present, we need to populate the
	// secrets into the test cluster
	if viper.GetString(config.Addons.TestHarnesses) != "" {
		secretsNamespace := "ci-secrets"
		h := helper.NewOutsideGinkgo()
		h.CreateProject(context.TODO(), secretsNamespace)

		_, err := h.Kube().CoreV1().Secrets("osde2e-"+secretsNamespace).Create(context.TODO(), &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ci-secrets",
				Namespace: "osde2e-" + secretsNamespace,
			},
			StringData: viper.GetStringMapString(config.NonOSDe2eSecrets),
		}, metav1.CreateOptions{})
		if err != nil {
			log.Printf("Error creating Prow secrets in-cluster: %s", err.Error())
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
func RunTests() int {
	var err error
	var exitCode int

	testing.Init()

	exitCode, err = runGinkgoTests()
	if err != nil {
		log.Printf("Tests failed: %v", err)
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

	if skip := viper.GetString(config.Tests.GinkgoSkip); skip != "" {
		suiteConfig.SkipStrings = append(suiteConfig.SkipStrings, skip)
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

	// setup reporter
	reportDir := viper.GetString(config.ReportDir)
	if reportDir == "" {
		reportDir, err = ioutil.TempDir("", "")

		if err != nil {
			return Failure, fmt.Errorf("error creating temporary directory: %v", err)
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
		return Failure, fmt.Errorf("unable to create build log in report directory: %v", err)
	}

	mw := io.MultiWriter(os.Stdout, buildLogWriter)
	log.SetOutput(mw)

	log.Printf("Outputting log to build log at %s", buildLogPath)

	// Get the cluster ID now to test against later
	providerCfg := viper.GetString(config.Provider)
	// setup OSD unless Kubeconfig is present
	if len(viper.GetString(config.Kubeconfig.Path)) > 0 && providerCfg == "mock" {
		log.Print("Found an existing Kubeconfig!")
		if provider, err = providers.ClusterProvider(); err != nil {
			return Failure, fmt.Errorf("could not setup cluster provider: %v", err)
		}
		metadata.Instance.SetEnvironment(provider.Environment())
	} else {
		if provider, err = providers.ClusterProvider(); err != nil {
			return Failure, fmt.Errorf("could not setup cluster provider: %v", err)
		}

		metadata.Instance.SetEnvironment(provider.Environment())

		// configure cluster and upgrade versions
		if err = ChooseVersions(); err != nil {
			// if we fail to choose versions, we should gracefully exit
			return Success, err
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
	}

	// Update the metadata object to use the report directory.
	metadata.Instance.SetReportDir(reportDir)

	log.Println("Running e2e tests...")

	if viper.GetString(config.Suffix) == "" {
		viper.Set(config.Suffix, util.RandomStr(5))
	}

	testsPassed, installTestCaseData := runTestsInPhase(phase.InstallPhase, "OSD e2e suite", suiteConfig, reporterConfig)
	getLogs()
	viper.Set(config.Cluster.Passing, testsPassed)
	upgradeTestsPassed := true
	var upgradeTestCaseData []db.CreateTestcaseParams

	// upgrade cluster if requested
	if viper.GetString(config.Upgrade.Image) != "" || viper.GetString(config.Upgrade.ReleaseName) != "" {
		if len(viper.GetString(config.Kubeconfig.Contents)) > 0 {
			// create route monitors for the upgrade
			var routeMonitorChan chan struct{}
			closeMonitorChan := make(chan struct{})
			if viper.GetBool(config.Upgrade.MonitorRoutesDuringUpgrade) && !suiteConfig.DryRun {
				routeMonitorChan = setupRouteMonitors(context.TODO(), closeMonitorChan)
				log.Println("Route Monitors created.")
			}

			// run the upgrade
			if err = upgrade.RunUpgrade(); err != nil {
				events.RecordEvent(events.UpgradeFailed)
				return Failure, fmt.Errorf("error performing upgrade: %v", err)
			}
			events.RecordEvent(events.UpgradeSuccessful)

			// test upgrade rescheduling if desired
			if !viper.GetBool(config.Upgrade.ManagedUpgradeRescheduled) {
				log.Println("Running e2e tests POST-UPGRADE...")
				viper.Set(config.Cluster.Passing, false)
				upgradeTestsPassed, upgradeTestCaseData = runTestsInPhase(
					phase.UpgradePhase,
					"OSD e2e suite post-upgrade",
					suiteConfig,
					reporterConfig,
				)
				viper.Set(config.Cluster.Passing, upgradeTestsPassed)
			}
			log.Println("Upgrade rescheduled, skip the POST-UPGRADE testing")

			// close route monitors
			if viper.GetBool(config.Upgrade.MonitorRoutesDuringUpgrade) && !suiteConfig.DryRun {
				close(routeMonitorChan)
				_ = <-closeMonitorChan
				log.Println("Route monitors reconciled")
			}

		} else {
			log.Println("No Kubeconfig found from initial cluster setup. Unable to run upgrade.")
		}
	}

	testsFinished := time.Now().UTC()

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		viper.GetString(config.Database.User),
		viper.GetString(config.Database.Pass),
		viper.GetString(config.Database.Host),
		viper.GetString(config.Database.Port),
		viper.GetString(config.Database.DatabaseName),
	)
	// connect to the db
	if viper.GetInt(config.JobID) > 0 {
		log.Printf("Storing data for Job ID: %s", viper.GetString(config.JobID))
		jobData := db.CreateJobParams{
			Provider: viper.GetString(config.Provider),
			JobName:  viper.GetString(config.JobName),
			JobID:    viper.GetString(config.JobID),
			Url: func() string {
				url, _ := prow.JobURL()
				return url
			}(),
			Started: func() time.Time {
				t, _ := time.Parse(time.RFC3339, viper.GetString(config.JobStartedAt))
				return t
			}(),
			Finished:       testsFinished,
			ClusterVersion: viper.GetString(config.Cluster.Version),
			UpgradeVersion: viper.GetString(config.Upgrade.ReleaseName),
			ClusterName:    viper.GetString(config.Cluster.Name),
			ClusterID:      viper.GetString(config.Cluster.ID),
			MultiAz:        viper.GetString(config.Cluster.MultiAZ),
			Channel:        viper.GetString(config.Cluster.Channel),
			Environment:    provider.Environment(),
			Region:         viper.GetString(config.CloudProvider.Region),
			NumbWorkerNodes: func() int32 {
				asString := viper.GetString(config.Cluster.NumWorkerNodes)
				asInt, _ := strconv.Atoi(asString)
				return int32(asInt)
			}(),
			NetworkProvider:    viper.GetString(config.Cluster.NetworkProvider),
			ImageContentSource: viper.GetString(config.Cluster.ImageContentSource),
			InstallConfig:      viper.GetString(config.Cluster.InstallConfig),
			HibernateAfterUse:  viper.GetString(config.Cluster.HibernateAfterUse) == "true",
			Reused:             viper.GetString(config.Cluster.Reused) == "true",
			Result: func() db.JobResult {
				if upgradeTestsPassed && testsPassed {
					return db.JobResultPassed
				}
				return db.JobResultFailed
			}(),
		}
		testData := append(installTestCaseData, upgradeTestCaseData...)
		if err := updateDatabaseAndPagerduty(dbURL, jobData, testData...); err != nil {
			log.Printf("failed updating database or pagerduty: %v", err)
		}
	}

	if reportDir != "" {
		if err = metadata.Instance.WriteToJSON(reportDir); err != nil {
			return Failure, fmt.Errorf("error while writing the custom metadata: %v", err)
		}

		// TODO: SDA-2594 Hotfix
		// checkBeforeMetricsGeneration()

		newMetrics := NewMetrics()
		if newMetrics == nil {
			return Failure, fmt.Errorf("error getting new metrics provider")
		}
		prometheusFilename, err := newMetrics.WritePrometheusFile(reportDir)
		if err != nil {
			return Failure, fmt.Errorf("error while writing prometheus metrics: %v", err)
		}

		jobName := viper.GetString(config.JobName)
		if jobName == "" {
			log.Printf("Skipping metrics upload for local osde2e run.")
		} else if strings.HasPrefix(jobName, "rehearse-") {
			log.Printf("Job %s is a rehearsal, so metrics upload is being skipped.", jobName)
		} else {
			if err := uploadFileToMetricsBucket(filepath.Join(reportDir, prometheusFilename)); err != nil {
				return Failure, fmt.Errorf("error while uploading prometheus metrics: %v", err)
			}
		}
	}

	if !suiteConfig.DryRun {
		getLogs()

		h := helper.NewOutsideGinkgo()

		if h == nil {
			return Failure, fmt.Errorf("Unable to generate helper object for cleanup")
		}

		cleanupAfterE2E(context.TODO(), h)

	}

	if !testsPassed || !upgradeTestsPassed {
		viper.Set(config.Cluster.Passing, false)
		return Failure, fmt.Errorf("please inspect logs for more details")
	}

	if viper.GetBool(config.Cluster.DestroyAfterTest) {
		log.Printf("Destroying cluster '%s'...", viper.GetString(config.Cluster.ID))

		if err = provider.DeleteCluster(viper.GetString(config.Cluster.ID)); err != nil {
			return Failure, fmt.Errorf("error deleting cluster: %s", err.Error())
		}
	} else {
		// When using a local kubeconfig, provider might not be set
		if provider != nil {
			log.Printf("For debugging, please look for cluster ID %s in environment %s", viper.GetString(config.Cluster.ID), provider.Environment())
		}
	}

	return Success, nil
}

// ManyGroupedFailureName is the incident title assigned to incidents reperesenting a large
// cluster of test failures.
const ManyGroupedFailureName = "A lot of tests failed together"

func cleanupAfterE2E(ctx context.Context, h *helper.H) (errors []error) {
	var err error
	clusterStatus := clusterproperties.StatusCompletedFailing
	defer ginkgo.GinkgoRecover()

	if viper.GetBool(config.MustGather) {
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
		h.InspectOLM(ctx)
	}

	log.Print("Gathering Cluster State...")
	clusterState := h.GetClusterState(ctx)
	stateResults := make(map[string][]byte, len(clusterState))
	for resource, list := range clusterState {
		data, err := json.MarshalIndent(list, "", "    ")
		if err != nil {
			log.Printf("error marshalling JSON for %s/%s/%s", resource.Group, resource.Version, resource.Resource)
			clusterStatus = clusterproperties.StatusCompletedError
		} else {
			var gbuf bytes.Buffer
			zw := gzip.NewWriter(&gbuf)
			_, err = zw.Write(data)
			if err != nil {
				log.Print("Error writing data to buffer")
				clusterStatus = clusterproperties.StatusCompletedError
			}
			err = zw.Close()
			if err != nil {
				log.Print("Error closing writer to buffer")
				clusterStatus = clusterproperties.StatusCompletedError
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
			clusterStatus = clusterproperties.StatusCompletedError
		}

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
		h.RunAddonTests(ctx, "addon-cleanup", 300, harnesses, arguments)
	}

	// We need to clean up our helper tests manually.
	h.Cleanup(ctx)

	// If this is a nightly test, we don't want to expire this immediately
	if viper.GetString(config.Cluster.InstallSpecificNightly) != "" || viper.GetString(config.Cluster.ReleaseImageLatest) != "" {
		viper.Set(config.Cluster.HibernateAfterUse, false)
		if viper.GetString(config.Cluster.ID) != "" {
			provider.Expire(viper.GetString(config.Cluster.ID))
		}
	}

	// We need a provider to hibernate
	// We need a cluster to hibernate
	// We need to check that the test run wants to hibernate after this run
	if provider != nil && viper.GetString(config.Cluster.ID) != "" && viper.GetBool(config.Cluster.HibernateAfterUse) && !viper.GetBool(config.Cluster.DestroyAfterTest) {
		msg := "Unable to hibernate %s"
		if provider.Hibernate(viper.GetString(config.Cluster.ID)) {
			msg = "Hibernating %s"
		}
		log.Printf(msg, viper.GetString(config.Cluster.ID))

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
) (bool, []db.CreateTestcaseParams) {
	var testCaseData []db.CreateTestcaseParams
	viper.Set(config.Phase, phase)
	reportDir := viper.GetString(config.ReportDir)
	phaseDirectory := filepath.Join(reportDir, phase)
	if _, err := os.Stat(phaseDirectory); os.IsNotExist(err) {
		if err := os.Mkdir(phaseDirectory, os.FileMode(0o755)); err != nil {
			log.Printf("error while creating phase directory %s", phaseDirectory)
			return false, testCaseData
		}
	}
	suffix := viper.GetString(config.Suffix)
	reporterConfig.JUnitReport = filepath.Join(phaseDirectory, fmt.Sprintf("junit_%v.xml", suffix))
	ginkgoPassed := false

	if !suiteConfig.DryRun {
		if !beforeSuite() {
			log.Println("Error getting kubeconfig from beforeSuite function")
			return false, testCaseData
		}
	}

	// We need this anonymous function to make sure GinkgoRecover runs where we want it to
	// and will still execute the rest of the function regardless whether the tests pass or fail.
	func() {
		defer ginkgo.GinkgoRecover()

		ginkgoPassed = ginkgo.RunSpecs(ginkgo.GinkgoT(), description, suiteConfig, reporterConfig)
	}()

	files, err := ioutil.ReadDir(phaseDirectory)
	if err != nil {
		log.Printf("error reading phase directory: %s", err.Error())
		return false, testCaseData
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
					return false, testCaseData
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

				// record each test case
				for _, suite := range suites {
					for _, test := range suite.Tests {
						testCaseData = append(testCaseData, db.CreateTestcaseParams{
							Result: func(s junit.Status) db.TestResult {
								switch s {
								case "passed":
									return db.TestResultPassed
								case "failure":
									return db.TestResultFailure
								case "skipped":
									return db.TestResultSkipped
								case "error":
									fallthrough
								default:
									return db.TestResultError
								}
							}(test.Status),
							Name: test.Name,
							Duration: pgtype.Interval{
								Microseconds: test.Duration.Microseconds(),
								Status:       pgtype.Present,
							},
							Error: func() string {
								if test.Error != nil {
									return test.Error.Error()
								}
								return ""
							}(),
							Stdout: test.SystemOut,
							Stderr: test.SystemErr,
						})
					}
				}
			}
		}
	}
	// If we could have opened new alerts, consolidate them
	if viper.GetString(config.JobType) == "periodic" {
		err := pagerduty.ProcessCICDIncidents(pd.NewClient(viper.GetString(config.Alert.PagerDutyUserToken)))
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
		return false, testCaseData
	}

	// Ensure all log metrics are zeroed out before running again
	metadata.Instance.ResetLogMetrics()

	// Ensure all before suite metrics are zeroed out before running again
	metadata.Instance.ResetBeforeSuiteMetrics()

	for _, file := range files {
		if logFileRegex.MatchString(file.Name()) {
			data, err := ioutil.ReadFile(filepath.Join(reportDir, file.Name()))
			if err != nil {
				log.Printf("error opening log file %s: %s", file.Name(), err.Error())
				return false, testCaseData
			}
			for _, metric := range config.GetLogMetrics() {
				metadata.Instance.IncrementLogMetric(metric.Name, metric.HasMatches(data))
			}
			for _, metric := range config.GetBeforeSuiteMetrics() {
				metadata.Instance.IncrementBeforeSuiteMetric(metric.Name, metric.HasMatches(data))
			}
		}
	}

	// Flagging this block for deletion.
	// Delete, Refactor, broken
	// logMetricTestSuite := reporters.JUnitTestSuite{
	// 	Name: "Log Metrics",
	// }

	// for name, value := range metadata.Instance.LogMetrics {
	// 	testCase := reporters.JUnitTestCase{
	// 		Classname: "Log Metrics",
	// 		Name:      fmt.Sprintf("[Log Metrics] %s", name),
	// 		Time:      float64(value),
	// 	}

	// 	if config.GetLogMetrics().GetMetricByName(name).IsPassing(value) {
	// 		testCase.SystemOut = fmt.Sprintf("Passed with %d matches", value)
	// 	} else {
	// 		testCase.Failure = &reporters.JUnitFailure{
	// 			Message: fmt.Sprintf("Failed with %d matches", value),
	// 		}
	// 		logMetricTestSuite.Failures++
	// 	}
	// 	logMetricTestSuite.Tests++

	// 	logMetricTestSuite.TestCases = append(logMetricTestSuite.TestCases, testCase)
	// }

	// data, err := xml.Marshal(&logMetricTestSuite)

	// err = ioutil.WriteFile(filepath.Join(phaseDirectory, "junit_logmetrics.xml"), data, 0644)
	// if err != nil {
	// 	log.Printf("error writing to junit file: %s", err.Error())
	// 	return false, testCaseData
	// }

	// beforeSuiteMetricTestSuite := reporters.JUnitTestSuite{
	// 	Name: "Before Suite Metrics",
	// }

	// for name, value := range metadata.Instance.BeforeSuiteMetrics {
	// 	testCase := reporters.JUnitTestCase{
	// 		Classname: "Before Suite Metrics",
	// 		Name:      fmt.Sprintf("[BeforeSuite] %s", name),
	// 		Time:      float64(value),
	// 	}

	// 	if config.GetBeforeSuiteMetrics().GetMetricByName(name).IsPassing(value) {
	// 		testCase.SystemOut = fmt.Sprintf("Passed with %d matches", value)
	// 	} else {
	// 		testCase.Failure = &reporters.JUnitFailure{
	// 			Message: fmt.Sprintf("Failed with %d matches", value),
	// 		}
	// 		beforeSuiteMetricTestSuite.Failures++
	// 	}
	// 	beforeSuiteMetricTestSuite.Tests++

	// 	beforeSuiteMetricTestSuite.TestCases = append(beforeSuiteMetricTestSuite.TestCases, testCase)
	// }

	// newdata, err := xml.Marshal(&beforeSuiteMetricTestSuite)

	// err = ioutil.WriteFile(filepath.Join(phaseDirectory, "junit_beforesuite.xml"), newdata, 0644)
	// if err != nil {
	// 	log.Printf("error writing to junit file: %s", err.Error())
	// 	return false, testCaseData
	// }

	clusterID := viper.GetString(config.Cluster.ID)

	clusterState := spi.ClusterStateUnknown

	if clusterID != "" {
		cluster, err := provider.GetCluster(clusterID)
		if err != nil {
			log.Printf("error getting cluster state after a test run: %v", err)
			return false, testCaseData
		}
		clusterState = cluster.State()
	}
	if !suiteConfig.DryRun && clusterState == spi.ClusterStateReady && viper.GetString(config.JobName) != "" {
		h := helper.NewOutsideGinkgo()
		if h == nil {
			log.Println("Unable to generate helper outside of ginkgo")
			return ginkgoPassed, testCaseData
		}
		dependencies, err := debug.GenerateDependencies(h.Kube())
		if err != nil {
			log.Printf("Error generating dependencies: %s", err.Error())
		} else {
			if err = ioutil.WriteFile(filepath.Join(phaseDirectory, "dependencies.txt"), []byte(dependencies), 0o644); err != nil {
				log.Printf("Error writing dependencies.txt: %s", err.Error())
			}

			err := debug.GenerateDiff(phase, dependencies)
			if err != nil {
				log.Printf("Error generating diff: %s", err.Error())
			}

		}
	}
	return ginkgoPassed, testCaseData
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
func setupRouteMonitors(ctx context.Context, closeChannel chan struct{}) chan struct{} {
	routeMonitorChan := make(chan struct{})
	go func() {
		// Set up the route monitors
		routeMonitors, err := routemonitors.Create(ctx)
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

func updateDatabaseAndPagerduty(dbURL string, jobData db.CreateJobParams, testData ...db.CreateTestcaseParams) error {
	var (
		problematicSet = make(map[string]db.ListProblematicTestsRow)
		alertData      map[string][]db.ListAlertableRecentTestFailuresRow
		jobID          int64
		err            error
	)

	// Record data from this job and extract data that we need to operate on PD.
	log.Printf("Storing data for Job ID: %s", viper.GetString(config.JobID))
	if err := db.WithDB(dbURL, func(pg *sql.DB) error {
		// ensure it's on the latest schema
		if err := db.WithMigrator(pg, func(m *migrate.Migrate) error {
			if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
				return err
			}
			return nil
		}); err != nil {
			return err
		}

		q := db.New(pg)

		// insert this job's info
		jobID, err = q.CreateJob(context.TODO(), jobData)
		if err != nil {
			return fmt.Errorf("failed creating job: %w", err)
		}

		log.Printf("This job has Database ID %d", jobID)

		for _, tc := range testData {
			tc.JobID = jobID
			_, err := q.CreateTestcase(context.TODO(), tc)
			if err != nil {
				return fmt.Errorf("failed creating test case: %w", err)
			}
		}

		alertData, err = q.AlertDataForJob(context.TODO(), jobID)
		if err != nil {
			return fmt.Errorf("failed creating alert data: %w", err)
		}

		problems, err := q.ListProblematicTests(context.TODO())
		if err != nil {
			return fmt.Errorf("failed listing problematic tests: %w", err)
		}

		for _, p := range problems {
			problematicSet[p.Name] = p
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed communicating with database: %w", err)
	}

	// Extract useful data into PD alerts and create the PD alerts.
	pdAlertClient := pagerduty.Config{
		IntegrationKey: viper.GetString(config.Alert.PagerDutyAPIToken),
	}
	alertSource := fmt.Sprintf("job %d", jobID)
	log.Println("Creating pagerduty alerts for job (if any)")
	// if too many things failed, open a single alert that isn't grouped with the others.
	if len(alertData) > 10 {
		event, err := pdAlertClient.FireAlert(pd.V2Payload{
			Summary:  ManyGroupedFailureName,
			Severity: "info",
			Source:   alertSource,
			Group:    "", // do not group
			Details:  alertData,
		})
		if err != nil {
			log.Printf("Failed creating pagerduty incident for failure: %v", err)
		}
		if err := alert.SendSlackMessage("sd-cicd-alerts", fmt.Sprintf(`@osde2e A bunch of tests failed at once:
pipeline: %s
URL: %s
PD info: %v`, jobData.JobName, jobData.Url, event)); err != nil {
			log.Printf("Failed sending slack message to CICD team: %v", err)
		}
	} else {
		// open many alerts
		for name, instances := range alertData {
			sort.Slice(instances, func(i, j int) bool {
				return instances[i].Started.Before(instances[j].Started)
			})
			var oldest, current db.ListAlertableRecentTestFailuresRow
			oldest = instances[0]
			for _, instance := range instances {
				if instance.ID == jobID {
					current = instance
					break
				}
			}
			if _, err := pdAlertClient.FireAlert(pd.V2Payload{
				Summary:  name + " failed",
				Severity: "info",
				Source:   alertSource,
				Group:    name, // group by test case
				Details: map[string]db.ListAlertableRecentTestFailuresRow{
					"oldest":  oldest,
					"current": current,
				},
			}); err != nil {
				return fmt.Errorf("failed creating PD alert: %w", err)
			}
		}
	}

	log.Println("Deduplicating pagerduty incidents")
	// Deduplicate incidents.
	pdClient := pd.NewClient(viper.GetString(config.Alert.PagerDutyUserToken))
	listOptions := pd.ListIncidentsOptions{
		ServiceIDs: []string{"P7VT2V5"},
		Statuses:   []string{"triggered", "acknowledged"},
		Limit:      100,
	}
	if err := pagerduty.EnsureIncidentsMerged(pdClient); err != nil {
		return fmt.Errorf("failed merging incidents: %w", err)
	}

	log.Println("Closing old pagerduty incidents")
	// Find all active PD incidents that aren't in our problem list. We
	// list them again since we just tried to merge some of them.
	var needsClose []pd.ManageIncidentsOptions
	if err := pagerduty.Incidents(
		pdClient,
		listOptions,
		func(incident pd.Incident) error {
			// convert the name to the notation we expect from the database
			name := strings.TrimPrefix(incident.Title, "[install] ")
			name = strings.TrimPrefix(name, "[upgrade] ")
			name = strings.TrimSuffix(name, " failed")
			if name == ManyGroupedFailureName {
				return nil
			}
			if _, ok := problematicSet[name]; !ok {
				// mark this PD incident as needing to be closed
				needsClose = append(needsClose, pd.ManageIncidentsOptions{
					ID:         incident.ID,
					Type:       incident.Type,
					Status:     "resolved",
					Resolution: fmt.Sprintf("Resolved by job %d", jobID),
				})
			}
			return nil
		},
	); err != nil {
		return fmt.Errorf("failed listing open PD incidents: %w", err)
	}

	// Resolve all of the incidents that aren't on our problem list.
	if _, err := pdClient.ManageIncidentsWithContext(context.TODO(), "", needsClose); err != nil {
		return fmt.Errorf("failed resolving incidents: %w", err)
	}
	return nil
}
