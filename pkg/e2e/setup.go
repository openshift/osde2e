package e2e

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/openshift/osde2e/pkg/common/cluster"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/events"
	"github.com/openshift/osde2e/pkg/common/metadata"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/util"
)

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

// Setup cluster before testing begins.
var _ = ginkgo.SynchronizedBeforeSuite(func() []byte {
	defer ginkgo.GinkgoRecover()
	err := setupCluster()
	events.HandleErrorWithEvents(err, events.InstallSuccessful, events.InstallFailed).ShouldNot(HaveOccurred(), "failed to setup cluster for testing")
	if err != nil {
		return []byte{}
	}

	if len(viper.GetString(config.Addons.IDs)) > 0 {
		err = installAddons()
		events.HandleErrorWithEvents(err, events.InstallAddonsSuccessful, events.InstallAddonsFailed).ShouldNot(HaveOccurred(), "failed while installing addons")
		if err != nil {
			return []byte{}
		}
	}

	if len(viper.GetString(config.Kubeconfig.Contents)) == 0 {
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
// TODO: SDA-2594 Hotfix
/*
var _ = ginkgo.JustAfterEach(getLogs)

func getLogs() {
	defer ginkgo.GinkgoRecover()

	clusterID := viper.GetString(config.Cluster.ID)
	if provider == nil {
		log.Println("OSD was not configured. Skipping log collection...")
	} else if clusterID == "" {
		log.Println("CLUSTER_ID is not set, likely due to a setup failure. Skipping log collection...")
	} else {
		logs, err := provider.Logs(clusterID)
		Expect(err).NotTo(HaveOccurred(), "failed to collect cluster logs")
		writeLogs(logs)
	}
}
func writeLogs(m map[string][]byte) {
	for k, v := range m {
		name := k + "-log.txt"
		filePath := filepath.Join(viper.GetString(config.ReportDir), name)
		err := ioutil.WriteFile(filePath, v, os.ModePerm)
		Expect(err).NotTo(HaveOccurred(), "failed to write log '%s'", filePath)
	}
}
*/

// setupCluster brings up a cluster, waits for it to be ready, then returns it's name.
func setupCluster() (err error) {
	// if TEST_KUBECONFIG has been set, skip configuring OCM
	if len(viper.GetString(config.Kubeconfig.Contents)) > 0 || len(viper.GetString(config.Kubeconfig.Path)) > 0 {
		return useKubeconfig()
	}

	provider, err := providers.ClusterProvider()

	if err != nil {
		return fmt.Errorf("error getting cluster provisioning client: %v", err)
	}

	// create a new cluster if no ID is specified
	clusterID := viper.GetString(config.Cluster.ID)
	if clusterID == "" {
		if viper.GetString(config.Cluster.Name) == "" {
			viper.Set(config.Cluster.Name, clusterName())
		}

		if clusterID, err = provider.LaunchCluster(); err != nil {
			return fmt.Errorf("could not launch cluster: %v", err)
		}
		viper.Set(config.Cluster.ID, clusterID)

	} else {
		log.Printf("CLUSTER_ID of '%s' was provided, skipping cluster creation and using it instead", clusterID)

		cluster, err := provider.GetCluster(clusterID)
		if err != nil {
			return fmt.Errorf("could not retrieve cluster information from OCM: %v", err)
		}

		viper.Set(config.Cluster.Name, cluster.Name())
		log.Printf("CLUSTER_NAME set to %s from OCM.", viper.GetString(config.Cluster.Name))

		viper.Set(config.Cluster.Version, cluster.Version())
		log.Printf("CLUSTER_VERSION set to %s from OCM.", viper.GetString(config.Cluster.Version))

		viper.Set(config.CloudProvider.CloudProviderID, cluster.CloudProvider())
		log.Printf("CLOUD_PROVIDER_ID set to %s from OCM.", viper.GetString(config.CloudProvider.CloudProviderID))

		viper.Set(config.CloudProvider.Region, cluster.Region())
		log.Printf("CLOUD_PROVIDER_REGION set to %s from OCM.", viper.GetString(config.CloudProvider.Region))

		log.Printf("Found addons: %s", strings.Join(cluster.Addons(), ","))
	}

	metadata.Instance.SetClusterName(viper.GetString(config.Cluster.Name))
	metadata.Instance.SetClusterID(clusterID)
	metadata.Instance.SetRegion(viper.GetString(config.CloudProvider.Region))

	if err = cluster.WaitForClusterReady(provider, clusterID); err != nil {
		return fmt.Errorf("failed waiting for cluster ready: %v", err)
	}

	var kubeconfigBytes []byte
	if kubeconfigBytes, err = provider.ClusterKubeconfig(clusterID); err != nil {
		return fmt.Errorf("could not get kubeconfig for cluster: %v", err)
	}
	viper.Set(config.Kubeconfig.Contents, string(kubeconfigBytes))

	return nil
}

// installAddons installs addons onto the cluster
func installAddons() (err error) {
	clusterID := viper.GetString(config.Cluster.ID)
	num, err := provider.InstallAddons(clusterID, strings.Split(viper.GetString(config.Addons.IDs), ","))
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
	_, err = clientcmd.RESTConfigFromKubeConfig([]byte(viper.GetString(config.Kubeconfig.Contents)))
	if err != nil {
		log.Println("Not an existing Kubeconfig, attempting to read file instead...")
	} else {
		log.Println("Existing valid kubeconfig!")
		return nil
	}

	kubeconfigPath := viper.GetString(config.Kubeconfig.Path)
	kubeconfigBytes, err := ioutil.ReadFile(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed reading '%s' which has been set as the TEST_KUBECONFIG: %v", kubeconfigPath, err)
	}
	viper.Set(config.Kubeconfig.Contents, string(kubeconfigBytes))
	log.Printf("Using a set TEST_KUBECONFIG of '%s' for Origin API calls.", kubeconfigPath)
	return nil
}

// cluster name format must be short enough to support all versions
func clusterName() string {
	vers := strings.TrimPrefix(viper.GetString(config.Cluster.Version), util.VersionPrefix)
	safeVersion := strings.Replace(vers, ".", "-", -1)
	return "ci-cluster-" + safeVersion + "-" + viper.GetString(config.Suffix)
}
