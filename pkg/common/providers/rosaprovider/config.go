package rosaprovider

import (
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/spf13/viper"
)

const (
	// Env is the OpenShift Dedicated environment used to provision clusters.
	Env = "rosa.env"

	// AWSAccessKeyID for provisioning clusters.
	AWSAccessKeyID = "rosa.awsAccessKey"

	// AWSSecretAccessKey for provisioning clusters.
	AWSSecretAccessKey = "rosa.awsSecretAccessKey"

	// AWSRegion for provisioning clusters.
	AWSRegion = "rosa.awsRegion"

	// MachineCIDR is the CIDR to use for machines.
	MachineCIDR = "rosa.machineCIDR"

	// ServiceCIDR is the CIDR to use for services.
	ServiceCIDR = "rosa.serviceCIDR"

	// PodCIDR is the CIDR to use for pods.
	PodCIDR = "rosa.podCIDR"

	// ComputeMachineType is instance size of the compute nodes in a cluster.
	ComputeMachineType = "rosa.computeMachineType"

	// ComputeNodes is number of compute nodes in a cluster.
	ComputeNodes = "rosa.computeNodes"

	// HostPrefix is the prefix for the hosts produced by ROSA.
	HostPrefix = "rosa.hostPrefix"
)

func init() {
	// ----- ROSA -----
	viper.SetDefault(Env, "prod")
	viper.BindEnv(Env, "ROSA_ENV")

	viper.BindEnv(AWSAccessKeyID, "ROSA_AWS_ACCESS_KEY_ID")
	config.RegisterSecret(AWSAccessKeyID, "rosa-aws-access-key")

	viper.BindEnv(AWSSecretAccessKey, "ROSA_AWS_SECRET_ACCESS_KEY")
	config.RegisterSecret(AWSSecretAccessKey, "rosa-aws-secret-access-key")

	viper.BindEnv(AWSRegion, "ROSA_AWS_REGION")
	config.RegisterSecret(AWSRegion, "rosa-aws-region")

	viper.BindEnv(MachineCIDR, "ROSA_MACHINE_CIDR")

	viper.BindEnv(ServiceCIDR, "ROSA_SERVICE_CIDR")

	viper.BindEnv(PodCIDR, "ROSA_POD_CIDR")

	viper.BindEnv(ComputeMachineType, "ROSA_COMPUTE_MACHINE_TYPE")

	viper.BindEnv(ComputeNodes, "ROSA_COMPUTE_NODES")

	viper.BindEnv(HostPrefix, "ROSA_HOST_PREFIX")
}
