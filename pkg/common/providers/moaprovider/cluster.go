package moaprovider

import (
	"fmt"
	"net"
	"time"

	"github.com/openshift/moactl/pkg/cluster"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/spf13/viper"
)

// LaunchCluster will provision an AWS cluster.
func (m *MOAProvider) LaunchCluster() (string, error) {
	clustersClient := m.ocmProvider.GetConnection().ClustersMgmt().V1().Clusters()

	// Calculate an expiration date for the cluster so that it will be automatically deleted if
	// we happen to forget to do it:
	expiration := time.Now().Add(viper.GetDuration(config.Cluster.ExpiryInMinutes) * time.Minute).UTC() // UTC() to workaround SDA-1567.

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

	clusterSpec := cluster.Spec{
		Name:               viper.GetString(config.Cluster.Name),
		Region:             viper.GetString(config.CloudProvider.Region),
		MultiAZ:            viper.GetBool(config.Cluster.MultiAZ),
		Version:            viper.GetString(config.Cluster.Version),
		Expiration:         expiration,
		ComputeMachineType: viper.GetString(ComputeMachineType),
		ComputeNodes:       viper.GetInt(ComputeNodes),

		MachineCIDR: *machineCIDRParsed,
		ServiceCIDR: *serviceCIDRParsed,
		PodCIDR:     *podCIDRParsed,
		HostPrefix:  viper.GetInt(HostPrefix),
	}

	createdCluster, err := cluster.CreateCluster(clustersClient, clusterSpec)

	if err != nil {
		return "", fmt.Errorf("error while creating cluster: %v", err)
	}

	return createdCluster.ID(), nil
}
