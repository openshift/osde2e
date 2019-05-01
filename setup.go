package osde2e

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/onsi/ginkgo"

	"github.com/openshift/osde2e/pkg/cluster"
	"github.com/openshift/osde2e/pkg/verify"
)

func init() {
	rand.Seed(time.Now().Unix())
}

// Setup cluster before testing begins.
var _ = ginkgo.SynchronizedBeforeSuite(func() []byte {
	setupCfgFromEnv()

	if Cfg.Prefix == "" {
		Cfg.Prefix = randomStr(5)
	}

	if Cfg.ClusterName == "" {
		Cfg.ClusterName = Cfg.Prefix + "-test-cluster"
	}

	if err := setupCluster(); err != nil {
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
	if err := Cfg.uhc.DeleteCluster(Cfg.clusterId); err != nil {
		ginkgo.Fail("failed to destroy cluster")
	}
})

// setupCluster brings up a cluster, waits for it to be ready, then returns it's name.
func setupCluster() (err error) {
	if Cfg.uhc, err = cluster.NewUHC(Cfg.UHCToken, !Cfg.UseProd); err != nil {
		return fmt.Errorf("could not setup UHC: %v", err)
	}

	if Cfg.clusterId, err = Cfg.uhc.LaunchCluster(Cfg.ClusterName, Cfg.AWSKeyId, Cfg.AWSAccessKey); err != nil {
		return fmt.Errorf("could not launch cluster: %v", err)
	}

	if err = Cfg.uhc.WaitForClusterReady(Cfg.clusterId); err != nil {
		return fmt.Errorf("failed waiting for cluster ready: %v", err)
	}

	if Cfg.kubeconfig, err = Cfg.uhc.ClusterKubeconfig(Cfg.clusterId); err != nil {
		return fmt.Errorf("could not get kubeconfig for cluster: %v", err)
	}

	if err = os.Setenv(verify.TestKubeconfigEnv, string(Cfg.kubeconfig)); err != nil {
		return fmt.Errorf("could not set kubeconfig: %v", err)
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
