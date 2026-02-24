package cluster

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/hashicorp/go-multierror"
	osconfig "github.com/openshift/client-go/config/clientset/versioned"
	"github.com/openshift/osde2e/pkg/common/cluster/healthchecks"
	"github.com/openshift/osde2e/pkg/common/clusterproperties"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/logging"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
	"github.com/openshift/osde2e/pkg/common/providers/rosaprovider"
	"github.com/openshift/osde2e/pkg/common/runner"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/util"
	"github.com/openshift/osde2e/pkg/common/versions"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	ReserveCount = 5
	// errorWindow is the number of checks made to determine if a cluster has truly failed.
	errorWindow = 20
	// pendingPodThreshold is the maximum number of times a pod is allowed to be in pending state before erroring out in PollClusterHealth.
	pendingPodThreshold = 10
)

// podErrorTracker is the data structure that keeps track of pending state counters for each pod against their pod UIDs.
var podErrorTracker healthchecks.PodErrorTracker

// ErrReserveFull is returned for early exit from provisioner
var ErrReserveFull = errors.New("reserve full")

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

// WaitForClusterReadyPostInstall blocks until the cluster is ready for testing using mechanisms appropriate
// for a newly-installed cluster.
func WaitForClusterReadyPostInstall(clusterID string, logger *log.Logger) error {
	logger = logging.CreateNewStdLoggerOrUseExistingLogger(logger)
	podErrorTracker.NewPodErrorTracker(pendingPodThreshold)
	provider, err := providers.ClusterProvider()
	if err != nil {
		return fmt.Errorf("error getting cluster provisioning client: %v", err)
	}

	cluster, err := provider.GetCluster(clusterID)
	if err != nil {
		return fmt.Errorf("failed getting cluster from provider: %w", err)
	}

	clusterConfig, _, err := ClusterConfig(clusterID)
	if err != nil {
		return fmt.Errorf("failed looking up cluster config for healthcheck: %w", err)
	}

	kubeClient, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return fmt.Errorf("error generating Kube Clientset: %w", err)
	}

	if viper.GetBool(config.Hypershift) {
		logger.Println("Waiting for nodes to be ready (up to 10 minutes)")
		err := wait.PollUntilContextTimeout(context.TODO(), 30*time.Second, 10*time.Minute, true, func(ctx context.Context) (bool, error) {
			nodes, err := kubeClient.CoreV1().Nodes().List(ctx, v1.ListOptions{})
			if err != nil {
				// if the api server is not ready yet
				if os.IsTimeout(err) {
					logger.Println(err)
					return false, nil
				}
				return false, err
			}
			for _, node := range nodes.Items {
				for _, condition := range node.Status.Conditions {
					if condition.Type == corev1.NodeReady && condition.Status != corev1.ConditionTrue {
						return false, nil
					}
				}
			}
			return len(nodes.Items) == cluster.NumComputeNodes(), nil
		})
		if err != nil {
			return fmt.Errorf("HyperShift healthcheck: %w", err)
		}
		return nil
	}

	if viper.GetBool(config.Tests.OnlyHealthCheckNodes) {
		logger.Println("Waiting up to 30 minutes for all nodes to be ready")
		err := wait.PollUntilContextTimeout(context.TODO(), 30*time.Second, 30*time.Minute, true, func(_ context.Context) (bool, error) {
			return healthchecks.CheckNodeHealth(kubeClient.CoreV1(), logger)
		})
		if err != nil {
			return fmt.Errorf("node health check failed: %w", err)
		}
		return nil
	}

	err = healthchecks.CheckHealthcheckJob(context.Background(), clusterConfig, logger)
	if err != nil {
		return fmt.Errorf("cluster failed health check: %w", err)
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

func WaitForOCMProvisioning(provider spi.Provider, clusterID string, logger *log.Logger, isUpgrade bool) (becameReadyAt time.Time, err error) {
	installTimeout := viper.GetInt64(config.Cluster.InstallTimeout)
	if viper.GetBool(config.Hypershift) {
		// Install timeout 30 minutes for hypershift
		installTimeout = 30
	}

	logger = logging.CreateNewStdLoggerOrUseExistingLogger(logger)

	logger.Printf("Waiting %v minutes for cluster to be ready...\n", installTimeout)

	readinessSet := false
	var readinessStarted time.Time

	healthcheckStatus := clusterproperties.StatusHealthCheck
	if isUpgrade {
		healthcheckStatus = clusterproperties.StatusUpgradeHealthCheck
	}

	return readinessStarted, wait.PollUntilContextTimeout(context.TODO(), 30*time.Second, time.Duration(installTimeout)*time.Minute, true, func(_ context.Context) (bool, error) {
		cluster, err := provider.GetCluster(clusterID)
		if err != nil {
			logger.Printf("Error fetching cluster details from provider: %s", err)
			return false, nil
		}

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
			if err := provider.AddProperty(cluster, clusterproperties.Status, healthcheckStatus); err != nil {
				logger.Printf("error trying to add health-check property to cluster ID %s: %v", cluster.ID(), err)
				return false, nil
			}

			readinessStarted = time.Now()
			return true, nil
		} else if cluster.State() == spi.ClusterStateError {
			logger.Print("cluster is in error state, check cloud provider for more details")
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
	cleanRunsNeeded := viper.GetInt(config.Cluster.CleanCheckRuns)
	cleanRuns := 0
	errRuns := 0

	_, err = WaitForOCMProvisioning(provider, clusterID, logger, isUpgrade)
	if err != nil {
		return fmt.Errorf("OCM never became ready: %w", err)
	}

	cluster, err := provider.GetCluster(clusterID)
	if err != nil {
		return fmt.Errorf("error fetching cluster details from provider: %w", err)
	}

	if pollErr := wait.PollUntilContextTimeout(context.TODO(), 30*time.Second, time.Duration(installTimeout)*time.Minute, true, func(_ context.Context) (bool, error) {
		if cluster.State() != spi.ClusterStateReady {
			logger.Printf("Cluster is not ready, current status '%s'.", cluster.State())
			return false, nil
		}

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
		kubeconfigBytes, err = os.ReadFile(kubeconfigPath)
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
	// get cluster ID from env
	clusterID := viper.GetString(config.Cluster.ID)
	ocmProvider, err := ocmprovider.NewWithEnv(viper.GetString(ocmprovider.Env))
	if err != nil {
		return nil, fmt.Errorf("could not setup ocm provider: %v", err)
	}
	// Only enable cluster reserve claiming for ROSA STS classic for now
	if viper.GetBool(config.Cluster.UseClusterReserve) && viper.GetString(config.Provider) == "rosa" && !viper.GetBool(config.Hypershift) && viper.GetBool(rosaprovider.STS) {
		clusterID = ocmProvider.ClaimClusterFromReserve(viper.GetString(config.Cluster.Version), "aws", "rosa")
	}
	if viper.GetBool(config.Cluster.Reserve) {
		logger.Printf("Cluster reserve provisioning requested, querying reserve")
		listResponse, err := ocmProvider.QueryReserve(viper.GetString(config.Cluster.Version), viper.GetString(config.CloudProvider.CloudProviderID), viper.GetString(config.Provider))
		if err != nil {
			return nil, fmt.Errorf("could not query reserve: %v", err)
		}
		logger.Printf("Reserve count: %d", listResponse.Total())
		if listResponse.Total() >= ReserveCount {
			// Provision one cluster per job run. Job should be scheduled every so often to keep adding to reserve so that count is met.
			return nil, ErrReserveFull
		}
	}

	// create a new cluster if no ID is specified
	if clusterID == "" {
		log.Printf("no clusterid found, provisioning cluster")
		name := viper.GetString(config.Cluster.Name)
		if name == "" || name == "random" {
			attemptLimit := 10
			for attempt := 1; attempt <= attemptLimit; attempt++ {
				name = clusterName()
				validName, err := provider.IsValidClusterName(name)
				if err != nil {
					fmt.Printf("an error occurred validating the cluster name %v\n", err)
				} else if validName {
					viper.Set(config.Cluster.Name, name)
					log.Printf("cluster name set to %s\n", name)
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

		clusterID, err = provider.LaunchCluster(name)
		if clusterID != "" {
			viper.Set(config.Cluster.ID, clusterID)
		}
		if err != nil {
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

	}

	return cluster, nil
}

// clusterName returns a cluster name with a format which must be short enough to support all versions
func clusterName() string {
	suffix := viper.GetString(config.Suffix)
	name := viper.GetString(config.Cluster.Name)

	if name == "random" {
		newName := ""
		prefixes := []string{"prod", "stg", "int", "p", "i", "s", "pre"}
		names := []string{"app", "db", "cache", "ocp", "openshift", "store", "control", "swap", "testing", "application", "user", "customer", "cust", "osd", "dedicated"}
		suffixes := []string{"0", "1", "2", "3", "5", "8", "13", "temp", "final"}

		doPrefix := rand.Intn(3)
		doSuffix := rand.Intn(3)

		newName = names[rand.Intn(len(names))]

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

func Provision(provider spi.Provider) (*spi.Cluster, error) {
	status, err := ConfigureVersion(provider)
	if status != config.Success {
		return nil, fmt.Errorf("failed configure cluster version: %v", err)
	}
	cluster, err := ProvisionCluster(nil)
	if err != nil {
		if errors.Is(err, ErrReserveFull) {
			log.Printf("Reserve full, exiting without provisioning")
		}
		return nil, fmt.Errorf("failed to set up or retrieve cluster: %w", err)
	}

	viper.Set(config.Cluster.ID, cluster.ID())
	log.Printf("CLUSTER_ID set to %s from OCM.", viper.GetString(config.Cluster.ID))
	_, err = WaitForOCMProvisioning(provider, viper.GetString(config.Cluster.ID), nil, false)
	if err != nil {
		return nil, fmt.Errorf("cluster never became ready: %v", err)
	}
	log.Printf("Cluster status is ready")

	if viper.GetString(config.SharedDir) != "" {
		if err = os.WriteFile(fmt.Sprintf("%s/cluster-id", viper.GetString(config.SharedDir)), []byte(cluster.ID()), 0o644); err != nil {
			log.Printf("Error writing cluster ID to shared directory: %v", err)
		} else {
			log.Printf("Wrote cluster ID to shared dir: %v", cluster.ID())
		}
	} else {
		log.Printf("No shared directory provided, skip writing cluster ID")
	}

	if (!viper.GetBool(config.Addons.SkipAddonList) || viper.GetString(config.Provider) != "mock") && len(cluster.Addons()) > 0 {
		log.Printf("Found addons: %s", strings.Join(cluster.Addons(), ","))
	}

	if err = provider.AddProperty(cluster, "UpgradeVersion", viper.GetString(config.Upgrade.ReleaseName)); err != nil {
		log.Printf("Error while adding upgrade version property to cluster via OCM: %v", err)
	}

	if !viper.GetBool(config.Tests.SkipClusterHealthChecks) {
		// If this is a new cluster, we should check the OSD Ready job unless skipped
		err = WaitForClusterReadyPostInstall(cluster.ID(), nil)
		if err != nil {
			log.Println("*******************")
			log.Printf("Cluster failed health check: %v", err)
			log.Println("*******************")
		} else {
			log.Println("Cluster is healthy and ready for testing")
		}
	} else {
		log.Println("Skipping health checks as requested")
	}

	var kubeconfigBytes []byte
	clusterConfigerr := wait.PollUntilContextTimeout(context.Background(), 2*time.Second, 5*time.Minute, true, func(ctx context.Context) (bool, error) {
		kubeconfigBytes, err = provider.ClusterKubeconfig(viper.GetString(config.Cluster.ID))
		if err != nil {
			log.Printf("Failed to retrieve kubeconfig: %v\nWaiting two seconds before retrying", err)
			return false, nil
		} else {
			log.Printf("Successfully retrieved kubeconfig from OCM.")
			viper.Set(config.Kubeconfig.Contents, string(kubeconfigBytes))
			return true, nil
		}
	})

	if clusterConfigerr != nil {
		return nil, fmt.Errorf("failed retrieving kubeconfig: %v", clusterConfigerr)
	}

	if viper.GetString(config.SharedDir) != "" {
		if err = os.WriteFile(fmt.Sprintf("%s/kubeconfig", viper.GetString(config.SharedDir)), kubeconfigBytes, 0o644); err != nil {
			log.Printf("Error writing cluster kubeconfig to shared directory: %v", err)
		} else {
			log.Printf("Passed kubeconfig to prow steps.")
		}
	}

	return cluster, nil
}

// set cluster infor into viper and metadata
func SetClusterIntoViperConfig(cluster *spi.Cluster) {
	viper.Set(config.Cluster.Channel, cluster.ChannelGroup())
	viper.Set(config.Cluster.Name, cluster.Name())
	log.Printf("CLUSTER_NAME set to %s from OCM.", viper.GetString(config.Cluster.Name))

	viper.Set(config.Cluster.Version, cluster.Version())
	log.Printf("CLUSTER_VERSION set to %s from OCM, for channel group %s", viper.GetString(config.Cluster.Version), viper.GetString(config.Cluster.Channel))

	viper.Set(config.CloudProvider.CloudProviderID, cluster.CloudProvider())
	log.Printf("CLOUD_PROVIDER_ID set to %s from OCM.", viper.GetString(config.CloudProvider.CloudProviderID))

	viper.Set(config.CloudProvider.Region, cluster.Region())
	log.Printf("CLOUD_PROVIDER_REGION set to %s from OCM.", viper.GetString(config.CloudProvider.Region))
}

func ConfigureVersion(provider spi.Provider) (int, error) {
	// configure cluster and upgrade versions
	versionSelector := versions.VersionSelector{Provider: provider}
	if err := versionSelector.SelectClusterVersions(); err != nil {
		// If we can't find a version to use, exit with an error code.
		return config.Failure, err
	}

	switch {
	case !viper.GetBool(config.Cluster.EnoughVersionsForOldestOrMiddleTest):
		return config.Aborted, fmt.Errorf("there were not enough available cluster image sets to choose and oldest or middle cluster image set to test against -- skipping tests")
	case !viper.GetBool(config.Cluster.PreviousVersionFromDefaultFound):
		return config.Aborted, fmt.Errorf("no previous version from default found with the given arguments")
	case viper.GetBool(config.Upgrade.UpgradeVersionEqualToInstallVersion):
		return config.Aborted, fmt.Errorf("install version and upgrade version are the same -- skipping tests")
	case viper.GetString(config.Upgrade.ReleaseName) == util.NoVersionFound:
		return config.Aborted, fmt.Errorf("no valid upgrade versions were found. Skipping tests")
	case viper.GetString(config.Cluster.Version) == "":
		returnState := config.Aborted
		if viper.GetBool(config.Cluster.LatestYReleaseAfterProdDefault) || viper.GetBool(config.Cluster.LatestZReleaseAfterProdDefault) {
			log.Println("At the latest available version with no newer targets. Exiting...")
			returnState = config.Success
		}
		return returnState, fmt.Errorf("no valid install version found")
	}
	return config.Success, nil
}

// ProvisionOrReuseCluster either provisions a new cluster or retrieves an existing one
// based on whether a kubeconfig is already available.
func ProvisionOrReuseCluster(provider spi.Provider) (*spi.Cluster, error) {
	var cluster *spi.Cluster
	var err error

	if viper.GetString(config.Kubeconfig.Contents) == "" {
		// Provision new cluster
		cluster, err = Provision(provider)
		if err != nil {
			return nil, fmt.Errorf("cluster provisioning failed: %w", err)
		}
	} else {
		// Reuse existing cluster
		log.Println("Using provided kubeconfig")
		clusterID := viper.GetString(config.Cluster.ID)
		cluster, err = provider.GetCluster(clusterID)
		if err != nil {
			return nil, fmt.Errorf("failed to get cluster %s: %w", clusterID, err)
		}
	}

	// Set cluster into viper config for downstream usage
	SetClusterIntoViperConfig(cluster)
	return cluster, nil
}

// InstallAddonsIfConfigured installs addons on the cluster if configured in viper.
// Returns true if addons were installed, false otherwise.
func InstallAddonsIfConfigured(provider spi.Provider, clusterID string) (bool, error) {
	addonIDsStr := viper.GetString(config.Addons.IDs)
	if len(addonIDsStr) == 0 {
		return false, nil
	}

	// Skip addon installation for mock provider
	if viper.GetString(config.Provider) == "mock" {
		log.Println("Skipping addon installation for mock provider")
		return false, nil
	}

	addonIDs := strings.Split(addonIDsStr, ",")
	params := make(map[string]map[string]string)

	// Parse addon parameters if provided
	if strParams := viper.GetString(config.Addons.Parameters); strParams != "" {
		var err error
		if err = json.Unmarshal([]byte(strParams), &params); err != nil {
			return false, fmt.Errorf("failed to unmarshal addon parameters: %w", err)
		}
	}

	// Install addons
	num, err := provider.InstallAddons(clusterID, addonIDs, params)
	if err != nil {
		return false, fmt.Errorf("failed to install addons: %w", err)
	}

	// Wait for cluster to be ready after addon installation
	if num > 0 {
		if err := WaitForClusterReadyPostInstall(clusterID, nil); err != nil {
			return false, fmt.Errorf("cluster not ready after addon installation: %w", err)
		}
	}

	return num > 0, nil
}

// DeleteCluster destroys the cluster if configured.
func DeleteCluster(provider spi.Provider) error {
	clusterID := viper.GetString(config.Cluster.ID)
	if clusterID == "" {
		return nil
	}

	if !viper.GetBool(config.Cluster.SkipDestroyCluster) {
		log.Printf("Destroying cluster '%s'...", clusterID)
		if err := provider.DeleteCluster(clusterID); err != nil {
			return fmt.Errorf("failed to delete cluster: %w", err)
		}
	} else {
		log.Printf("Cluster %s preserved in environment %s", clusterID, provider.Environment())
	}

	return nil
}

// UpdateClusterProperties updates cluster metadata in the provider.
func UpdateClusterProperties(provider spi.Provider, status string) error {
	clusterID := viper.GetString(config.Cluster.ID)
	cluster, err := provider.GetCluster(clusterID)
	if err != nil {
		return fmt.Errorf("failed to get cluster: %w", err)
	}

	log.Printf("Cluster state: %v, flavor: %s", cluster.State(), cluster.Flavour())

	if viper.GetBool(config.Cluster.Passing) {
		status = clusterproperties.StatusCompletedPassing
	}

	properties := map[string]string{
		clusterproperties.Status:       status,
		clusterproperties.JobID:        "",
		clusterproperties.JobName:      "",
		clusterproperties.Availability: clusterproperties.Used,
	}

	for key, value := range properties {
		if err := provider.AddProperty(cluster, key, value); err != nil {
			return fmt.Errorf("failed to set property %s: %w", key, err)
		}
	}

	return nil
}

// HandleExpirationExtension extends cluster expiration for specific scenarios.
func HandleExpirationExtension(provider spi.Provider) error {
	clusterID := viper.GetString(config.Cluster.ID)
	if clusterID == "" || !viper.GetBool(config.Cluster.SkipDestroyCluster) {
		return nil
	}

	// Extend expiration for nightly builds
	if viper.GetString(config.Cluster.InstallSpecificNightly) != "" || viper.GetString(config.Cluster.ReleaseImageLatest) != "" {
		if err := provider.Expire(clusterID, 30*time.Minute); err != nil {
			return err
		}
	}

	// Extend expiration for passing clusters without addons
	if !viper.GetBool(config.Cluster.ClaimedFromReserve) && viper.GetString(config.Addons.IDs) == "" {
		cluster, err := provider.GetCluster(clusterID)
		if err != nil {
			return err
		}

		if !cluster.ExpirationTimestamp().Add(6 * time.Hour).After(cluster.CreationTimestamp().Add(24 * time.Hour)) {
			if err := provider.ExtendExpiry(clusterID, 6, 0, 0); err != nil {
				return err
			}
		}
	}

	return nil
}

// LoadClusterContext loads kubeconfig and cluster ID from configuration.
// This should be called before provisioning to ensure context is available.
func LoadClusterContext() error {
	// Load kubeconfig if available
	if err := config.LoadKubeconfig(); err != nil {
		log.Printf("Not loading kubeconfig: %v", err)
	}

	// Load cluster ID from shared directory
	if err := config.LoadClusterId(); err != nil {
		log.Printf("Not loading cluster id: %v", err)
		return fmt.Errorf("failed to load cluster ID: %w", err)
	}

	return nil
}

// RunMustGather executes must-gather and collects results.
func RunMustGather(ctx context.Context, h *helper.H) error {
	log.Print("Running must-gather...")
	h.SetServiceAccount(ctx, "system:serviceaccount:%s:cluster-admin")

	r := h.Runner(fmt.Sprintf("oc adm must-gather --dest-dir=%v", runner.DefaultRunner.OutputDir))
	r.Name = "must-gather"
	r.Tarball = true

	if err := r.Run(1800, make(chan struct{})); err != nil {
		return fmt.Errorf("must-gather failed: %w", err)
	}

	results, err := r.RetrieveResults()
	if err != nil {
		return fmt.Errorf("failed to retrieve must-gather results: %w", err)
	}

	h.WriteResults(results)
	return nil
}

// InspectClusterState gathers cluster state information.
func InspectClusterState(ctx context.Context, h *helper.H) {
	log.Print("Gathering project states...")
	h.InspectState(ctx)

	log.Print("Gathering OLM state...")
	if err := h.InspectOLM(ctx); err != nil {
		log.Printf("Error inspecting OLM: %v", err)
	}
}
