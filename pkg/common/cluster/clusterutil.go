package cluster

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/hashicorp/go-multierror"
	osconfig "github.com/openshift/client-go/config/clientset/versioned"
	"github.com/openshift/osde2e/pkg/common/cluster/healthchecks"
	"github.com/openshift/osde2e/pkg/common/clusterproperties"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/logging"
	"github.com/openshift/osde2e/pkg/common/metadata"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/util"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// errorWindow is the number of checks made to determine if a cluster has truly failed.
	errorWindow = 20
)

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

	return waitForClusterReadyWithOverrideAndExpectedNumberOfNodes(clusterID, nil, true)
}

// WaitForClusterReady blocks until the cluster is ready for testing.
func WaitForClusterReady(clusterID string, logger *log.Logger) error {
	return waitForClusterReadyWithOverrideAndExpectedNumberOfNodes(clusterID, logger, false)
}

func waitForClusterReadyWithOverrideAndExpectedNumberOfNodes(clusterID string, logger *log.Logger, overrideSkipCheck bool) error {
	logger = logging.CreateNewStdLoggerOrUseExistingLogger(logger)

	provider, err := providers.ClusterProvider()

	if err != nil {
		return fmt.Errorf("error getting cluster provisioning client: %v", err)
	}

	installTimeout := viper.GetInt64(config.Cluster.InstallTimeout)
	logger.Printf("Waiting %v minutes for cluster '%s' to be ready...\n", installTimeout, clusterID)
	cleanRunsNeeded := viper.GetInt(config.Cluster.CleanCheckRuns)
	cleanRuns := 0
	errRuns := 0

	clusterStarted := time.Now()
	var readinessStarted time.Time
	ocmReady := false
	readinessSet := false

	if !viper.GetBool(config.Tests.SkipClusterHealthChecks) || overrideSkipCheck {
		return wait.PollImmediate(30*time.Second, time.Duration(installTimeout)*time.Minute, func() (bool, error) {
			cluster, err := provider.GetCluster(clusterID)
			if err != nil {
				log.Printf("Error fetching cluster details from provider: %s", err)
				return false, nil
			}

			properties := cluster.Properties()
			currentStatus := properties[clusterproperties.Status]

			if currentStatus == clusterproperties.StatusProvisioning && !readinessSet {
				err = provider.AddProperty(cluster, clusterproperties.Status, clusterproperties.StatusWaitingForReady)
				if err != nil {
					log.Printf("Error adding property to cluster: %s", err.Error())
					return false, nil
				}

			}

			if cluster != nil && cluster.State() == spi.ClusterStateReady {
				// This is the first time that we've entered this section, so we'll consider this the time until OCM has said the cluster is ready
				if !ocmReady {
					ocmReady = true
					if metadata.Instance.TimeToOCMReportingInstalled == 0 {
						metadata.Instance.SetTimeToOCMReportingInstalled(time.Since(clusterStarted).Seconds())
					}

					if currentStatus == clusterproperties.StatusUpgrading {
						err = provider.AddProperty(cluster, clusterproperties.Status, clusterproperties.StatusUpgradeHealthCheck)
					} else {
						err = provider.AddProperty(cluster, clusterproperties.Status, clusterproperties.StatusHealthCheck)
					}

					if err != nil {
						log.Printf("error trying to add health-check property to cluster ID %s: %v", cluster.ID(), err)
						return false, nil
					}

					readinessStarted = time.Now()
				}
				if success, failures, err := PollClusterHealth(clusterID, logger); success {
					cleanRuns++
					logger.Printf("Clean run %d/%d...", cleanRuns, cleanRunsNeeded)
					errRuns = 0
					if cleanRuns == cleanRunsNeeded {
						if metadata.Instance.TimeToClusterReady == 0 {
							metadata.Instance.SetTimeToClusterReady(time.Since(readinessStarted).Seconds())
						} else {
							metadata.Instance.SetTimeToUpgradedClusterReady(time.Since(readinessStarted).Seconds())
						}

						if currentStatus == clusterproperties.StatusUpgradeHealthCheck {
							err = provider.AddProperty(cluster, clusterproperties.Status, clusterproperties.StatusUpgradeHealthy)
						} else {
							err = provider.AddProperty(cluster, clusterproperties.Status, clusterproperties.StatusHealthy)
						}

						if err != nil {
							log.Printf("error trying to add healthy property to cluster ID %s: %v", cluster.ID(), err)
							return false, nil
						}

						return true, nil
					}
					return false, nil
				} else {
					if err != nil {
						errRuns++
						logger.Printf("Error in PollClusterHealth: %v", err)
						if errRuns >= errorWindow {
							if currentStatus == clusterproperties.StatusUpgradeHealthCheck {
								err = provider.AddProperty(cluster, clusterproperties.Status, clusterproperties.StatusUpgradeUnhealthy)
							} else {
								err = provider.AddProperty(cluster, clusterproperties.Status, clusterproperties.StatusUnhealthy)
							}

							if err != nil {
								log.Printf("error trying to add unhealthy property to cluster ID %s: %v", clusterID, err)
								return false, nil
							}
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
			} else if cluster == nil {
				return false, fmt.Errorf("the cluster is null despite there being no error: please check the logs")
			} else if cluster.State() == spi.ClusterStateError {
				return false, fmt.Errorf("the installation of cluster '%s' has errored", clusterID)
			} else {
				logger.Printf("Cluster is not ready, current status '%s'.", cluster.State())
			}
			return false, nil
		})
	}
	return nil
}

// PollClusterHealth looks at CVO data to determine if a cluster is alive/healthy or not
func PollClusterHealth(clusterID string, logger *log.Logger) (status bool, failures []string, err error) {
	logger = logging.CreateNewStdLoggerOrUseExistingLogger(logger)

	provider, err := providers.ClusterProvider()

	if err != nil {
		return false, nil, fmt.Errorf("error getting cluster provisioning client: %v", err)
	}

	logger.Print("Polling Cluster Health...\n")
	restConfig, err := getRestConfig(provider, clusterID)
	if err != nil {
		logger.Printf("Error generating Rest Config: %v\n", err)
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
	switch provider.Type() {
	case "moa":
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

		if check, err := healthchecks.CheckPodHealth(kubeClient.CoreV1(), logger); !check || err != nil {
			healthErr = multierror.Append(healthErr, err)
			failures = append(failures, "pod")
			clusterHealthy = false
		}

		if check, err := healthchecks.CheckCerts(kubeClient.CoreV1(), logger); !check || err != nil {
			healthErr = multierror.Append(healthErr, err)
			failures = append(failures, "cert")
			clusterHealthy = false
		}
	default:
		logger.Printf("No provisioner-specific logic for %s", provider.Type())
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

	// if TEST_KUBECONFIG has been set, skip configuring OCM
	if len(viper.GetString(config.Kubeconfig.Contents)) > 0 || len(viper.GetString(config.Kubeconfig.Path)) > 0 {
		return nil, useKubeconfig(logger)
	}

	provider, err := providers.ClusterProvider()

	if err != nil {
		return nil, fmt.Errorf("error getting cluster provisioning client: %v", err)
	}

	var cluster *spi.Cluster
	// create a new cluster if no ID is specified
	clusterID := viper.GetString(config.Cluster.ID)
	if clusterID == "" {
		name := viper.GetString(config.Cluster.Name)
		if name == "" {
			name = clusterName()
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
	}

	return cluster, nil
}

// useKubeconfig reads the path provided for a TEST_KUBECONFIG and uses it for testing.
func useKubeconfig(logger *log.Logger) (err error) {
	_, err = clientcmd.RESTConfigFromKubeConfig([]byte(viper.GetString(config.Kubeconfig.Contents)))
	if err != nil {
		logger.Println("Not an existing Kubeconfig, attempting to read file instead...")
	} else {
		logger.Println("Existing valid kubeconfig!")
		return nil
	}

	kubeconfigPath := viper.GetString(config.Kubeconfig.Path)
	_, err = ioutil.ReadFile(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed reading '%s' which has been set as the TEST_KUBECONFIG: %v", kubeconfigPath, err)
	}
	logger.Printf("Using a set TEST_KUBECONFIG of '%s' for Origin API calls.", kubeconfigPath)
	return nil
}

// clusterName returns a cluster name with a format which must be short enough to support all versions
func clusterName() string {
	suffix := viper.GetString(config.Suffix)

	if suffix == "" {
		suffix = util.RandomStr(5)
	}

	return "osde2e-" + suffix
}
