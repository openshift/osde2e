package e2e

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/onsi/ginkgo"
	ginkgoconfig "github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/events"
	"github.com/openshift/osde2e/pkg/common/load"
	"github.com/openshift/osde2e/pkg/common/metadata"
	"github.com/openshift/osde2e/pkg/common/osd"
	"github.com/openshift/osde2e/pkg/common/state"
)

func init() {
	testing.Init()

	// Load config and initial state
	if err := load.IntoObject(config.Instance); err != nil {
		panic(fmt.Errorf("error loading config: %v", err))
	}

	if err := load.IntoObject(state.Instance); err != nil {
		panic(fmt.Errorf("error loading initial state: %v", err))
	}

	ginkgoconfig.DefaultReporterConfig.NoisySkippings = !config.Instance.Tests.SuppressSkipNotifications

	if len(config.Instance.Tests.TestsToRun) > 0 {
		ginkgo.BeforeEach(func() {
			testText := ginkgo.CurrentGinkgoTestDescription().TestText
			testContext := strings.TrimSpace(strings.TrimSuffix(ginkgo.CurrentGinkgoTestDescription().FullTestText, testText))

			shouldRun := false
			for _, testToRun := range config.Instance.Tests.TestsToRun {
				if strings.HasPrefix(testContext, testToRun) {
					shouldRun = true
					break
				}
			}

			if !shouldRun {
				ginkgo.Skip(fmt.Sprintf("test %s will not be run as its context (%s) is not specified as part of the tests to run", ginkgo.CurrentGinkgoTestDescription().FullTestText, testContext))
			}
		})
	}
}

// Setup cluster before testing begins.
var _ = ginkgo.SynchronizedBeforeSuite(func() []byte {
	defer ginkgo.GinkgoRecover()
	cfg := config.Instance
	state := state.Instance

	err := setupCluster()
	events.HandleErrorWithEvents(err, events.InstallSuccessful, events.InstallFailed).ShouldNot(HaveOccurred(), "failed to setup cluster for testing")
	if err != nil {
		return []byte{}
	}

	if len(cfg.Addons.IDs) > 0 {
		err = installAddons()
		events.HandleErrorWithEvents(err, events.InstallAddonsSuccessful, events.InstallAddonsFailed).ShouldNot(HaveOccurred(), "failed while installing addons")
		if err != nil {
			return []byte{}
		}
	}

	if len(state.Kubeconfig.Contents) == 0 {
		// Give the cluster some breathing room.
		log.Println("OSD cluster installed. Sleeping for 600s.")
		time.Sleep(600 * time.Second)
	} else {
		log.Printf("No kubeconfig contents found, but there should be some by now.")
	}

	return []byte{}
}, func(data []byte) {
	// only needs to run once
})

// Collect logs after each test
var _ = ginkgo.AfterSuite(func() {
	log.Printf("Getting logs for cluster '%s'...", state.Instance.Cluster.ID)
	getLogs()
})
var _ = ginkgo.JustAfterEach(getLogs)

func getLogs() {
	defer ginkgo.GinkgoRecover()
	state := state.Instance

	if OSD == nil {
		log.Println("OSD was not configured. Skipping log collection...")
	} else if state.Cluster.ID == "" {
		log.Println("CLUSTER_ID is not set, likely due to a setup failure. Skipping log collection...")
	} else {
		logs, err := OSD.FullLogs(state.Cluster.ID)
		Expect(err).NotTo(HaveOccurred(), "failed to collect cluster logs")
		writeLogs(logs)
	}
}

// setupCluster brings up a cluster, waits for it to be ready, then returns it's name.
func setupCluster() (err error) {
	cfg := config.Instance
	state := state.Instance

	// if TEST_KUBECONFIG has been set, skip configuring OCM
	if len(state.Kubeconfig.Contents) > 0 || len(cfg.Kubeconfig.Path) > 0 {
		return useKubeconfig()
	}

	// create a new cluster if no ID is specified
	if state.Cluster.ID == "" {
		if state.Cluster.Name == "" {
			state.Cluster.Name = clusterName()
		}

		if state.Cluster.ID, err = OSD.LaunchCluster(); err != nil {
			return fmt.Errorf("could not launch cluster: %v", err)
		}
	} else {
		log.Printf("CLUSTER_ID of '%s' was provided, skipping cluster creation and using it instead", state.Cluster.ID)

		if state.Cluster.Name == "" {
			cluster, err := OSD.GetCluster(state.Cluster.ID)
			if err != nil {
				return fmt.Errorf("could not retrieve cluster information from OCM: %v", err)
			}

			if cluster.Name() == "" {
				return fmt.Errorf("cluster name from OCM is empty, and this shouldn't be possible")
			}

			state.Cluster.Name = cluster.Name()
			log.Printf("CLUSTER_NAME not provided, retrieved %s from OCM.", state.Cluster.Name)
		}
	}

	metadata.Instance.SetClusterName(state.Cluster.Name)
	metadata.Instance.SetClusterID(state.Cluster.ID)

	if err = OSD.WaitForClusterReady(); err != nil {
		return fmt.Errorf("failed waiting for cluster ready: %v", err)
	}

	if state.Kubeconfig.Contents, err = OSD.ClusterKubeconfig(state.Cluster.ID); err != nil {
		return fmt.Errorf("could not get kubeconfig for cluster: %v", err)
	}

	return nil
}

// installAddons installs addons onto the cluster
func installAddons() (err error) {
	num, err := OSD.InstallAddons(config.Instance.Addons.IDs)
	if err != nil {
		return fmt.Errorf("could not install addons: %s", err.Error())
	}
	if num > 0 {
		if err = OSD.WaitForClusterReady(); err != nil {
			return fmt.Errorf("failed waiting for cluster ready: %v", err)
		}
	}

	return nil
}

// useKubeconfig reads the path provided for a TEST_KUBECONFIG and uses it for testing.
func useKubeconfig() (err error) {
	cfg := config.Instance
	state := state.Instance

	_, err = clientcmd.RESTConfigFromKubeConfig(state.Kubeconfig.Contents)
	if err != nil {
		log.Println("Not an existing Kubeconfig, attempting to read file instead...")
	} else {
		log.Println("Existing valid kubeconfig!")
		return nil
	}

	state.Kubeconfig.Contents, err = ioutil.ReadFile(cfg.Kubeconfig.Path)
	if err != nil {
		return fmt.Errorf("failed reading '%s' which has been set as the TEST_KUBECONFIG: %v", cfg.Kubeconfig.Path, err)
	}
	log.Printf("Using a set TEST_KUBECONFIG of '%s' for Origin API calls.", cfg.Kubeconfig.Path)
	return nil
}

// cluster name format must be short enough to support all versions
func clusterName() string {
	vers := strings.TrimPrefix(state.Instance.Cluster.Version, osd.VersionPrefix)
	safeVersion := strings.Replace(vers, ".", "-", -1)
	return "ci-cluster-" + safeVersion + "-" + config.Instance.Suffix
}

func writeLogs(m map[string][]byte) {
	for k, v := range m {
		name := k + "-log.txt"
		filePath := filepath.Join(config.Instance.ReportDir, name)
		err := ioutil.WriteFile(filePath, v, os.ModePerm)
		Expect(err).NotTo(HaveOccurred(), "failed to write log '%s'", filePath)
	}
}
