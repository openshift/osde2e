// Package common launches an OSD cluster, performs tests on it, and destroys it.
package common

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/onsi/ginkgo"
	ginkgoConfig "github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/config"
	"github.com/openshift/osde2e/pkg/events"
	"github.com/openshift/osde2e/pkg/helper"
	"github.com/openshift/osde2e/pkg/metadata"
	"github.com/openshift/osde2e/pkg/osd"
	"github.com/openshift/osde2e/pkg/state"
	"github.com/openshift/osde2e/pkg/upgrade"

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

// RunE2ETests runs the osde2e test suite using.
func RunE2ETests(t *testing.T) {
	var err error
	gomega.RegisterFailHandler(ginkgo.Fail)

	cfg := config.Instance

	ginkgoConfig.GinkgoConfig.SkipString = cfg.Tests.GinkgoSkip
	ginkgoConfig.GinkgoConfig.FocusString = cfg.Tests.GinkgoFocus

	state := state.Instance

	// setup OSD unless Kubeconfig is present
	if len(cfg.Kubeconfig.Path) > 0 {
		log.Print("Found an existing Kubeconfig!")
	} else {
		if OSD, err = osd.New(cfg.OCM.Token, cfg.OCM.Env, cfg.OCM.Debug); err != nil {
			t.Fatalf("could not setup OSD: %v", err)
		}

		metadata.Instance.Environment = cfg.OCM.Env

		// check that enough quota exists for this test if creating cluster
		if len(state.Cluster.ID) == 0 {
			if enoughQuota, err := OSD.CheckQuota(); err != nil {
				log.Printf("Failed to check if enough quota is available: %v", err)
			} else if !enoughQuota {
				t.Fatal("Currently not enough quota exists to run this test, failing...")
			}
		}

		// configure cluster and upgrade versions
		if err = ChooseVersions(OSD); err != nil {
			t.Fatalf("failed to configure versions: %v", err)
		}
	}

	// setup reporter
	if err = os.Mkdir(cfg.ReportDir, os.ModePerm); err != nil {
		log.Printf("Could not create reporter directory: %v", err)
	}

	if !cfg.DryRun {
		log.Println("Running e2e tests...")

		runTestsInPhase(t, "install", "OSD e2e suite")

		// upgrade cluster if requested
		if state.Upgrade.Image != "" || cfg.Upgrade.ReleaseStream != "" {
			if state.Kubeconfig.Contents != nil {
				if err = upgrade.RunUpgrade(OSD); err != nil {
					events.RecordEvent(events.UpgradeFailed)
					t.Errorf("error performing upgrade: %v", err)
				}
				events.RecordEvent(events.UpgradeSuccessful)

				log.Println("Running e2e tests POST-UPGRADE...")
				runTestsInPhase(t, "upgrade", "OSD e2e suite post-upgrade")
			} else {
				log.Println("No Kubeconfig found from initial cluster setup. Unable to run upgrade.")
			}
		}

		if cfg.ReportDir != "" {
			if err = metadata.Instance.WriteToJSON(cfg.ReportDir); err != nil {
				t.Errorf("error while writing the custom metadata: %v", err)
			}

			checkBeforeMetricsGeneration()

			prometheusFilename, err := NewMetrics().WritePrometheusFile(cfg.ReportDir)
			if err != nil {
				t.Errorf("error while writing prometheus metrics: %v", err)
			}

			if cfg.Tests.UploadMetrics {
				if strings.HasPrefix(cfg.JobName, "rehearse-") {
					log.Printf("Job %s is a rehearsal, so metrics upload is being skipped.", cfg.JobName)
				} else {
					if err := uploadFileToMetricsBucket(filepath.Join(cfg.ReportDir, prometheusFilename)); err != nil {
						t.Errorf("error while uploading prometheus metrics: %v", err)
					}
				}
			}
		}

		if OSD != nil {
			if cfg.Cluster.DestroyAfterTest {
				log.Printf("Destroying cluster '%s'...", state.Cluster.ID)
				if err = OSD.DeleteCluster(state.Cluster.ID); err != nil {
					t.Errorf("error deleting cluster: %s", err.Error())
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
	}
}

func runTestsInPhase(t *testing.T, phase string, description string) {
	cfg := config.Instance
	state := state.Instance

	state.Phase = phase
	phaseDirectory := filepath.Join(cfg.ReportDir, phase)
	if _, err := os.Stat(phaseDirectory); os.IsNotExist(err) {
		if err := os.Mkdir(phaseDirectory, os.FileMode(0755)); err != nil {
			t.Fatalf("error while creating phase directory %s", phaseDirectory)
		}
	}
	phaseReportPath := filepath.Join(phaseDirectory, fmt.Sprintf("junit_%v.xml", cfg.Suffix))
	phaseReporter := reporters.NewJUnitReporter(phaseReportPath)
	ginkgo.RunSpecsWithDefaultAndCustomReporters(t, description, []ginkgo.Reporter{phaseReporter})

	files, err := ioutil.ReadDir(phaseDirectory)
	if err != nil {
		t.Fatalf("error reading phase directory: %s", err.Error())
	}

	for _, file := range files {
		if file != nil {
			// Process the jUnit XML result files
			if junitFileRegex.MatchString(file.Name()) {
				data, err := ioutil.ReadFile(filepath.Join(phaseDirectory, file.Name()))
				if err != nil {
					t.Fatalf("error opening junit file %s: %s", file.Name(), err.Error())
				}
				// Use Ginkgo's JUnitTestSuite to unmarshal the JUnit XML file
				var testSuite reporters.JUnitTestSuite

				if err = xml.Unmarshal(data, &testSuite); err != nil {
					t.Fatalf("error unmarshalling junit xml: %s", err.Error())
				}

				for i, testcase := range testSuite.TestCases {
					testSuite.TestCases[i].Name = fmt.Sprintf("[%s] %s", phase, testcase.Name)
				}

				data, err = xml.Marshal(&testSuite)

				err = ioutil.WriteFile(filepath.Join(phaseDirectory, file.Name()), data, 0644)
				if err != nil {
					t.Fatalf("error writing to junit file: %s", err.Error())
				}
			}
		}
	}
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
