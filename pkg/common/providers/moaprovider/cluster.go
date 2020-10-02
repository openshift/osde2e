package moaprovider

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/Masterminds/semver"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift/moactl/pkg/cluster"
	"github.com/openshift/moactl/pkg/ocm/versions"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/util"
	"github.com/spf13/viper"
)

// LaunchCluster will provision an AWS cluster.
func (m *MOAProvider) LaunchCluster(clusterName string) (string, error) {
	clustersClient := m.ocmProvider.GetConnection().ClustersMgmt().V1().Clusters()

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
	clusterProperties["moa_use_marketplace_ami"] = "true"

	if err != nil {
		return "", fmt.Errorf("error generating cluster properties: %v", err)
	}

	var createdCluster *cmv1.Cluster

	// MOA uses the AWS provider in the background, so we'll determine region this way.
	region, err := m.ocmProvider.DetermineRegion("aws")

	if err != nil {
		return "", fmt.Errorf("error determining region to use: %v", err)
	}

	callAndSetAWSSession(func() {
		clusterSpec := cluster.Spec{
			Name:               clusterName,
			Region:             region,
			MultiAZ:            viper.GetBool(config.Cluster.MultiAZ),
			Version:            viper.GetString(config.Cluster.Version),
			Expiration:         expiration,
			ComputeMachineType: viper.GetString(ComputeMachineType),
			ComputeNodes:       viper.GetInt(ComputeNodes),

			CustomProperties: clusterProperties,
			MachineCIDR:      *machineCIDRParsed,
			ServiceCIDR:      *serviceCIDRParsed,
			PodCIDR:          *podCIDRParsed,
			HostPrefix:       viper.GetInt(HostPrefix),
		}

		createdCluster, err = cluster.CreateCluster(clustersClient, clusterSpec)
	})

	if err != nil {
		return "", fmt.Errorf("error while creating cluster: %v", err)
	}

	return createdCluster.ID(), nil
}

// Versions will call Versions from the OCM provider.
func (m *MOAProvider) Versions() (*spi.VersionList, error) {
	clustersClient := m.ocmProvider.GetConnection().ClustersMgmt().V1()
	versionResponse, err := versions.GetVersions(clustersClient, "")
	if err != nil {
		return nil, err
	}
	spiVersions := []*spi.Version{}
	var defaultVersionOverride *semver.Version = nil

	for _, v := range versionResponse {
		if version, err := util.OpenshiftVersionToSemver(v.ID()); err != nil {
			log.Printf("could not parse version '%s': %v", v.ID(), err)
		} else if v.Enabled() {
			if (m.Environment() == "stage" || m.Environment() == "prod") && v.ChannelGroup() != "stable" {
				continue
			}
			if v.Default() {
				defaultVersionOverride = version
			}
			spiVersions = append(spiVersions, spi.NewVersionBuilder().
				Version(version).
				Default(v.Default()).
				Build())
		}
	}

	versionList := spi.NewVersionListBuilder().
		AvailableVersions(spiVersions).
		DefaultVersionOverride(defaultVersionOverride).
		Build()

	return versionList, nil
}
