package osde2e

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/onsi/ginkgo"

	"github.com/openshift/osde2e/pkg/cluster"
	"github.com/openshift/osde2e/pkg/verify"
)

func init() {
	rand.Seed(time.Now().Unix())
}

const (
	DefaultVersion = "openshift-v4.0-beta4"
)

// Setup cluster before testing begins.
var _ = ginkgo.SynchronizedBeforeSuite(func() []byte {
	if Cfg.ClusterVersion == "" {
		Cfg.ClusterVersion = DefaultVersion
	}

	if Cfg.Suffix == "" {
		Cfg.Suffix = randomStr(5)
	}

	if Cfg.ClusterName == "" {
		safeVersion := strings.Replace(Cfg.ClusterVersion, ".", "-", -1)
		Cfg.ClusterName = "ci-cluster-" + safeVersion + "-" + Cfg.Suffix
	}

	if Cfg.ReportDir == "" {
		if dir, err := ioutil.TempDir("", "osde2e"); err == nil {
			Cfg.ReportDir = dir
		}
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
	if err := UHC.DeleteCluster(ClusterId); err != nil {
		ginkgo.Fail("failed to destroy cluster")
	}
})

// setupCluster brings up a cluster, waits for it to be ready, then returns it's name.
func setupCluster() (err error) {
	if UHC, err = cluster.NewUHC(Cfg.UHCToken, !Cfg.UseProd); err != nil {
		return fmt.Errorf("could not setup UHC: %v", err)
	}

	if ClusterId, err = UHC.LaunchCluster(Cfg.ClusterName, Cfg.ClusterVersion, Cfg.AWSKeyId, Cfg.AWSAccessKey); err != nil {
		return fmt.Errorf("could not launch cluster: %v", err)
	}

	if err = UHC.WaitForClusterReady(ClusterId); err != nil {
		return fmt.Errorf("failed waiting for cluster ready: %v", err)
	}

	if Kubeconfig, err = UHC.ClusterKubeconfig(ClusterId); err != nil {
		return fmt.Errorf("could not get kubeconfig for cluster: %v", err)
	}

	if err = os.Setenv(verify.TestKubeconfigEnv, string(Kubeconfig)); err != nil {
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
