package osde2e

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/onsi/ginkgo"

	"github.com/openshift/osde2e/pkg/cluster"
	"github.com/openshift/osde2e/pkg/config"
)

func init() {
	rand.Seed(time.Now().Unix())
}

const (
	DefaultVersion = "openshift-v4.0-beta4"
)

// Setup cluster before testing begins.
var _ = ginkgo.SynchronizedBeforeSuite(func() []byte {
	cfg := config.Cfg
	if cfg.ClusterVersion == "" {
		cfg.ClusterVersion = DefaultVersion
	}

	if cfg.Suffix == "" {
		cfg.Suffix = randomStr(5)
	}

	if cfg.ClusterName == "" {
		safeVersion := strings.Replace(cfg.ClusterVersion, ".", "-", -1)
		cfg.ClusterName = "ci-cluster-" + safeVersion + "-" + cfg.Suffix
	}

	if cfg.ReportDir == "" {
		if dir, err := ioutil.TempDir("", "osde2e"); err == nil {
			cfg.ReportDir = dir
		}
	}

	if err := setupCluster(cfg); err != nil {
		msg := fmt.Sprintf("Failed to setup cluster for testing: %v", err)
		ginkgo.Fail(msg)
	}
	return []byte{}
}, func(data []byte) {
	// only needs to run once
})

// Destroy cluster after testing.
var _ = ginkgo.SynchronizedAfterSuite(func() {
	// only run on one
}, func() {
	cfg := config.Cfg
	if cfg.NoDestroy {
		log.Println("NO_DESTROY is set, skipping deleting cluster.")
		return
	}

	if err := UHC.DeleteCluster(cfg.ClusterId); err != nil {
		ginkgo.Fail("failed to destroy cluster")
	}
})

// setupCluster brings up a cluster, waits for it to be ready, then returns it's name.
func setupCluster(cfg *config.Config) (err error) {
	if UHC, err = cluster.NewUHC(cfg.UHCToken, !cfg.UseProd); err != nil {
		return fmt.Errorf("could not setup UHC: %v", err)
	}

	// create a new cluster if no ID is specified
	if cfg.ClusterId == "" {
		if cfg.ClusterId, err = UHC.LaunchCluster(cfg.ClusterName, cfg.ClusterVersion, cfg.AWSKeyId, cfg.AWSAccessKey); err != nil {
			return fmt.Errorf("could not launch cluster: %v", err)
		}
	} else {
		log.Printf("CLUSTER_ID of '%s' was provided, skipping cluster creation and using it instead", cfg.ClusterId)
	}

	if err = UHC.WaitForClusterReady(cfg.ClusterId); err != nil {
		return fmt.Errorf("failed waiting for cluster ready: %v", err)
	}

	if len(cfg.Kubeconfig) == 0 {
		if cfg.Kubeconfig, err = UHC.ClusterKubeconfig(cfg.ClusterId); err != nil {
			return fmt.Errorf("could not get kubeconfig for cluster: %v", err)
		}
	} else {
		filename := string(cfg.Kubeconfig)
		cfg.Kubeconfig, err = ioutil.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("failed reading '%s' which has been set as the TEST_KUBECONFIG: %v", filename, err)
		}
		log.Printf("Using a set TEST_KUBECONFIG of '%s' for Origin API calls.", filename)
	}
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
