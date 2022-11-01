package cluster

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/hashicorp/go-multierror"
	osconfig "github.com/openshift/client-go/config/clientset/versioned"
	"github.com/openshift/osde2e/pkg/common/cluster/healthchecks"
	"github.com/openshift/osde2e/pkg/common/clusterproperties"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/logging"
	"github.com/openshift/osde2e/pkg/common/metadata"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/util"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// errorWindow is the number of checks made to determine if a cluster has truly failed.
	errorWindow = 20
	// pendingPodThreshold is the maximum number of times a pod is allowed to be in pending state before erroring out in PollClusterHealth.
	pendingPodThreshold = 10
)

// podErrorTracker is the data structure that keeps track of pending state counters for each pod against their pod UIDs.
var podErrorTracker healthchecks.PodErrorTracker

// GetClusterVersion will get the current cluster version for the cluster.
func GetClusterVersion(provider spi.Provider, clusterID string) (*semver.Version, error) {
	restConfig, err := getRestConfig(provider, clusterID)
	if err != nil {
		return nil, fmt.Errorf("error getting rest config: %v", err)
	}

	oscfg, err := osconfig.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("error generating OpenShift Clientset: %v", err)
	}

	cvo, err := healthchecks.GetClusterVersionObject(oscfg.ConfigV1())
	if err != nil {
		return nil, fmt.Errorf("error getting cluster version object: %v", err)
	}

	version, err := semver.NewVersion(cvo.Status.Desired.Version)
	if err != nil {
		return nil, fmt.Errorf("error parsing version from server: %v", err)
	}

	return version, err
}

// ScaleCluster will scale the cluster up to the provided size.
func ScaleCluster(clusterID string, numComputeNodes int) error {
	provider, err := providers.ClusterProvider()

	if err != nil {
		return fmt.Errorf("error getting cluster provisioning client: %v", err)
	}

	err = provider.ScaleCluster(clusterID, numComputeNodes)
	if err != nil {
		return fmt.Errorf("error trying to scale cluster: %v", err)
	}

	podErrorTracker.NewPodErrorTracker(pendingPodThreshold)
	return waitForClusterReadyWithOverrideAndExpectedNumberOfNodes(clusterID, nil, false, true)
}

// WaitForClusterReadyPostInstall blocks until the cluster is ready for testing using mechanisms appropriate
// for a newly-installed cluster.
func WaitForClusterReadyPostInstall(clusterID string, logger *log.Logger) error {
	logger = logging.CreateNewStdLoggerOrUseExistingLogger(logger)
	podErrorTracker.NewPodErrorTracker(pendingPodThreshold)
	provider, err := providers.ClusterProvider()
	if err != nil {
		return fmt.Errorf("error getting cluster provisioning client: %v", err)
	}

	installTimeout := viper.GetInt64(config.Cluster.InstallTimeout)
	logger.Printf("Waiting %v minutes for cluster '%s' to be ready...\n", installTimeout, clusterID)

	_, err = waitForOCMProvisioning(provider, clusterID, installTimeout, logger, false)
	if err != nil {
		return fmt.Errorf("OCM never became ready: %w", err)
	}
	logger.Println("Cluster is provisioned in OCM")

	clusterConfig, _, err := ClusterConfig(clusterID)
	if err != nil {
		return fmt.Errorf("failed looking up cluster config for healthcheck: %w", err)
	}

	kubeClient, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return fmt.Errorf("error generating Kube Clientset: %w", err)
	}

	duration, err := time.ParseDuration(viper.GetString(config.Tests.ClusterHealthChecksTimeout))
	if err != nil {
		return fmt.Errorf("failed parsing health check timeout: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()
	err = healthchecks.CheckHealthcheckJob(kubeClient, ctx, nil)
	if err != nil {
		return fmt.Errorf("cluster failed health check: %w", err)
	}

	cluster, err := provider.GetCluster(clusterID)
	if err != nil {
		return fmt.Errorf("failed getting cluster from provider: %w", err)
	}

	if err := provider.AddProperty(cluster, clusterproperties.Status, clusterproperties.StatusHealthy); err != nil {
		return fmt.Errorf("error trying to add healthy property to cluster ID %s: %w", cluster.ID(), err)
	}
	return nil
}

// WaitForClusterReadyPostUpgrade blocks until the cluster is ready for testing using healthcheck mechanisms appropriate
// for after a cluster version upgrade.
func WaitForClusterReadyPostUpgrade(clusterID string, logger *log.Logger) error {
	podErrorTracker.NewPodErrorTracker(pendingPodThreshold)
	return waitForClusterReadyWithOverrideAndExpectedNumberOfNodes(clusterID, logger, true, false)
}

// WaitForClusterReadyPostScale blocks until the cluster is ready for testing and uses healthcheck mechanisms appropriate
// for after the cluster has been scaled.
func WaitForClusterReadyPostScale(clusterID string, logger *log.Logger) error {
	podErrorTracker.NewPodErrorTracker(pendingPodThreshold)
	return waitForClusterReadyWithOverrideAndExpectedNumberOfNodes(clusterID, logger, false, false)
}

// WaitForClusterReadyPostWake blocks until the cluster is ready for testing, deletes errored pods, and then uses
// healthcheck mechanisms appropriate for after the cluster resumed from hibernation.
func WaitForClusterReadyPostWake(clusterID string, logger *log.Logger) error {
	log.Printf("Cluster %s just woke up, waiting for 10 minutes...", clusterID)
	provider, err := providers.ClusterProvider()
	if err != nil {
		return fmt.Errorf("error getting cluster provider: %s", err.Error())
	}
	cluster, err := provider.GetCluster(clusterID)
	if err != nil {
		return fmt.Errorf("error getting cluster from provider: %s", err.Error())
	}
	provider.AddProperty(cluster, clusterproperties.Status, clusterproperties.StatusHealthCheck)
	time.Sleep(10 * time.Minute)

	restConfig, _, err := ClusterConfig(clusterID)
	if err != nil {
		return fmt.Errorf("error getting cluster config: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("error generating Kube Clientset: %s", err.Error())
	}

	var continueToken string
	nextPods := func() (*corev1.PodList, error) {
		return kubeClient.CoreV1().Pods("").List(context.TODO(), v1.ListOptions{Continue: continueToken})
	}
	for list, err := nextPods(); len(list.Items) > 0; list, err = nextPods() {
		if err != nil {
			return fmt.Errorf("error retrieving pod list: %s", err.Error())
		}
		for _, pod := range list.Items {
			if pod.Status.Phase == corev1.PodFailed || pod.Status.Phase == corev1.PodPending {
				log.Printf("Cleaning up stale pod: %s", pod.Name)
				if len(pod.Finalizers) > 0 {
					log.Printf("Removing finalizers from %s", pod.Name)
					pod.Finalizers = []string{}
					kubeClient.CoreV1().Pods(pod.Namespace).Update(context.TODO(), &pod, v1.UpdateOptions{})
				}
				log.Printf("Deleting pod %s", pod.Name)
				err = kubeClient.CoreV1().Pods(pod.Namespace).Delete(context.TODO(), pod.Name, v1.DeleteOptions{})
				if err != nil {
					log.Printf("Error deleting stale pod: %s", err.Error())
				}
			}
			if len(pod.OwnerReferences) > 0 && pod.OwnerReferences[0].Kind == "Job" {
				err = kubeClient.BatchV1().Jobs(pod.Namespace).Delete(context.TODO(), pod.OwnerReferences[0].Name, v1.DeleteOptions{})
				if err != nil {
					log.Printf("Error deleting stale job: %s", err.Error())
				}
			}
		}
		if list.Continue == "" {
			break
		}
		continueToken = list.Continue
	}

	podErrorTracker.NewPodErrorTracker(pendingPodThreshold)
	viper.Set(config.Cluster.CleanCheckRuns, 5)
	return waitForClusterReadyWithOverrideAndExpectedNumberOfNodes(clusterID, logger, false, false)
}

func waitForOCMProvisioning(provider spi.Provider, clusterID string, installTimeout int64, logger *log.Logger, isUpgrade bool) (becameReadyAt time.Time, err error) {
	logger = logging.CreateNewStdLoggerOrUseExistingLogger(logger)
	readinessSet := false
	var readinessStarted time.Time
	clusterStarted := time.Now()

	healthcheckStatus := clusterproperties.StatusHealthCheck
	if isUpgrade {
		healthcheckStatus = clusterproperties.StatusUpgradeHealthCheck
	}

	return readinessStarted, wait.PollImmediate(30*time.Second, time.Duration(installTimeout)*time.Minute, func() (bool, error) {
		cluster, err := provider.GetCluster(clusterID)
		if err != nil {
			logger.Printf("Error fetching cluster details from provider: %s", err)
			return false, nil
		}

		metadata.Instance.IncrementHealthcheckIteration()
		properties := cluster.Properties()
		currentStatus := properties[clusterproperties.Status]

		if currentStatus == clusterproperties.StatusProvisioning && !readinessSet {
			err = provider.AddProperty(cluster, clusterproperties.Status, clusterproperties.StatusWaitingForReady)
			if err != nil {
				logger.Printf("Error adding property to cluster: %s", err.Error())
				return false, nil
			}
		}

		if cluster.State() == spi.ClusterStateReady {
			if metadata.Instance.TimeToOCMReportingInstalled == 0 {
				metadata.Instance.SetTimeToOCMReportingInstalled(time.Since(clusterStarted).Seconds())
			}

			if err := provider.AddProperty(cluster, clusterproperties.Status, healthcheckStatus); err != nil {
				logger.Printf("error trying to add health-check property to cluster ID %s: %v", cluster.ID(), err)
				return false, nil
			}

			readinessStarted = time.Now()
			return true, nil
		} else if cluster.State() == spi.ClusterStateError {
			log.Print("cluster is in error state, check AWS for more details")
		}
		logger.Printf("cluster is not ready, state is: %v", cluster.State())
		return false, nil
	})
}

func waitForClusterReadyWithOverrideAndExpectedNumberOfNodes(clusterID string, logger *log.Logger, isUpgrade, overrideSkipCheck bool) error {
	logger = logging.CreateNewStdLoggerOrUseExistingLogger(logger)
	if viper.GetBool(config.Tests.SkipClusterHealthChecks) && !overrideSkipCheck {
		logger.Println("Skipping health checks...")
		return nil
	}

	provider, err := providers.ClusterProvider()

	if err != nil {
		return fmt.Errorf("error getting cluster provisioning client: %v", err)
	}

	unhealthyStatus := clusterproperties.StatusUnhealthy
	healthyStatus := clusterproperties.StatusHealthy
	if isUpgrade {
		unhealthyStatus = clusterproperties.StatusUpgradeUnhealthy
		healthyStatus = clusterproperties.StatusUpgradeHealthy
	}

	installTimeout := viper.GetInt64(config.Cluster.InstallTimeout)
	logger.Printf("Waiting %v minutes for cluster '%s' to be ready...\n", installTimeout, clusterID)
	cleanRunsNeeded := viper.GetInt(config.Cluster.CleanCheckRuns)
	cleanRuns := 0
	errRuns := 0

	readinessStarted, err := waitForOCMProvisioning(provider, clusterID, installTimeout, logger, isUpgrade)
	if err != nil {
		return fmt.Errorf("OCM never became ready: %w", err)
	}

	cluster, err := provider.GetCluster(clusterID)
	if err != nil {
		return fmt.Errorf("Error fetching cluster details from provider: %w", err)
	}

	if pollErr := wait.PollImmediate(30*time.Second, time.Duration(installTimeout)*time.Minute, func() (bool, error) {
		if cluster.State() != spi.ClusterStateReady {
			logger.Printf("Cluster is not ready, current status '%s'.", cluster.State())
			return false, nil
		}

		metadata.Instance.IncrementHealthcheckIteration()
		properties := cluster.Properties()
		currentStatus := properties[clusterproperties.Status]

		if success, failures, err := PollClusterHealth(clusterID, logger); success {
			cleanRuns++
			logger.Printf("Clean run %d/%d...", cleanRuns, cleanRunsNeeded)
			errRuns = 0
			if cleanRuns == cleanRunsNeeded {
				return true, nil
			}
			return false, nil
		} else {
			if err != nil {
				errRuns++
				logger.Printf("Error in PollClusterHealth: %v", err)
				if errRuns >= errorWindow {
					if err := provider.AddProperty(cluster, clusterproperties.Status, unhealthyStatus); err != nil {
						log.Printf("error trying to add unhealthy property to cluster ID %s: %v", clusterID, err)
					}
					return false, nil
				}
			}
			cleanRuns = 0

			failureString := strings.Join(failures, ",")
			if currentStatus != failureString {
				err = provider.AddProperty(cluster, clusterproperties.Status, failureString)
				if err != nil {
					log.Printf("error trying to add property to cluster ID %s: %v", clusterID, err)
				}
			}

			return false, nil
		}
	}); pollErr != nil {
		return fmt.Errorf("failed polling for cluster health: %w", err)
	}
	// polling succeeded and the cluster is healthy
	if metadata.Instance.TimeToClusterReady == 0 {
		metadata.Instance.SetTimeToClusterReady(time.Since(readinessStarted).Seconds())
	} else {
		metadata.Instance.SetTimeToUpgradedClusterReady(time.Since(readinessStarted).Seconds())
	}

	if err := provider.AddProperty(cluster, clusterproperties.Status, healthyStatus); err != nil {
		return fmt.Errorf("error trying to add healthy property to cluster ID %s: %w", cluster.ID(), err)
	}
	return nil
}

// ClusterConfig returns the rest API config for a given cluster as well as the provider it
// inferred to discover the config.
// param clusterID: If specified, Provider will be discovered through OCM. If the empty string,
// assume we are running in a cluster and use in-cluster REST config instead.
func ClusterConfig(clusterID string) (restConfig *rest.Config, providerType string, err error) {
	if clusterID == "" {
		if restConfig, err = rest.InClusterConfig(); err != nil {
			return nil, "", fmt.Errorf("error getting in-cluster rest config: %w", err)
		}

		// FIXME: Is there a way to discover this from within the cluster?
		// For now, ocm and rosa behave the same, so hardcode either.
		providerType = "ocm"
		return

	}
	provider, err := providers.ClusterProvider()

	if err != nil {
		return nil, "", fmt.Errorf("error getting cluster provisioning client: %w", err)
	}
	providerType = provider.Type()

	restConfig, err = getRestConfig(provider, clusterID)
	if err != nil {

		return nil, "", fmt.Errorf("error generating rest config: %w", err)
	}

	return
}

// PollClusterHealth looks at CVO data to determine if a cluster is alive/healthy or not
// param clusterID: If specified, Provider will be discovered through OCM. If the empty string,
// assume we are running in a cluster and use in-cluster REST config instead.
func PollClusterHealth(clusterID string, logger *log.Logger) (status bool, failures []string, err error) {
	logger = logging.CreateNewStdLoggerOrUseExistingLogger(logger)

	logger.Print("Polling Cluster Health...\n")

	restConfig, providerType, err := ClusterConfig(clusterID)
	if err != nil {
		logger.Printf("Error getting cluster config: %v\n", err)
		return false, nil, nil
	}

	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		logger.Printf("Error generating Kube Clientset: %v\n", err)
		return false, nil, nil
	}

	oscfg, err := osconfig.NewForConfig(restConfig)
	if err != nil {
		logger.Printf("Error generating OpenShift Clientset: %v\n", err)
		return false, nil, nil
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		logger.Printf("Error generating Dynamic Clientset: %v\n", err)
		return false, nil, nil
	}

	clusterHealthy := true

	var healthErr *multierror.Error
	switch providerType {
	case "rosa":
		fallthrough
	case "ocm":
		if check, err := healthchecks.CheckCVOReadiness(oscfg.ConfigV1(), logger); !check || err != nil {
			healthErr = multierror.Append(healthErr, err)
			failures = append(failures, "cvo")
			clusterHealthy = false
		}

		if check, err := healthchecks.CheckNodeHealth(kubeClient.CoreV1(), logger); !check || err != nil {
			healthErr = multierror.Append(healthErr, err)
			failures = append(failures, "node")
			clusterHealthy = false
		}

		if check, err := healthchecks.CheckMachinesObjectState(dynamicClient, logger); !check || err != nil {
			healthErr = multierror.Append(healthErr, err)
			failures = append(failures, "machine")
			clusterHealthy = false
		}

		if check, err := healthchecks.CheckOperatorReadiness(oscfg.ConfigV1(), logger); !check || err != nil {
			healthErr = multierror.Append(healthErr, err)
			failures = append(failures, "operator")
			clusterHealthy = false
		}

		if check, err := healthchecks.CheckCerts(kubeClient.CoreV1(), logger); !check || err != nil {
			healthErr = multierror.Append(healthErr, err)
			failures = append(failures, "cert")
			clusterHealthy = false
		}

		if check, err := healthchecks.CheckReplicaCountForDaemonSets(kubeClient.AppsV1(), logger); !check || err != nil {
			healthErr = multierror.Append(healthErr, err)
			failures = append(failures, "daemonset")
			clusterHealthy = false
		}

		if check, err := healthchecks.CheckReplicaCountForReplicaSets(kubeClient.AppsV1(), logger); !check || err != nil {
			healthErr = multierror.Append(healthErr, err)
			failures = append(failures, "replicaset")
			clusterHealthy = false
		}

	default:
		logger.Printf("No provisioner-specific logic for %q", providerType)
	}

	return clusterHealthy, failures, healthErr.ErrorOrNil()
}

func getRestConfig(provider spi.Provider, clusterID string) (*rest.Config, error) {
	var err error

	var kubeconfigBytes []byte
	kubeconfigContents := viper.GetString(config.Kubeconfig.Contents)
	kubeconfigPath := viper.GetString(config.Kubeconfig.Path)
	if len(kubeconfigContents) == 0 && len(kubeconfigPath) == 0 {
		if kubeconfigBytes, err = provider.ClusterKubeconfig(clusterID); err != nil {
			return nil, fmt.Errorf("could not get kubeconfig for cluster: %v", err)
		}
	} else if len(kubeconfigPath) != 0 {
		kubeconfigPath := viper.GetString(config.Kubeconfig.Path)
		kubeconfigBytes, err = ioutil.ReadFile(kubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed reading '%s' which has been set as the TEST_KUBECONFIG: %v", kubeconfigPath, err)
		}
	} else {
		kubeconfigBytes = []byte(kubeconfigContents)
	}

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeconfigBytes)
	if err != nil {
		return nil, fmt.Errorf("error generating rest config: %v", err)
	}

	return restConfig, nil
}

// ProvisionCluster will provision a cluster and immediately return.
func ProvisionCluster(logger *log.Logger) (*spi.Cluster, error) {
	logger = logging.CreateNewStdLoggerOrUseExistingLogger(logger)

	provider, err := providers.ClusterProvider()
	if err != nil {
		return nil, fmt.Errorf("error getting cluster provisioning client: %v", err)
	}

	var cluster *spi.Cluster
	// create a new cluster if no ID is specified
	clusterID := viper.GetString(config.Cluster.ID)
	if clusterID == "" {
		name := viper.GetString(config.Cluster.Name)
		if name == "" || name == "random" {
			attemptLimit := 10
			for attempt := 1; attempt <= attemptLimit; attempt++ {
				name = clusterName()
				validName, err := provider.IsValidClusterName(name)
				if err != nil {
					fmt.Printf("an error occurred validating the cluster name %v\n", err)
				} else if validName {
					break
				} else {
					fmt.Printf("cluster name %s already exists.\n", name)
				}
				fmt.Printf("retrying to validate cluster name. Attempt %d of %d\n", attempt, attemptLimit)
				if attempt == attemptLimit {
					return nil, fmt.Errorf("could not validate cluster name. timed out")
				}
			}
		}

		if clusterID, err = provider.LaunchCluster(name); err != nil {
			return nil, fmt.Errorf("could not launch cluster: %v", err)
		}

		if cluster, err = provider.GetCluster(clusterID); err != nil {
			return nil, fmt.Errorf("could not get cluster after launching: %v", err)
		}
	} else {
		logger.Printf("CLUSTER_ID of '%s' was provided, skipping cluster creation and using it instead", clusterID)

		cluster, err = provider.GetCluster(clusterID)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve cluster information from OCM: %v", err)
		}

		if cluster.State() == spi.ClusterStateHibernating && !provider.Resume(cluster.ID()) {
			return cluster, fmt.Errorf("cluster errored while resuming")
		}
	}

	return cluster, nil
}

// clusterName returns a cluster name with a format which must be short enough to support all versions
func clusterName() string {
	suffix := viper.GetString(config.Suffix)
	name := viper.GetString(config.Cluster.Name)

	if name == "random" {
		seed := time.Now().UTC().UnixNano()
		rand.Seed(seed)
		newName := ""
		prefixes := []string{"prod", "stg", "int", "p", "i", "s", "pre"}
		names := []string{"app", "db", "cache", "ocp", "openshift", "store", "control", "swap", "testing", "application", "user", "customer", "cust", "osd", "dedicated"}
		suffixes := []string{"0", "1", "2", "3", "5", "8", "13", "temp", "final"}

		doPrefix := rand.Intn(3)
		doSuffix := rand.Intn(3)

		newName = fmt.Sprintf("%s", names[rand.Intn(len(names))])

		if doPrefix > 0 && len(newName) <= 10 {
			newName = fmt.Sprintf("%s-%s", prefixes[rand.Intn(len(prefixes))], newName)
		}

		if doSuffix > 0 && len(newName) <= 10 {
			newName = fmt.Sprintf("%s-%s", newName, suffixes[rand.Intn(len(suffixes))])
		}

		if len(newName) > 15 {
			log.Printf("%s is longer than 15 characters. Generating a new name...", newName)
			newName = clusterName()
		}

		return newName
	}

	if suffix == "" {
		suffix = util.RandomStr(5)
	}

	return "osde2e-" + suffix
}
