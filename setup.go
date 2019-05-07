package osde2e

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/onsi/ginkgo"

	"github.com/openshift/osde2e/pkg/config"
	"github.com/openshift/osde2e/pkg/osd"
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
	if err := setupCluster(cfg); err != nil {
		msg := fmt.Sprintf("Failed to setup cluster for testing: %v", err)
		log.Println(msg)
		ginkgo.Fail(msg)
	}
	return []byte{}
}, func(data []byte) {
	// only needs to run once
})

// Destroy cluster after testing.
var _ = ginkgo.AfterSuite(func() {
	defer ginkgo.GinkgoRecover()
	cfg := config.Cfg

	if UHC != nil {
		log.Printf("Getting logs for cluster '%s'...", cfg.ClusterId)

		logs, err := UHC.Logs(cfg.ClusterId, 200)
		if err != nil {
			msg := fmt.Sprintf("Failed to collect cluster logs: %v", err)
			log.Println(msg)
			ginkgo.Fail(msg)
		} else {
			writeLogs(cfg, logs)
		}
	}

	if cfg.NoDestroy {
		log.Println("NO_DESTROY is set, skipping deleting cluster.")
		return
	}

	if err := UHC.DeleteCluster(cfg.ClusterId); err != nil {
		msg := fmt.Sprintf("Failed to destroy cluster: %v", err)
		log.Println(msg)
		ginkgo.Fail(msg)
	}
})

// setupCluster brings up a cluster, waits for it to be ready, then returns it's name.
func setupCluster(cfg *config.Config) (err error) {
	if UHC, err = osd.NewUHC(cfg.UHCToken, !cfg.UseProd, cfg.DebugUHC); err != nil {
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

func writeLogs(cfg *config.Config, m map[string][]byte) {
	for k, v := range m {
		name := k + "-log.txt"
		filePath := filepath.Join(cfg.ReportDir, name)
		if err := ioutil.WriteFile(filePath, v, os.ModePerm); err != nil {
			msg := fmt.Sprintf("Failed to write log '%s': %v", filePath, err)
			log.Println(msg)
			ginkgo.Fail(msg)
		}
	}
}
