package rosaprovider

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/util"
	rosaLogin "github.com/openshift/rosa/cmd/login"
	"github.com/openshift/rosa/pkg/logging"
	"github.com/openshift/rosa/pkg/ocm"
	rprtr "github.com/openshift/rosa/pkg/reporter"
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
func (m *ROSAProvider) LaunchCluster(clusterName string) (string, error) {
	// Calculate an expiration date for the cluster so that it will be automatically deleted if
	// we happen to forget to do it:
	var expiration time.Time
	var err error

	ocmClient, err := m.ocmLogin()
	if err != nil {
		return "", err
	}

	expiryInMinutes := viper.GetDuration(config.Cluster.ExpiryInMinutes)
	if expiryInMinutes > 0 {
		expiration = time.Now().Add(expiryInMinutes * time.Minute).UTC() // UTC() to workaround SDA-1567.
	}

	var machineCIDRParsed = &net.IPNet{}
	machineCIDRString := viper.GetString(MachineCIDR)
	if machineCIDRString != "" {
		_, machineCIDRParsed, err = net.ParseCIDR(machineCIDRString)

		if err != nil {
			return "", fmt.Errorf("error while parsing machine CIDR: %v", err)
		}
	}

	var serviceCIDRParsed = &net.IPNet{}
	serviceCIDRString := viper.GetString(ServiceCIDR)
	if serviceCIDRString != "" {
		_, serviceCIDRParsed, err = net.ParseCIDR(ServiceCIDR)

		if err != nil {
			return "", fmt.Errorf("error while parsing service CIDR: %v", err)
		}
	}

	var podCIDRParsed = &net.IPNet{}
	podCIDRString := viper.GetString(PodCIDR)
	if podCIDRString != "" {
		_, podCIDRParsed, err = net.ParseCIDR(podCIDRString)

		if err != nil {
			return "", fmt.Errorf("error while parsing pod CIDR: %v", err)
		}
	}

	clusterProperties, err := m.ocmProvider.GenerateProperties()

	if err != nil {
		return "", fmt.Errorf("error generating cluster properties: %v", err)
	}

	installConfig := ""

	// This skips setting install_config for any prod job OR any periodic addon job.
	// To invoke this logic locally you will have to set JOB_TYPE to "periodic".
	if m.Environment() != "prod" {
		if os.Getenv("JOB_TYPE") == "periodic" && !strings.Contains(os.Getenv("JOB_NAME"), "addon") {
			imageSource := viper.GetString(config.Cluster.ImageContentSource)
			installConfig += "\n" + m.ChooseImageSource(imageSource)
			installConfig += "\n" + m.GetNetworkConfig(viper.GetString(config.Cluster.NetworkProvider))
		}
	}

	if viper.GetString(config.Cluster.InstallConfig) != "" {
		installConfig += "\n" + viper.GetString(config.Cluster.InstallConfig)
	}

	if installConfig != "" {
		log.Println("Install config:", installConfig)
		clusterProperties["install_config"] = installConfig
	}

	var createdCluster *cmv1.Cluster

	// ROSA uses the AWS provider in the background, so we'll determine region this way.
	region, err := m.DetermineRegion("aws")
	if err != nil {
		return "", fmt.Errorf("error determining region to use: %v", err)
	}
	falseValue := false

	rosaClusterVersion := viper.GetString(config.Cluster.Version)

	rosaClusterVersion = strings.Replace(rosaClusterVersion, "-fast", "", -1)
	rosaClusterVersion = strings.Replace(rosaClusterVersion, "-candidate", "", -1)
	if !strings.HasSuffix(rosaClusterVersion, "-nightly") {
		rosaClusterVersion = fmt.Sprintf("%s-%s", rosaClusterVersion, viper.GetString(config.Cluster.Channel))
	} else {
		viper.Set(config.Cluster.Channel, "nightly")
	}
	rosaClusterVersion = strings.Replace(rosaClusterVersion, "-stable", "", -1)

	log.Printf("ROSA cluster version: %s", rosaClusterVersion)

	clusterSpec := ocm.Spec{
		Name:               clusterName,
		Region:             region,
		ChannelGroup:       viper.GetString(config.Cluster.Channel),
		MultiAZ:            viper.GetBool(config.Cluster.MultiAZ),
		Version:            rosaClusterVersion,
		Expiration:         expiration,
		ComputeMachineType: viper.GetString(ComputeMachineType),
		ComputeNodes:       viper.GetInt(ComputeNodes),
		DryRun:             &falseValue,

		CustomProperties:  clusterProperties,
		MachineCIDR:       *machineCIDRParsed,
		ServiceCIDR:       *serviceCIDRParsed,
		PodCIDR:           *podCIDRParsed,
		HostPrefix:        viper.GetInt(HostPrefix),
		SubnetIds:         []string{},
		AvailabilityZones: []string{},
	}

	if viper.GetBool(config.Cluster.UseExistingCluster) && viper.GetString(config.Addons.IDs) == "" {
		if clusterID := m.ocmProvider.FindRecycledCluster(clusterSpec.Version, "aws", "rosa"); clusterID != "" {
			return clusterID, nil
		}
	}

	err = callAndSetAWSSession(func() error {
		createdCluster, err = ocmClient.CreateCluster(clusterSpec)
		if err != nil {
			return fmt.Errorf("error creating cluster: %s", err.Error())
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	return createdCluster.ID(), nil
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

	reporter := rprtr.CreateReporterOrExit()
	logger := logging.CreateLoggerOrExit(reporter)

	ocmClient, err := ocm.NewClient().Logger(logger).Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create OCM client: %s", err.Error())
	}

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

func (m *ROSAProvider) GetNetworkConfig(networkProvider string) string {
	if networkProvider == config.DefaultNetworkProvider {
		return ""
	}
	if networkProvider != "OVNKubernetes" {
		return ""
	}
	return `
networking:
  clusterNetwork:
  - cidr: 10.128.0.0/14
    hostPrefix: 23
  machineCIDR: 10.0.0.0/16
  machineNetwork:
  - cidr: 10.0.0.0/16
  networkType: OVNKubernetes
  serviceNetwork:
  - 172.30.0.0/16
`
}
