package proxy

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	configv1 "github.com/openshift/api/config/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/cluster"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/logging"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/util"
)

var (
	postInstallProxyTestName string = "[Suite: proxy] Post-Install Cluster Proxy"
	logger                          = logging.CreateNewStdLoggerOrUseExistingLogger(nil)
)

const (
	// How long to wait for proxy changes to be reflected in the resource
	proxyConfigSyncDuration = 15 * time.Minute
	// How long to wait for proxy changes to be applied and cluster to return to health
	proxyHealthCheckWaitDuration = 45 * time.Minute
)

func init() {
	alert.RegisterGinkgoAlert(postInstallProxyTestName, "SD-SREP", "@sd-srep-team-hulk", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(postInstallProxyTestName, func() {
	ginkgo.Context("Day 2 Configuration", func() {
		testAddProxy()
		testRemoveProxy()
	})
})

func testAddProxy() {
	util.GinkgoIt("can add a proxy to the cluster successfully", func(ctx context.Context) {
		// setup helper
		h := helper.New()

		clusterID := viper.GetString(config.Cluster.ID)
		clusterProvider, err := providers.ClusterProvider()
		Expect(err).NotTo(HaveOccurred())

		httpsProxy := viper.GetString(config.Proxy.HttpsProxy)
		httpProxy := viper.GetString(config.Proxy.HttpProxy)
		userCABundle := viper.GetString(config.Proxy.UserCABundle)
		userCABundleData, err := clusterProvider.LoadUserCaBundleData(userCABundle)
		Expect(err).NotTo(HaveOccurred())

		logger.Printf("Setting cluster-wide proxy to httpsProxy=%v,httpProxy=%v,settingCA=%v",
			httpsProxy, httpProxy, userCABundle != "")
		err = clusterProvider.AddClusterProxy(clusterID, httpsProxy, httpProxy, userCABundle)
		Expect(err).NotTo(HaveOccurred())

		// Wait to see proxy reflected on the cluster
		logger.Printf("Validating state of proxy on cluster within %v minutes", proxyConfigSyncDuration.Minutes())
		err = wait.Poll(30*time.Second, proxyConfigSyncDuration, func() (bool, error) {
			var proxy *configv1.Proxy
			var cabundle *v1.ConfigMap

			// Validate state of proxy on-cluster vs what values it should have
			proxy, err = h.Cfg().ConfigV1().Proxies().Get(ctx, "cluster", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			if userCABundle != "" {
				if proxy.Spec.TrustedCA.Name != "user-ca-bundle" {
					return false, nil
				}
				// Check if the ConfigMap exists
				cabundle, err = h.Kube().CoreV1().ConfigMaps("openshift-config").Get(ctx, "user-ca-bundle", metav1.GetOptions{})
				// don't treat the cabundle not existing as an error
				if err != nil && !apierrors.IsNotFound(err) {
					return false, err
				}
				cabundleData, found := cabundle.Data["ca-bundle.crt"]
				// if the configmap exists, so should this key, and the value should match
				if !found || strings.TrimSpace(cabundleData) != strings.TrimSpace(userCABundleData) {
					logger.Printf("User CA Bundle still not reflected on cluster")
					return false, nil
				}
			}
			if httpsProxy != "" {
				// Check if status reflects the HTTPS Proxy value
				if proxy.Status.HTTPSProxy != httpsProxy {
					logger.Printf("HTTPS Proxy still not reflected on cluster")
					return false, nil
				}
			}
			if httpProxy != "" {
				// Check if status reflects the HTTPS Proxy value
				if proxy.Status.HTTPProxy != httpProxy {
					logger.Printf("HTTP Proxy still not reflected on cluster")
					return false, nil
				}
			}
			return true, nil
		})
		Expect(err).NotTo(HaveOccurred())

		// The cluster's proxy and configmap state reflects what we expect
		// So now, is the cluster still healthy?
		logger.Printf("Verifying cluster health after proxy addition..")
		err = wait.PollImmediate(30*time.Second, proxyHealthCheckWaitDuration, func() (bool, error) {
			isHealthy, failures, _ := cluster.PollClusterHealth(clusterID, logger)
			if isHealthy {
				logger.Printf("cluster is healthy after proxy addition\n")
				return true, nil
			}
			log.Printf("cluster is not healthy after proxy addition\n")
			if len(failures) > 0 {
				logger.Printf("Currently failing %s health checks", strings.Join(failures, ", "))
			}
			return false, nil
		})
		Expect(err).NotTo(HaveOccurred())
	}, proxyConfigSyncDuration.Seconds()+proxyHealthCheckWaitDuration.Seconds())
}

func testRemoveProxy() {
	util.GinkgoIt("can remove proxy from the cluster successfully", func(ctx context.Context) {
		// setup helper
		h := helper.New()

		clusterID := viper.GetString(config.Cluster.ID)
		clusterProvider, err := providers.ClusterProvider()
		Expect(err).NotTo(HaveOccurred())

		err = clusterProvider.RemoveClusterProxy(clusterID)
		Expect(err).NotTo(HaveOccurred())

		// Wait to see proxy reflected on the cluster
		logger.Printf("Validating state of proxy on cluster within %v minutes", proxyConfigSyncDuration.Minutes())
		err = wait.Poll(30*time.Second, proxyConfigSyncDuration, func() (bool, error) {
			var proxy *configv1.Proxy

			// Validate state of proxy on-cluster vs what values it should have
			proxy, err = h.Cfg().ConfigV1().Proxies().Get(ctx, "cluster", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			if proxy.Spec.HTTPProxy != "" || proxy.Status.HTTPProxy != "" {
				return false, nil
			}
			if proxy.Spec.HTTPSProxy != "" || proxy.Status.HTTPSProxy != "" {
				return false, nil
			}

			_, err = h.Kube().CoreV1().ConfigMaps("openshift-config").Get(ctx, "user-ca-bundle", metav1.GetOptions{})
			if !apierrors.IsNotFound(err) {
				return false, nil
			}
			if proxy.Spec.TrustedCA.Name != "" {
				return false, nil
			}

			return true, nil
		})
		Expect(err).NotTo(HaveOccurred())
		// The cluster's proxy and configmap state reflects what we expect
		// So now, is the cluster still healthy?
		logger.Printf("Verifying cluster health after proxy removed..")
		err = wait.PollImmediate(30*time.Second, proxyHealthCheckWaitDuration, func() (bool, error) {
			isHealthy, failures, _ := cluster.PollClusterHealth(clusterID, logger)
			if isHealthy {
				logger.Printf("cluster is healthy after proxy removed\n")
				return true, nil
			}
			log.Printf("cluster is not healthy after proxy removed\n")
			if len(failures) > 0 {
				logger.Printf("Currently failing %s health checks", strings.Join(failures, ", "))
			}
			return false, nil
		})
		Expect(err).NotTo(HaveOccurred())
	}, proxyConfigSyncDuration.Seconds()+proxyHealthCheckWaitDuration.Seconds())
}
