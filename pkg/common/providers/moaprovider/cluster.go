package moaprovider

import (
	"fmt"
	"net"
	"time"

	"github.com/openshift/moactl/pkg/cluster"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/state"
)

// LaunchCluster will provision an AWS cluster.
func (m *MOAProvider) LaunchCluster() (string, error) {
	cfg := config.Instance
	state := state.Instance
	clustersClient := m.ocmProvider.GetConnection().ClustersMgmt().V1().Clusters()

	// Calculate an expiration date for the cluster so that it will be automatically deleted if
	// we happen to forget to do it:
	expiration := time.Now().Add(time.Duration(cfg.Cluster.ExpiryInMinutes) * time.Minute).UTC() // UTC() to workaround SDA-1567.

	var err error
	var machineCIDR = &net.IPNet{}
	if cfg.MOA.MachineCIDR != "" {
		_, machineCIDR, err = net.ParseCIDR(cfg.MOA.MachineCIDR)

		if err != nil {
			return "", fmt.Errorf("error while parsing machine CIDR: %v", err)
		}
	}

	var serviceCIDR = &net.IPNet{}
	if cfg.MOA.ServiceCIDR != "" {
		_, serviceCIDR, err = net.ParseCIDR(cfg.MOA.ServiceCIDR)

		if err != nil {
			return "", fmt.Errorf("error while parsing service CIDR: %v", err)
		}
	}

	var podCIDR = &net.IPNet{}
	if cfg.MOA.PodCIDR != "" {
		_, podCIDR, err = net.ParseCIDR(cfg.MOA.PodCIDR)

		if err != nil {
			return "", fmt.Errorf("error while parsing pod CIDR: %v", err)
		}
	}

	clusterSpec := cluster.ClusterSpec{
		Name:               state.Cluster.Name,
		Region:             state.CloudProvider.Region,
		MultiAZ:            cfg.Cluster.MultiAZ,
		Version:            state.Cluster.Version,
		Expiration:         expiration,
		ComputeMachineType: cfg.MOA.ComputeMachineType,
		ComputeNodes:       cfg.MOA.ComputeNodes,

		MachineCIDR: *machineCIDR,
		ServiceCIDR: *serviceCIDR,
		PodCIDR:     *podCIDR,
		HostPrefix:  cfg.MOA.HostPrefix,
	}

	createdCluster, err := cluster.CreateCluster(clustersClient, clusterSpec)

	if err != nil {
		return "", fmt.Errorf("error while creating cluster: %v", err)
	}

	return createdCluster.ID(), nil
}
