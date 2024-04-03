package rosaprovider

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"

	rosaprovider "github.com/openshift/osde2e-common/pkg/openshift/rosa"
)

// LaunchCluster creates the cluster
func (m *ROSAProvider) LaunchCluster(clusterName string) (string, error) {
	var (
		ctx = context.Background()

		channelGroup = viper.GetString(config.Cluster.Channel)
		version      = viper.GetString(config.Cluster.Version)
	)

	version = strings.ReplaceAll(strings.ReplaceAll(version, "openshift-v", ""), fmt.Sprintf("-%s", channelGroup), "")
	replicas, err := strconv.Atoi(viper.GetString(Replicas))
	if err != nil {
		return "", fmt.Errorf("rosa launch cluster failed to convert replicas to integer: %v", err)
	}

	hostPrefix, err := strconv.Atoi(viper.GetString(HostPrefix))
	if err != nil {
		return "", fmt.Errorf("rosa launch cluster failed to convert host prefix to integer: %v", err)
	}

	clusterProperties, err := m.ocmProvider.GenerateProperties()
	if err != nil {
		return "", fmt.Errorf("error generating cluster properties: %v", err)
	}

	artifactDir := viper.GetString(config.ReportDir)
	workingDir := viper.GetString(config.ReportDir)
	if viper.GetString(config.SharedDir) != "" {
		workingDir = viper.GetString(config.SharedDir)
	}

	clusterProperties["workingDir"] = workingDir

	installConfig := ""

	// This skips setting install_config for any prod job OR any periodic addon job.
	// To invoke this logic locally you will have to set JOB_TYPE to "periodic".
	if m.Environment() != "prod" {
		if viper.GetString(config.JobType) == "periodic" && !strings.Contains(viper.GetString(config.JobName), "addon") {
			imageSource := viper.GetString(config.Cluster.ImageContentSource)
			installConfig += "\n" + m.ChooseImageSource(imageSource)
		}
	}

	if viper.GetString(config.Cluster.InstallConfig) != "" {
		installConfig += "\n" + viper.GetString(config.Cluster.InstallConfig)
	}

	if installConfig != "" {
		log.Println("Install config:", installConfig)
		clusterProperties["install_config"] = installConfig
	}

	installTimeout := viper.GetInt64(config.Cluster.InstallTimeout)

	healthCheckTimeout, err := time.ParseDuration(viper.GetString(config.Tests.ClusterHealthChecksTimeout))
	if err != nil {
		return "", fmt.Errorf("error parsing health check timeout: %v", err)
	}

	if viper.GetBool(config.Hypershift) {
		installTimeout = 30
		healthCheckTimeout = 20 * time.Minute
	}

	return m.provider.CreateCluster(
		ctx,
		&rosaprovider.CreateClusterOptions{
			ClusterName:                  clusterName,
			ChannelGroup:                 channelGroup,
			Version:                      version,
			SubnetIDs:                    viper.GetString(config.AWSVPCSubnetIDs),
			ComputeMachineType:           viper.GetString(ComputeMachineType),
			MachineCidr:                  viper.GetString(MachineCIDR),
			ArtifactDir:                  artifactDir,
			WorkingDir:                   workingDir,
			ServiceCIDR:                  viper.GetString(ServiceCIDR),
			PodCIDR:                      viper.GetString(PodCIDR),
			NetworkType:                  viper.GetString(config.Cluster.NetworkProvider),
			HTTPProxy:                    viper.GetString(config.Proxy.HttpProxy),
			HTTPSProxy:                   viper.GetString(config.Proxy.HttpsProxy),
			AdditionalTrustBundleFile:    viper.GetString(config.Proxy.UserCABundle),
			OidcConfigID:                 viper.GetString(OIDCConfigID),
			Properties:                   clusterProperties,
			HostPrefix:                   hostPrefix,
			Replicas:                     replicas,
			STS:                          viper.GetBool(STS),
			HostedCP:                     viper.GetBool(config.Hypershift),
			FIPS:                         viper.GetBool(config.Cluster.EnableFips),
			MultiAZ:                      viper.GetBool(config.Cluster.MultiAZ),
			ExpirationDuration:           viper.GetDuration(config.Cluster.ExpiryInMinutes) * time.Minute,
			SkipHealthCheck:              viper.GetBool(config.Tests.SkipClusterHealthChecks),
			UseDefaultAccountRolesPrefix: viper.GetBool(STSUseDefaultAccountRolesPrefix),
			InstallTimeout:               time.Duration(installTimeout) * time.Minute,
			HealthCheckTimeout:           healthCheckTimeout,
		},
	)
}

// DeleteCluster deletes the provisioned cluster
func (m *ROSAProvider) DeleteCluster(clusterID string) error {
	ctx := context.Background()

	defer func() {
		_ = m.provider.Client.Close()
		_ = m.provider.Uninstall(ctx)
	}()

	workingDir, err := m.GetProperty(clusterID, "workingDir")
	if err != nil {
		return fmt.Errorf("rosa delete cluster: failed to get clusters report directory: %v", err)
	}

	deleteHostedCPVPC := true
	if viper.GetString(config.AWSVPCSubnetIDs) != "" {
		deleteHostedCPVPC = false
	}

	deleteOidcConfigID := true
	if viper.GetString(OIDCConfigID) != "" {
		deleteOidcConfigID = false
	}

	return m.provider.DeleteCluster(
		ctx,
		&rosaprovider.DeleteClusterOptions{
			ArtifactDir:        viper.GetString(config.ReportDir),
			WorkingDir:         workingDir,
			ClusterName:        clusterID,
			HostedCP:           viper.GetBool(config.Hypershift),
			STS:                viper.GetBool(STS),
			DeleteHostedCPVPC:  deleteHostedCPVPC,
			DeleteOidcConfigID: deleteOidcConfigID,
		},
	)
}

// IsValidClusterName validates whether the cluster name is valid or not before launching it
func (m *ROSAProvider) IsValidClusterName(clusterName string) (bool, error) {
	ctx := context.Background()
	size := 50
	page := 1

	collection := m.ocmProvider.GetConnection().ClustersMgmt().V1().Clusters()

	for {
		response, err := collection.List().
			Search(fmt.Sprintf("name = '%s'", clusterName)).
			Size(size).
			Page(page).
			SendContext(ctx)
		if err != nil {
			return false, fmt.Errorf("can't retrieve page %d: %s", page, err)
		}

		if response.Total() != 0 {
			return false, nil
		}

		if response.Size() < size {
			break
		}

		page++
	}

	return true, nil
}
