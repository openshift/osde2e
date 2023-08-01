package rosaprovider

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	createCluster "github.com/openshift/rosa/cmd/create/cluster"
	"github.com/openshift/rosa/cmd/dlt/oidcprovider"
	"github.com/openshift/rosa/cmd/dlt/operatorrole"
	rosaLogin "github.com/openshift/rosa/cmd/login"
	rosaAws "github.com/openshift/rosa/pkg/aws"
	"github.com/openshift/rosa/pkg/logging"
	"github.com/openshift/rosa/pkg/ocm"
	"k8s.io/apimachinery/pkg/util/wait"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/util"
)

// IsValidClusterName validates the clustername prior to proceeding with it
// in launching a cluster.
func (m *ROSAProvider) IsValidClusterName(clusterName string) (bool, error) {
	// Create a context:
	ctx := context.Background()

	collection := m.ocmProvider.GetConnection().ClustersMgmt().V1().Clusters()

	// Retrieve the list of clusters using pages of ten items, till we get a page that has less
	// items than requests, as that marks the end of the collection:
	size := 50
	page := 1
	searchPhrase := fmt.Sprintf("name = '%s'", clusterName)
	for {
		// Retrieve the page:
		response, err := collection.List().
			Search(searchPhrase).
			Size(size).
			Page(page).
			SendContext(ctx)
		if err != nil {
			return false, fmt.Errorf("can't retrieve page %d: %s", page, err)
		}

		if response.Total() != 0 {
			return false, nil
		}

		// Break the loop if the size of the page is less than requested, otherwise go to
		// the next page:
		if response.Size() < size {
			break
		}
		page++
	}

	// Name is valid.
	return true, nil
}

// LaunchCluster will provision an AWS cluster.
// nolint:gocyclo
func (m *ROSAProvider) LaunchCluster(clusterName string) (string, error) {
	// Calculate an expiration date for the cluster so that it will be automatically deleted if
	// we happen to forget to do it:
	var expiration time.Time
	var err error
	var awsCreator *rosaAws.Creator
	var awsVPCSubnetIds string
	var accountRoles *AccountRoles

	clusterProperties, err := m.ocmProvider.GenerateProperties()
	if err != nil {
		return "", fmt.Errorf("error generating cluster properties: %v", err)
	}

	if viper.GetBool(config.Hypershift) {
		validRegion, err := m.IsRegionValidForHCP()
		if err != nil || !validRegion {
			return "", err
		}
	}

	if viper.GetString(config.AWSVPCSubnetIDs) == "" && viper.GetBool(config.Hypershift) {
		var vpc *HyperShiftVPC
		vpc, err = createHyperShiftVPC()
		if err != nil {
			return "error creating aws vpc", err
		}

		awsVPCSubnetIds = fmt.Sprintf("%s,%s", vpc.PrivateSubnet, vpc.PublicSubnet)
		log.Printf("AWS VPC created at runtime, subnet ids: %s", awsVPCSubnetIds)

		// Save the report directory to cluster properties for later use when locating terraform state file to destroy vpc
		clusterProperties["reportDir"] = viper.GetString(config.ReportDir)
	}

	rosaClusterVersion := viper.GetString(config.Cluster.Version)
	channelGroup := viper.GetString(config.Cluster.Channel)
	rosaClusterVersion = strings.ReplaceAll(rosaClusterVersion, "openshift-v", "")
	rosaClusterVersion = strings.ReplaceAll(rosaClusterVersion, fmt.Sprintf("-%s", channelGroup), "")

	log.Printf("ROSA cluster version: %s", rosaClusterVersion)

	if viper.GetBool(config.Cluster.UseExistingCluster) && viper.GetString(config.Addons.IDs) == "" {
		if clusterID := m.ocmProvider.FindRecycledCluster(rosaClusterVersion, "aws", "rosa"); clusterID != "" {
			return clusterID, nil
		}
	}

	expiryInMinutes := viper.GetDuration(config.Cluster.ExpiryInMinutes)
	if expiryInMinutes > 0 {
		expiration = time.Now().Add(expiryInMinutes * time.Minute).UTC() // UTC() to workaround SDA-1567.
	}

	// ROSA uses the AWS provider in the background, so we'll determine region this way.
	_, err = m.DetermineRegion("aws")
	if err != nil {
		return "", fmt.Errorf("error determining region to use: %v", err)
	}

	createClusterArgs := []string{
		"--cluster-name", clusterName,
		"--region", m.awsRegion,
		"--channel-group", viper.GetString(config.Cluster.Channel),
		"--version", rosaClusterVersion,
		"--expiration-time", expiration.Format(time.RFC3339),
		"--compute-machine-type", viper.GetString(ComputeMachineType),
		"--machine-cidr", viper.GetString(MachineCIDR),
		"--service-cidr", viper.GetString(ServiceCIDR),
		"--pod-cidr", viper.GetString(PodCIDR),
		"--host-prefix", viper.GetString(HostPrefix),
		"--mode", "auto",
		"--yes",
	}

	// Auto create account roles if required
	if viper.GetBool(STS) {
		version, err := util.OpenshiftVersionToSemver(rosaClusterVersion)
		if err != nil {
			return "", fmt.Errorf("error parsing %s to semantic version: %v", rosaClusterVersion, err)
		}
		majorMinor := fmt.Sprintf("%d.%d", version.Major(), version.Minor())
		if accountRoles, err = m.createAccountRoles(majorMinor, channelGroup); err != nil {
			return "", err
		}

		createClusterArgs = append(createClusterArgs, "--controlplane-iam-role", accountRoles.ControlPlaneRoleARN)
		createClusterArgs = append(createClusterArgs, "--role-arn", accountRoles.InstallerRoleARN)
		createClusterArgs = append(createClusterArgs, "--support-role-arn", accountRoles.SupportRoleARN)
		createClusterArgs = append(createClusterArgs, "--worker-iam-role", accountRoles.WorkerRoleARN)
	}

	if viper.GetString(config.AWSVPCSubnetIDs) != "" {
		subnetIDs := viper.GetString(config.AWSVPCSubnetIDs)
		createClusterArgs = append(createClusterArgs, "--subnet-ids", subnetIDs)
		if viper.GetBool(config.Cluster.UseProxyForInstall) {
			if httpProxy := viper.GetString(config.Proxy.HttpProxy); httpProxy != "" {
				createClusterArgs = append(createClusterArgs, "--http-proxy", httpProxy)
			}
			if httpsProxy := viper.GetString(config.Proxy.HttpsProxy); httpsProxy != "" {
				createClusterArgs = append(createClusterArgs, "--https-proxy", httpsProxy)
			}
			if userCaBundle := viper.GetString(config.Proxy.UserCABundle); userCaBundle != "" {
				createClusterArgs = append(createClusterArgs, "--additional-trust-bundle-file", userCaBundle)
			}
		}
	} else if len(awsVPCSubnetIds) > 0 {
		createClusterArgs = append(createClusterArgs, "--subnet-ids", awsVPCSubnetIds)
	}

	if viper.GetBool(config.Cluster.MultiAZ) {
		createClusterArgs = append(createClusterArgs, "--multi-az")
	}
	// 3 minimum compute nodes are required for multi AZ. Osde2e default is 2 for all cluster types. Increase it unless greater is provided.
	if viper.GetBool(config.Cluster.MultiAZ) && viper.GetInt(Replicas) < 3 {
		createClusterArgs = append(createClusterArgs, "--replicas", "3")
	} else {
		createClusterArgs = append(createClusterArgs, "--replicas", viper.GetString(Replicas))
	}
	networkProvider := viper.GetString(config.Cluster.NetworkProvider)
	if networkProvider != config.DefaultNetworkProvider {
		createClusterArgs = append(createClusterArgs,
			"--network-type", networkProvider,
		)
	}

	if viper.GetBool(config.Hypershift) {
		createClusterArgs = append(createClusterArgs, "--hosted-cp")
		if viper.GetString(config.AWSVPCSubnetIDs) == "" && awsVPCSubnetIds == "" {
			return "", fmt.Errorf("ROSA hosted control plane requires a VPC.\n" +
				"You can BYOVPC by providing the subnet IDs using vault key 'subnet-ids'/set environment variable 'AWS_VPC_SUBNET_IDS' " +
				"or have osde2e create it by not supplying any subnet ids")
		}
		if rosaOIDCConfigID := viper.GetString(OIDCConfigID); rosaOIDCConfigID != "" {
			createClusterArgs = append(createClusterArgs, "--oidc-config-id", rosaOIDCConfigID)
		} else {
			oidcConfigID, err := m.createOIDCConfig(clusterName, accountRoles.InstallerRoleARN)
			if err != nil {
				return "", fmt.Errorf("failed to create oidc-config-id: %v", err)
			}
			createClusterArgs = append(createClusterArgs, "--oidc-config-id", oidcConfigID)
		}
	}

	if viper.GetBool(config.Cluster.EnableFips) {
		createClusterArgs = append(createClusterArgs, "--fips")
	}

	err = callAndSetAWSSession(func() error {
		// Retrieve AWS Account info
		logger := logging.NewLogger()

		awsClient, err := rosaAws.NewClient().
			Logger(logger).
			Region(m.awsRegion).
			Build()
		if err != nil {
			return err
		}

		awsCreator, err = awsClient.GetCreator()
		if err != nil {
			return fmt.Errorf("unable to get IAM credentials: %v", err)
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	if viper.GetBool(STS) {
		createClusterArgs = append(createClusterArgs, "--sts")
	}

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

	for k, v := range clusterProperties {
		createClusterArgs = append(createClusterArgs, "--properties", fmt.Sprintf("%s:%s", k, v))
	}

	// Filter out any create arguments that are blank.
	// This assumes `--arg ""` and will remove both elements
	for k, v := range createClusterArgs {
		if (k + 1) < len(createClusterArgs) {
			if len(strings.TrimSpace(createClusterArgs[k+1])) == 0 || createClusterArgs[k+1] == `""` {
				log.Printf("Pruning `%s` and `%s`", v, createClusterArgs[k+1])
				createClusterArgs = append(createClusterArgs[:k], createClusterArgs[k+1:]...)
			}
		}
	}

	log.Printf("%v", createClusterArgs)

	newCluster := createCluster.Cmd
	newCluster.SetArgs(createClusterArgs)
	err = callAndSetAWSSession(func() error {
		return newCluster.Execute()
	})
	if err != nil {
		log.Print("Error creating cluster: ", err)
		return "", err
	}

	ocmClient, err := m.ocmLogin()
	if err != nil {
		return "", err
	}

	cluster, err := ocmClient.GetCluster(clusterName, awsCreator)
	if err != nil {
		return "", fmt.Errorf("failed to get cluster '%s': %v", clusterName, err)
	}

	return cluster.ID(), nil
}

// DeleteCluster will call DeleteCluster from the OCM provider then delete
// additional AWS resources if STS is in use.
func (m *ROSAProvider) DeleteCluster(clusterID string) error {
	var reportDir *string
	if viper.GetString(config.AWSVPCSubnetIDs) == "" && viper.GetBool(config.Hypershift) {
		value, err := m.GetProperty(clusterID, "reportDir")
		if err != nil {
			return fmt.Errorf("unable to delete auto-generated vpc, failed to locate directory with terraform state file: %v", err)
		}
		reportDir = &value
	}

	_, err := m.ocmLogin()
	if err != nil {
		return err
	}

	if err := m.ocmProvider.DeleteCluster(clusterID); err != nil {
		return err
	}

	if viper.GetBool(STS) {
		err := m.stsClusterCleanup(clusterID)
		if err != nil {
			return err
		}
	}

	if viper.GetBool(config.Hypershift) {
		if viper.GetString(OIDCConfigID) == "" {
			if err := m.deleteOIDCConfig(viper.GetString(config.Cluster.Name)); err != nil {
				return err
			}
		}

		if viper.GetString(config.AWSVPCSubnetIDs) == "" {
			err := deleteHyperShiftVPC(*reportDir)
			if err != nil {
				return fmt.Errorf("error deleting aws vpc: %v", err)
			}
		}
	}

	return nil
}

func (m *ROSAProvider) stsClusterCleanup(clusterID string) error {
	response, err := m.ocmProvider.GetConnection().ClustersMgmt().V1().Clusters().List().
		Search(fmt.Sprintf("product.id = 'rosa' AND id = '%s'", clusterID)).
		Page(1).
		Size(1).
		Send()
	if response.Total() == 0 || err != nil {
		return fmt.Errorf("failed to locate cluster in ocm: %v", err)
	}

	cluster := response.Items().Slice()[0]
	operatorRolePrefix := cluster.AWS().STS().OperatorRolePrefix()
	oidcConfigID := cluster.AWS().STS().OidcConfig().ID()

	// wait for the cluster to no longer be available
	wait.PollImmediate(2*time.Minute, 30*time.Minute, func() (bool, error) {
		clusters, err := m.ocmProvider.ListClusters(fmt.Sprintf("id = '%s'", clusterID))
		if err != nil {
			return false, err
		}
		log.Printf("Waiting for cluster %s to be deleted", clusterID)
		return len(clusters) == 0, nil
	})

	return callAndSetAWSSession(func() error {
		var err error
		defaultArgs := []string{"--mode", "auto", "--yes"}

		deleteOperatorRolesArgs := append([]string{"operator-roles"}, defaultArgs...)

		if oidcConfigID != "" {
			deleteOperatorRolesArgs = append(deleteOperatorRolesArgs, "--prefix", operatorRolePrefix)
		} else {
			deleteOperatorRolesArgs = append(deleteOperatorRolesArgs, "--cluster", clusterID)
		}

		deleteOperatorRolesCmd := operatorrole.Cmd
		deleteOperatorRolesCmd.SetArgs(deleteOperatorRolesArgs)
		log.Printf("%v", deleteOperatorRolesArgs)

		deleteOIDCProviderArgs := append([]string{"oidc-provider"}, defaultArgs...)

		if oidcConfigID != "" {
			deleteOIDCProviderArgs = append(deleteOIDCProviderArgs, "--oidc-config-id", oidcConfigID)
		} else {
			deleteOIDCProviderArgs = append(deleteOIDCProviderArgs, "--cluster", clusterID)
		}

		deleteOIDCProviderCmd := oidcprovider.Cmd
		deleteOIDCProviderCmd.SetArgs(deleteOIDCProviderArgs)
		log.Printf("%v", deleteOIDCProviderArgs)

		if err = deleteOperatorRolesCmd.Execute(); err != nil {
			log.Printf("Error deleting operator roles: %v", err)
			return err
		}
		log.Printf("Deleted operator roles for cluster %s", clusterID)

		if err = deleteOIDCProviderCmd.Execute(); err != nil {
			log.Printf("Error deleting OIDC provider: %v", err)
			return err
		}
		log.Printf("Deleted OIDC provider for cluster %s", clusterID)

		return nil
	})
}

// DetermineRegion will return the region provided by configs. This mainly wraps the random functionality for use
// by the ROSA provider.
func (m *ROSAProvider) DetermineRegion(cloudProvider string) (string, error) {
	// If a region is set to "random", it will poll OCM for all the regions available
	// It then will pull a random entry from the list of regions and set the ID to that
	if m.awsRegion == "random" {
		var regions []*v1.CloudRegion
		// We support multiple cloud providers....
		if cloudProvider == "aws" {
			awsCredentials, err := v1.NewAWS().
				AccessKeyID(m.awsCredentials.AccessKeyID).
				SecretAccessKey(m.awsCredentials.SecretAccessKey).
				Build()
			if err != nil {
				return "", err
			}

			response, err := m.ocmProvider.GetConnection().ClustersMgmt().V1().CloudProviders().CloudProvider(cloudProvider).AvailableRegions().Search().Body(awsCredentials).Send()
			if err != nil {
				return "", err
			}
			regions = response.Items().Slice()
		}

		// But we don't support passing GCP credentials yet :)
		if cloudProvider == "gcp" {
			log.Println("Random GCP region not supported yet. Setting region to us-east1")
			return "us-east1", nil
		}

		cloudRegion, found := ChooseRandomRegion(toCloudRegions(regions)...)
		if !found {
			return "", fmt.Errorf("unable to choose a random enabled region")
		}

		m.awsRegion = cloudRegion.ID()

		log.Printf("Random region requested, selected %s region.", m.awsRegion)

		// Update the Config with the selected random region
		viper.Set(config.CloudProvider.Region, m.awsRegion)
		viper.Set(config.AWSRegion, m.awsRegion)
	}

	return m.awsRegion, nil
}

// Determine whether the region provided is supported for hosted control plane clusters
func (m *ROSAProvider) IsRegionValidForHCP() (bool, error) {
	err := callAndSetAWSSession(func() error {
		ocmClient, err := m.ocmLogin()
		if err != nil {
			return fmt.Errorf("failed to login to ocm: %w", err)
		}

		availableRegions, err := ocmClient.GetRegions("", "")
		if err != nil {
			return fmt.Errorf("failed to get regions: %w", err)
		}

		var supportedRegions []string

		for _, r := range availableRegions {
			if r.SupportsHypershift() {
				supportedRegions = append(supportedRegions, r.ID())
			}
		}

		for _, r := range supportedRegions {
			if m.awsRegion == r {
				return nil
			}
		}
		return fmt.Errorf("region '%s' does not support hosted-cp. valid regions '%s'", m.awsRegion, supportedRegions)
	})
	if err != nil {
		return false, err
	}

	return true, nil
}

// ChooseRandomRegion chooses a random enabled region from the provided options. Its
// second return parameter indicates whether it was successful in finding an enabled
// region.
func ChooseRandomRegion(regions ...CloudRegion) (CloudRegion, bool) {
	// remove disabled regions from consideration
	enabledRegions := make([]CloudRegion, 0, len(regions))
	for _, region := range regions {
		if region.Enabled() {
			enabledRegions = append(enabledRegions, region)
		}
	}
	// randomize the order of the candidates
	rand.Shuffle(len(enabledRegions), func(i, j int) {
		enabledRegions[i], enabledRegions[j] = enabledRegions[j], enabledRegions[i]
	})
	// return the first element if the list is not empty
	for _, regionObj := range enabledRegions {
		return regionObj, true
	}
	// indicate that there were no enabled candidates
	return nil, false
}

func (m *ROSAProvider) ocmLogin() (*ocm.Client, error) {
	err := os.Setenv("OCM_CONFIG", "/tmp/ocm.json")
	if err != nil {
		return nil, err
	}

	// URLAliases allows the value of the `--env` option to map to the various API URLs.
	URLAliases := map[string]string{
		"prod":  "https://api.openshift.com",
		"stage": "https://api.stage.openshift.com",
		"int":   "https://api.integration.openshift.com",
	}
	url, ok := URLAliases[viper.GetString(Env)]
	if !ok {
		url = URLAliases["prod"]
	}

	newLogin := rosaLogin.Cmd
	newLogin.SetArgs([]string{"--token", viper.GetString("ocm.token"), "--env", url})
	err = newLogin.Execute()
	if err != nil {
		return nil, fmt.Errorf("unable to login to OCM: %s", err.Error())
	}

	logger := logging.NewLogger()

	ocmClient, err := ocm.NewClient().Logger(logger).Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create OCM client: %s", err.Error())
	}

	log.Print("created OCM client")
	return ocmClient, err
}

// Versions will call Versions from the OCM provider.
func (m *ROSAProvider) Versions() (*spi.VersionList, error) {
	ocmClient, err := m.ocmLogin()
	if err != nil {
		return nil, err
	}

	versionResponse, err := ocmClient.GetVersions(viper.GetString(config.Cluster.Channel))
	if err != nil {
		return nil, err
	}

	spiVersions := []*spi.Version{}
	var defaultVersionOverride *semver.Version = nil

	for _, v := range versionResponse {
		if viper.GetBool(config.Hypershift) && !v.HostedControlPlaneEnabled() {
			continue
		} else if version, err := util.OpenshiftVersionToSemver(v.ID()); err != nil {
			log.Printf("could not parse version '%s': %v", v.ID(), err)
		} else if v.Enabled() {
			if v.Default() {
				defaultVersionOverride = version
			}
			spiVersion := spi.NewVersionBuilder().
				Version(version).
				Default(v.Default()).
				Build()

			for _, upgrade := range v.AvailableUpgrades() {
				if version, err := util.OpenshiftVersionToSemver(upgrade); err == nil {
					spiVersion.AddUpgradePath(version)
				}
			}

			spiVersions = append(spiVersions, spiVersion)
		}
	}

	versionList := spi.NewVersionListBuilder().
		AvailableVersions(spiVersions).
		DefaultVersionOverride(defaultVersionOverride).
		Build()

	return versionList, nil
}

// CloudRegion provides an interface for methods on *v1.CloudRegion so that
// compatible types can be instantiated from tests.
type CloudRegion interface {
	ID() string
	Enabled() bool
}

// ensure *v1.CloudRegion implements CloudRegion at compile time
var _ CloudRegion = &v1.CloudRegion{}

// toCloudRegions converts a slice of *v1.CloudRegion into a slice of CloudRegion.
// This helper can be removed once generics lands in Go, as this will no longer be
// necessary.
func toCloudRegions(in []*v1.CloudRegion) []CloudRegion {
	out := make([]CloudRegion, 0, len(in))
	for i := range in {
		out = append(out, in[i])
	}
	return out
}
