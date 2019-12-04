package common

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/openshift/osde2e/pkg/config"
	"github.com/openshift/osde2e/pkg/metadata"
	"github.com/openshift/osde2e/pkg/osd"
)

func init() {
	rand.Seed(time.Now().Unix())
}

// Setup cluster before testing begins.
var _ = ginkgo.SynchronizedBeforeSuite(func() []byte {
	defer ginkgo.GinkgoRecover()
	cfg := config.Cfg

	err := setupCluster(cfg)
	Expect(err).ShouldNot(HaveOccurred(), "failed to setup cluster for testing")

	if len(cfg.Kubeconfig) == 0 {
		// Give the cluster some breathing room.
		log.Println("OSD cluster installed. Sleeping for 600s.")
		time.Sleep(600 * time.Second)
	}

	return []byte{}
}, func(data []byte) {
	// only needs to run once
})

// Destroy cluster after testing.
var _ = ginkgo.AfterSuite(func() {
	defer ginkgo.GinkgoRecover()
	cfg := config.Cfg

	if OSD == nil {
		log.Println("OSD was not configured. Skipping AfterSuite...")
	} else if cfg.ClusterID == "" {
		log.Println("CLUSTER_ID is not set, likely due to a setup failure. Skipping AfterSuite...")
	} else {
		log.Printf("Getting logs for cluster '%s'...", cfg.ClusterID)

		logs, err := OSD.FullLogs(cfg.ClusterID)
		Expect(err).NotTo(HaveOccurred(), "failed to collect cluster logs")
		writeLogs(cfg, logs)
	}
})

// setupCluster brings up a cluster, waits for it to be ready, then returns it's name.
func setupCluster(cfg *config.Config) (err error) {
	// if TEST_KUBECONFIG has been set, skip configuring UHC
	if len(cfg.Kubeconfig) > 0 {
		return useKubeconfig(cfg)
	}

	// create a new cluster if no ID is specified
	if cfg.ClusterID == "" {
		if cfg.ClusterName == "" {
			cfg.ClusterName = clusterName(cfg)
		}

		if cfg.ClusterID, err = OSD.LaunchCluster(cfg); err != nil {
			return fmt.Errorf("could not launch cluster: %v", err)
		}
	} else {
		log.Printf("CLUSTER_ID of '%s' was provided, skipping cluster creation and using it instead", cfg.ClusterID)

		if cfg.ClusterName == "" {
			cluster, err := OSD.GetCluster(cfg.ClusterID)
			if err != nil {
				return fmt.Errorf("could not retrieve cluster information from OCM: %v", err)
			}

			if cluster.Name() == "" {
				return fmt.Errorf("cluster name from OCM is empty, and this shouldn't be possible")
			}

			cfg.ClusterName = cluster.Name()
			log.Printf("CLUSTER_NAME not provided, retrieved %s from OCM.", cfg.ClusterName)
		}
	}

	metadata.Instance.ClusterName = cfg.ClusterName
	metadata.Instance.ClusterID = cfg.ClusterID

	if err = OSD.WaitForClusterReady(cfg); err != nil {
		return fmt.Errorf("failed waiting for cluster ready: %v", err)
	}

	if cfg.Kubeconfig, err = OSD.ClusterKubeconfig(cfg.ClusterID); err != nil {
		return fmt.Errorf("could not get kubeconfig for cluster: %v", err)
	}
	return nil
}

// useKubeconfig reads the path provided for a TEST_KUBECONFIG and uses it for testing.
func useKubeconfig(cfg *config.Config) (err error) {
	filename := string(cfg.Kubeconfig)

	_, err = clientcmd.RESTConfigFromKubeConfig(cfg.Kubeconfig)
	if err != nil {
		log.Println("Not an existing Kubeconfig, attempting to read file instead...")
	} else {
		log.Println("Existing valid kubeconfig!")
		return nil
	}

	cfg.Kubeconfig, err = ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed reading '%s' which has been set as the TEST_KUBECONFIG: %v", filename, err)
	}
	log.Printf("Using a set TEST_KUBECONFIG of '%s' for Origin API calls.", filename)
	return nil
}

// cluster name format must be short enough to support all versions
func clusterName(cfg *config.Config) string {
	vers := strings.TrimPrefix(cfg.ClusterVersion, osd.VersionPrefix)
	safeVersion := strings.Replace(vers, ".", "-", -1)
	return "ci-cluster-" + safeVersion + "-" + cfg.Suffix
}

func randomStr(length int) (str string) {
	chars := "0123456789abcdefghijklmnopqrstuvwxyz"
	for i := 0; i < length; i++ {
		c := string(chars[rand.Intn(len(chars))])
		str += c
	}
	return
}

func writeLogs(cfg *config.Config, m map[string][]byte) {
	for k, v := range m {
		name := k + "-log.txt"
		filePath := filepath.Join(cfg.ReportDir, name)
		err := ioutil.WriteFile(filePath, v, os.ModePerm)
		Expect(err).NotTo(HaveOccurred(), "failed to write log '%s'", filePath)
	}
}
