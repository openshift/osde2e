package e2e

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/openshift/osde2e/pkg/common/cluster"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/events"
	"github.com/openshift/osde2e/pkg/common/metadata"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/state"
	"github.com/openshift/osde2e/pkg/common/util"
)

// Check if the test should run
var _ = ginkgo.BeforeEach(func() {
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
var _ = ginkgo.JustAfterEach(getLogs)

func getLogs() {
	defer ginkgo.GinkgoRecover()
	state := state.Instance

	if provider == nil {
		log.Println("OSD was not configured. Skipping log collection...")
	} else if state.Cluster.ID == "" {
		log.Println("CLUSTER_ID is not set, likely due to a setup failure. Skipping log collection...")
	} else {
		logs, err := provider.Logs(state.Cluster.ID)
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

	provider, err := providers.ClusterProvider()

	if err != nil {
		return fmt.Errorf("error getting cluster provisioning client: %v", err)
	}

	// create a new cluster if no ID is specified
	if state.Cluster.ID == "" {
		if state.Cluster.Name == "" {
			state.Cluster.Name = clusterName()
		}

		if state.Cluster.ID, err = provider.LaunchCluster(); err != nil {
			return fmt.Errorf("could not launch cluster: %v", err)
		}
	} else {
		log.Printf("CLUSTER_ID of '%s' was provided, skipping cluster creation and using it instead", state.Cluster.ID)

		cluster, err := provider.GetCluster(state.Cluster.ID)
		if err != nil {
			return fmt.Errorf("could not retrieve cluster information from OCM: %v", err)
		}

		state.Cluster.Name = cluster.Name()
		log.Printf("CLUSTER_NAME set to %s from OCM.", state.Cluster.Name)

		state.Cluster.Version = cluster.Version()
		log.Printf("CLUSTER_VERSION set to %s from OCM.", state.Cluster.Version)

		state.CloudProvider.CloudProviderID = cluster.CloudProvider()
		log.Printf("CLOUD_PROVIDER_ID set to %s from OCM.", state.CloudProvider.CloudProviderID)

		state.CloudProvider.Region = cluster.Region()
		log.Printf("CLOUD_PROVIDER_REGION set to %s from OCM.", state.CloudProvider.Region)

		log.Printf("Found addons: %s", strings.Join(cluster.Addons(), ","))
	}

	metadata.Instance.SetClusterName(state.Cluster.Name)
	metadata.Instance.SetClusterID(state.Cluster.ID)

	if err = cluster.WaitForClusterReady(provider, state.Cluster.ID); err != nil {
		return fmt.Errorf("failed waiting for cluster ready: %v", err)
	}

	if state.Kubeconfig.Contents, err = provider.ClusterKubeconfig(state.Cluster.ID); err != nil {
		return fmt.Errorf("could not get kubeconfig for cluster: %v", err)
	}

	return nil
}

// installAddons installs addons onto the cluster
func installAddons() (err error) {
	clusterID := state.Instance.Cluster.ID
	num, err := provider.InstallAddons(clusterID, config.Instance.Addons.IDs)
	if err != nil {
		return fmt.Errorf("could not install addons: %s", err.Error())
	}
	if num > 0 {
		if err = cluster.WaitForClusterReady(provider, clusterID); err != nil {
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
	vers := strings.TrimPrefix(state.Instance.Cluster.Version, util.VersionPrefix)
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
