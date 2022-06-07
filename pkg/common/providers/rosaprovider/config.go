package rosaprovider

import (
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
)

const (
	// Env is the OpenShift Dedicated environment used to provision clusters.
	Env = "rosa.env"

	// MachineCIDR is the CIDR to use for machines.
	MachineCIDR = "rosa.machineCIDR"

	// ServiceCIDR is the CIDR to use for services.
	ServiceCIDR = "rosa.serviceCIDR"

	// PodCIDR is the CIDR to use for pods.
	PodCIDR = "rosa.podCIDR"

	// ComputeMachineType is instance size of the compute nodes in a cluster.
	ComputeMachineType = "rosa.computeMachineType"

	// ComputeMachineType is instance size of the compute nodes in a cluster.
	ComputeMachineTypeRegex = "rosa.computeMachineTypeRegex"

	// ComputeNodes is number of compute nodes in a cluster.
	ComputeNodes = "rosa.computeNodes"

	// HostPrefix is the prefix for the hosts produced by ROSA.
	HostPrefix = "rosa.hostPrefix"

	// STS is a boolean tracking whether or not this cluster should be provisioned using the STS workflow
	STS = "rosa.STS"
)

func init() {
	// ----- ROSA -----
	viper.SetDefault(Env, "prod")
	viper.BindEnv(Env, "ROSA_ENV")

	viper.BindEnv(MachineCIDR, "ROSA_MACHINE_CIDR")

	viper.BindEnv(ServiceCIDR, "ROSA_SERVICE_CIDR")

	viper.BindEnv(PodCIDR, "ROSA_POD_CIDR")

	viper.BindEnv(ComputeMachineType, "ROSA_COMPUTE_MACHINE_TYPE")
	viper.BindEnv(ComputeMachineTypeRegex, "ROSA_COMPUTE_MACHINE_TYPE")

	viper.BindEnv(ComputeNodes, "ROSA_COMPUTE_NODES")
	viper.SetDefault(ComputeNodes, 2)

	viper.BindEnv(HostPrefix, "ROSA_HOST_PREFIX")
	viper.SetDefault(HostPrefix, 0)

	viper.BindEnv(STS, "ROSA_STS")
	viper.SetDefault(STS, false)
}
