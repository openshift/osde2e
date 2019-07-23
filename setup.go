package osde2e

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

	"github.com/openshift/osde2e/pkg/config"
	"github.com/openshift/osde2e/pkg/upgrade"
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

	// upgrade cluster if requested
	if cfg.UpgradeImage != "" || cfg.UpgradeReleaseStream != "" {
		err = upgrade.RunUpgrade(cfg)
		Expect(err).ShouldNot(HaveOccurred(), "failed performing upgrade")
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

		logs, err := OSD.Logs(cfg.ClusterID, 200)
		Expect(err).NotTo(HaveOccurred(), "failed to collect cluster logs")
		writeLogs(cfg, logs)

		if cfg.NoDestroy {
			log.Println("NO_DESTROY is set, skipping deleting cluster.")
			return
		}

		log.Printf("Destroying cluster '%s'...", cfg.ClusterID)
		err = OSD.DeleteCluster(cfg.ClusterID)
		Expect(err).NotTo(HaveOccurred(), "failed to destroy cluster")
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
			safeVersion := strings.Replace(cfg.ClusterVersion, ".", "-", -1)
			cfg.ClusterName = "ci-cluster-" + safeVersion + "-" + cfg.Suffix
		}

		if cfg.ClusterID, err = OSD.LaunchCluster(cfg); err != nil {
			return fmt.Errorf("could not launch cluster: %v", err)
		}
	} else {
		log.Printf("CLUSTER_ID of '%s' was provided, skipping cluster creation and using it instead", cfg.ClusterID)
	}

	if err = OSD.WaitForClusterReady(cfg.ClusterID, cfg.ClusterUpTimeout); err != nil {
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
	cfg.Kubeconfig, err = ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed reading '%s' which has been set as the TEST_KUBECONFIG: %v", filename, err)
	}
	log.Printf("Using a set TEST_KUBECONFIG of '%s' for Origin API calls.", filename)
	return nil
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
