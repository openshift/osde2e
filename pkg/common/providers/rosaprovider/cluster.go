package rosaprovider

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	accountRoles "github.com/openshift/rosa/cmd/create/accountroles"
	createCluster "github.com/openshift/rosa/cmd/create/cluster"
	"github.com/openshift/rosa/cmd/dlt/oidcprovider"
	"github.com/openshift/rosa/cmd/dlt/operatorrole"
	rosaLogin "github.com/openshift/rosa/cmd/login"
	"github.com/openshift/rosa/pkg/aws"
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
	var awsCreator *aws.Creator

	rosaClusterVersion := viper.GetString(config.Cluster.Version)

	rosaClusterVersion = strings.Replace(rosaClusterVersion, "-fast", "", -1)
	rosaClusterVersion = strings.Replace(rosaClusterVersion, "-candidate", "", -1)
	if !strings.HasSuffix(rosaClusterVersion, "-nightly") {
		rosaClusterVersion = fmt.Sprintf("%s-%s", rosaClusterVersion, viper.GetString(config.Cluster.Channel))
	} else {
		viper.Set(config.Cluster.Channel, "nightly")
	}

	//Refactor: Was this not redundant with the above?
	rosaClusterVersion = strings.Replace(rosaClusterVersion, "-stable", "", -1)
	rosaClusterVersion = strings.Replace(rosaClusterVersion, "openshift-v", "", -1)

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
	region, err := m.DetermineRegion("aws")
	if err != nil {
		return "", fmt.Errorf("error determining region to use: %v", err)
	}

	createClusterArgs := []string{
		"--cluster-name", clusterName,
		"--region", region,
		"--channel-group", viper.GetString(config.Cluster.Channel),
		"--version", rosaClusterVersion,
		"--expiration-time", expiration.Format(time.RFC3339),
		"--compute-machine-type", viper.GetString(ComputeMachineType),
		"--replicas", viper.GetString(Replicas),
		"--machine-cidr", viper.GetString(MachineCIDR),
		"--service-cidr", viper.GetString(ServiceCIDR),
		"--pod-cidr", viper.GetString(PodCIDR),
		"--host-prefix", viper.GetString(HostPrefix),
	}
	if viper.GetString(SubnetIDs) != "" {
		subnetIDs := viper.GetString(SubnetIDs)
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
	}
	if viper.GetBool(config.Cluster.MultiAZ) {
		createClusterArgs = append(createClusterArgs, "--multi-az")
	}
	networkProvider := viper.GetString(config.Cluster.NetworkProvider)
	if networkProvider != config.DefaultNetworkProvider {
		createClusterArgs = append(createClusterArgs,
			"--network-type", networkProvider,
		)
	}

	awsAccountID := ""

	err = callAndSetAWSSession(func() error {
		// Retrieve AWS Account info
		logger := logging.NewLogger()

		awsClient, err := aws.NewClient().
			Logger(logger).
			Region(aws.DefaultRegion).
			Build()

		if err != nil {
			return err
		}

		awsCreator, err = awsClient.GetCreator()
		if err != nil {
			return fmt.Errorf("unable to get IAM credentials: %v", err)
		}

		awsAccountID = awsCreator.AccountID

		return nil
	})
	if err != nil {
		return "", err
	}

	if viper.GetBool(STS) {
		parsedVersion := semver.MustParse(rosaClusterVersion)
		majorMinor := fmt.Sprintf("%d.%d", parsedVersion.Major(), parsedVersion.Minor())

		err = m.stsAccountSetup(majorMinor)
		if err != nil {
			return "", err
		}
		createClusterArgs = append(createClusterArgs,
			"--role-arn", fmt.Sprintf("arn:aws:iam::%s:role/ManagedOpenShift-%s-Installer-Role", awsAccountID, majorMinor),
			"--support-role-arn", fmt.Sprintf("arn:aws:iam::%s:role/ManagedOpenShift-%s-Support-Role", awsAccountID, majorMinor),
			"--controlplane-iam-role", fmt.Sprintf("arn:aws:iam::%s:role/ManagedOpenShift-%s-ControlPlane-Role", awsAccountID, majorMinor),
			"--worker-iam-role", fmt.Sprintf("arn:aws:iam::%s:role/ManagedOpenShift-%s-Worker-Role", awsAccountID, majorMinor),
		)

	}

	clusterProperties, err := m.ocmProvider.GenerateProperties()

	if err != nil {
		return "", fmt.Errorf("error generating cluster properties: %v", err)
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
	if err := m.ocmProvider.DeleteCluster(clusterID); err != nil {
		return err
	}

	if viper.GetBool(STS) {
		return m.stsClusterCleanup(clusterID)
	}

	return nil
}

func (m *ROSAProvider) stsAccountSetup(version string) error {
	newAccountRoles := accountRoles.Cmd
	args := []string{"--version", version, "--prefix", fmt.Sprintf("ManagedOpenShift-%s", version), "--mode", "auto", "--yes"}
	log.Printf("%v", args)
	newAccountRoles.SetArgs(args)
	return callAndSetAWSSession(func() error {
		return newAccountRoles.Execute()
	})
}

func (m *ROSAProvider) stsClusterCleanup(clusterID string) error {
	// wait for the cluster to no longer be available
	wait.PollImmediate(3*time.Minute, 15*time.Minute, func() (bool, error) {
		clusters, err := m.ocmProvider.ListClusters(fmt.Sprintf("id = '%s'", clusterID))
		if err != nil {
			return false, err
		}
		log.Printf("Waiting for cluster %s to be deleted", clusterID)
		return len(clusters) == 0, nil
	})

	return callAndSetAWSSession(func() error {
		var err error
		defaultArgs := []string{"--cluster", clusterID, "--mode", "auto", "--yes"}

		deleteOperatorRolesArgs := append([]string{"operator-roles"}, defaultArgs...)
		deleteOperatorRolesCmd := operatorrole.Cmd
		deleteOperatorRolesCmd.SetArgs(deleteOperatorRolesArgs)
		log.Printf("%v", deleteOperatorRolesArgs)

		deleteOIDCProviderArgs := append([]string{"oidc-provider"}, defaultArgs...)
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
	region := viper.GetString(AWSRegion)

	// If a region is set to "random", it will poll OCM for all the regions available
	// It then will pull a random entry from the list of regions and set the ID to that
	if region == "random" {
		var regions []*v1.CloudRegion
		// We support multiple cloud providers....
		if cloudProvider == "aws" {
			if viper.GetString(AWSAccessKeyID) == "" || viper.GetString(AWSSecretAccessKey) == "" {
				log.Println("Random region requested but cloud credentials not supplied. Defaulting to us-east-1")
				return "us-east-1", nil
			}
			awsCredentials, err := v1.NewAWS().
				AccessKeyID(viper.GetString(AWSAccessKeyID)).
				SecretAccessKey(viper.GetString(AWSSecretAccessKey)).
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

		region = cloudRegion.ID()

		log.Printf("Random region requested, selected %s region.", region)

		// Update the Config with the selected random region
		viper.Set(config.CloudProvider.Region, region)
	}

	return region, nil
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
	var URLAliases = map[string]string{
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

	versionResponse, err := ocmClient.GetVersions("")
	if err != nil {
		return nil, err
	}

	if viper.GetString(config.Cluster.Channel) != "stable" {
		versionResponseChannel, err := ocmClient.GetVersions(viper.GetString(config.Cluster.Channel))
		if err != nil {
			return nil, err
		}
		versionResponse = append(versionResponse, versionResponseChannel...)
	}

	spiVersions := []*spi.Version{}
	var defaultVersionOverride *semver.Version = nil

	for _, v := range versionResponse {
		if version, err := util.OpenshiftVersionToSemver(v.ID()); err != nil {
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
