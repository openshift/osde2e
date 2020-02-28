// Package e2e launches an OSD cluster, performs tests on it, and destroys it.
package e2e

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/onsi/ginkgo"
	ginkgoConfig "github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/events"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/metadata"
	"github.com/openshift/osde2e/pkg/common/osd"
	"github.com/openshift/osde2e/pkg/common/state"
	"github.com/openshift/osde2e/pkg/common/upgrade"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// hiveLog is the name of the hive log file.
	hiveLog string = "hive-log.txt"
)

// OSD is used to deploy and manage clusters.
var OSD *osd.OSD

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
func runGinkgoTests() error {
	var err error
	gomega.RegisterFailHandler(ginkgo.Fail)

	cfg := config.Instance

	ginkgoConfig.DefaultReporterConfig.NoisySkippings = !config.Instance.Tests.SuppressSkipNotifications
	ginkgoConfig.GinkgoConfig.SkipString = cfg.Tests.GinkgoSkip
	ginkgoConfig.GinkgoConfig.FocusString = cfg.Tests.GinkgoFocus

	state := state.Instance

	// setup OSD unless Kubeconfig is present
	if len(cfg.Kubeconfig.Path) > 0 {
		log.Print("Found an existing Kubeconfig!")
	} else {
		if OSD, err = osd.New(cfg.OCM.Token, cfg.OCM.Env, cfg.OCM.Debug); err != nil {
			return fmt.Errorf("could not setup OSD: %v", err)
		}

		metadata.Instance.SetEnvironment(cfg.OCM.Env)

		// check that enough quota exists for this test if creating cluster
		if len(state.Cluster.ID) == 0 {
			if enoughQuota, err := OSD.CheckQuota(); err != nil {
				log.Printf("Failed to check if enough quota is available: %v", err)
			} else if !enoughQuota {
				return fmt.Errorf("currently not enough quota exists to run this test")
			}
		}

		// configure cluster and upgrade versions
		if err = ChooseVersions(OSD); err != nil {
			return fmt.Errorf("failed to configure versions: %v", err)
		}

		if state.Upgrade.UpgradeVersionEqualToInstallVersion {
			log.Printf("Install version and upgrade version are the same. Skipping tests.")
			return nil
		}
	}

	// setup reporter
	if err = os.Mkdir(cfg.ReportDir, os.ModePerm); err != nil {
		log.Printf("Could not create reporter directory: %v", err)
	}

	if !cfg.DryRun {
		log.Println("Running e2e tests...")

		testsPassed := runTestsInPhase("install", "OSD e2e suite")
		upgradeTestsPassed := true

		if testsPassed {
			// upgrade cluster if requested
			if state.Upgrade.Image != "" || state.Upgrade.ReleaseName != "" {
				if state.Kubeconfig.Contents != nil {
					if err = upgrade.RunUpgrade(OSD); err != nil {
						events.RecordEvent(events.UpgradeFailed)
						return fmt.Errorf("error performing upgrade: %v", err)
					}
					events.RecordEvent(events.UpgradeSuccessful)

					log.Println("Running e2e tests POST-UPGRADE...")
					upgradeTestsPassed = runTestsInPhase("upgrade", "OSD e2e suite post-upgrade")
				} else {
					log.Println("No Kubeconfig found from initial cluster setup. Unable to run upgrade.")
				}
			}
		} else {
			log.Print("Install tests did not pass. Skipping upgrade tests.")
		}

		if cfg.ReportDir != "" {
			if err = metadata.Instance.WriteToJSON(cfg.ReportDir); err != nil {
				return fmt.Errorf("error while writing the custom metadata: %v", err)
			}

			checkBeforeMetricsGeneration()

			prometheusFilename, err := NewMetrics().WritePrometheusFile(cfg.ReportDir)
			if err != nil {
				return fmt.Errorf("error while writing prometheus metrics: %v", err)
			}

			if cfg.Tests.UploadMetrics {
				if strings.HasPrefix(cfg.JobName, "rehearse-") {
					log.Printf("Job %s is a rehearsal, so metrics upload is being skipped.", cfg.JobName)
				} else {
					if err := uploadFileToMetricsBucket(filepath.Join(cfg.ReportDir, prometheusFilename)); err != nil {
						return fmt.Errorf("error while uploading prometheus metrics: %v", err)
					}
				}
			}
		}

		if OSD != nil {
			if cfg.Cluster.DestroyAfterTest {
				log.Printf("Destroying cluster '%s'...", state.Cluster.ID)
				if err = OSD.DeleteCluster(state.Cluster.ID); err != nil {
					return fmt.Errorf("error deleting cluster: %s", err.Error())
				}
			} else {
				log.Printf("For debugging, please look for cluster ID %s in environment %s", state.Cluster.ID, cfg.OCM.Env)
			}
		} else {
			// If we run against an arbitrary cluster and not a ci-specific cluster
			// we need to clean up our workload tests manually.
			h := &helper.H{
				State: state,
			}
			h.SetupNoProj()

			log.Printf("Cleaning up workloads tests")
			workloads := h.GetWorkloads()
			for _, project := range workloads {
				log.Printf("Deleting Project: %s", project)
				h.SetProjectByName(project)
				h.Project().ProjectV1().Projects().Delete(project, &metav1.DeleteOptions{})
			}
		}

		if !testsPassed || !upgradeTestsPassed {
			return fmt.Errorf("please inspect logs for more details")
		}
	}

	return nil
}

func runTestsInPhase(phase string, description string) bool {
	cfg := config.Instance
	state := state.Instance

	state.Phase = phase
	phaseDirectory := filepath.Join(cfg.ReportDir, phase)
	if _, err := os.Stat(phaseDirectory); os.IsNotExist(err) {
		if err := os.Mkdir(phaseDirectory, os.FileMode(0755)); err != nil {
			log.Printf("error while creating phase directory %s", phaseDirectory)
			return false
		}
	}
	phaseReportPath := filepath.Join(phaseDirectory, fmt.Sprintf("junit_%v.xml", cfg.Suffix))
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

	logMetricsRegexs := make(map[string]*regexp.Regexp)
	for name, match := range cfg.LogMetrics {
		logMetricsRegexs[name] = regexp.MustCompile(match)
	}

	files, err = ioutil.ReadDir(cfg.ReportDir)
	if err != nil {
		log.Printf("error reading phase directory: %s", err.Error())
		return false
	}

	for _, file := range files {
		//log.Printf("Parsing log metrics in %s", filepath.Join(cfg.ReportDir, file.Name()))
		if logFileRegex.MatchString(file.Name()) {
			data, err := ioutil.ReadFile(filepath.Join(cfg.ReportDir, file.Name()))
			if err != nil {
				log.Printf("error opening log file %s: %s", file.Name(), err.Error())
				return false
			}
			for name, matchRegex := range logMetricsRegexs {
				matches := matchRegex.FindAll(data, -1)
				//log.Printf("-- Found %d matches for %s", len(matches), name)
				metadata.Instance.IncrementLogMetric(name, len(matches))
			}
		}
	}

	return ginkgoPassed
}

// checkBeforeMetricsGeneration runs a variety of checks before generating metrics.
func checkBeforeMetricsGeneration() error {
	// Check for hive-log.txt
	if _, err := os.Stat(filepath.Join(config.Instance.ReportDir, hiveLog)); os.IsNotExist(err) {
		events.RecordEvent(events.NoHiveLogs)
	}

	return nil
}

// uploadFileToMetricsBucket uploads the given file (with absolute path) to the metrics S3 bucket "incoming" directory.
func uploadFileToMetricsBucket(filename string) error {
	// We're very intentionally using the shared configs here.
	// This allows us to configure the AWS client at a system level and this should behave as expected.
	// This is particularly useful if we want to, at some point in the future, run this on an AWS host with an instance profile
	// that doesn't need explicit credentials.
	session, err := session.NewSessionWithOptions(session.Options{SharedConfigState: session.SharedConfigEnable})
	if err != nil {
		return err
	}

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	uploader := s3manager.NewUploader(session)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(config.Instance.Tests.MetricsBucket),
		Key:    aws.String(path.Join("incoming", filepath.Base(filename))),
		Body:   file,
	})

	return err
}
