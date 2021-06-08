package rosaprovider

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/util"
	"github.com/openshift/rosa/pkg/cluster"
	"github.com/openshift/rosa/pkg/ocm/versions"
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
			return false, fmt.Errorf("Can't retrieve page %d: %s\n", page, err)
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

	expiryInMinutes := viper.GetDuration(config.Cluster.ExpiryInMinutes)
	if expiryInMinutes > 0 {
		expiration = time.Now().Add(expiryInMinutes * time.Minute).UTC() // UTC() to workaround SDA-1567.
	}

	var err error
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

	var createdCluster *cmv1.Cluster

	// ROSA uses the AWS provider in the background, so we'll determine region this way.
	region, err := m.ocmProvider.DetermineRegion("aws")

	if err != nil {
		return "", fmt.Errorf("error determining region to use: %v", err)
	}
	falseValue := false

	clustersClient := m.ocmProvider.GetConnection().ClustersMgmt().V1().Clusters()

	rosaClusterVersion := strings.Replace(viper.GetString(config.Cluster.Version), "-fast", "", -1)
	rosaClusterVersion = strings.Replace(rosaClusterVersion, "-candidate", "", -1)
	rosaClusterVersion = fmt.Sprintf("%s-%s", rosaClusterVersion, viper.GetString(config.Cluster.Channel))
	rosaClusterVersion = strings.Replace(rosaClusterVersion, "-stable", "", -1)
	log.Printf("ROSA cluster version: %s", rosaClusterVersion)

	clusterSpec := cluster.Spec{
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
		createdCluster, err = cluster.CreateCluster(clustersClient, clusterSpec)
		if err != nil {
			return fmt.Errorf("Error creating cluster: %s", err.Error())
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	return createdCluster.ID(), nil
}

// Versions will call Versions from the OCM provider.
func (m *ROSAProvider) Versions() (*spi.VersionList, error) {
	clustersClient := m.ocmProvider.GetConnection().ClustersMgmt().V1()
	versionResponse, err := versions.GetVersions(clustersClient, "")
	if err != nil {
		return nil, err
	}

	if viper.GetString(config.Cluster.Channel) != "stable" {
		versionResponseChannel, err := versions.GetVersions(clustersClient, viper.GetString(config.Cluster.Channel))
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
