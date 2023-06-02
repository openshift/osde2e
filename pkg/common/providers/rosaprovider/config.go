package rosaprovider

import (
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
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

	// Replicas is number of compute nodes in a cluster.
	Replicas = "rosa.replicas"

	// HostPrefix is the prefix for the hosts produced by ROSA.
	HostPrefix = "rosa.hostPrefix"

	// STS is a boolean tracking whether or not this cluster should be provisioned using the STS workflow
	STS = "rosa.STS"

	// PrivateLink is a boolean var to be specified for ROSA privatelink cluster. If this is true, a private subnet ID as ROSA_SUBNET_IDS must be provided.
	PrivateLink = "rosa.privateLink"

	// OIDCConfigID is the ID of the oidc-config created through ROSA CLI (required for HCP)
	OIDCConfigID = "rosa.oidcConfigID"
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

	viper.BindEnv(Replicas, "ROSA_REPLICAS")
	viper.SetDefault(Replicas, 2)

	viper.BindEnv(HostPrefix, "ROSA_HOST_PREFIX")
	viper.SetDefault(HostPrefix, 0)

	viper.BindEnv(STS, "ROSA_STS")
	viper.SetDefault(STS, false)

	viper.BindEnv(PrivateLink, "ROSA_PRIVATELINK")
	viper.SetDefault(PrivateLink, false)

	viper.SetDefault(OIDCConfigID, "")
	viper.BindEnv(OIDCConfigID, "ROSA_OIDC_CONFIG_ID")
	config.RegisterSecret(OIDCConfigID, "rosa-oidc-config-id")
}
