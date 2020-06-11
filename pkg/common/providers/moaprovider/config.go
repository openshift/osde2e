package moaprovider

import (
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/spf13/viper"
)

const (
	// Env is the OpenShift Dedicated environment used to provision clusters.
	Env = "moa.env"

	// AWSAccessKeyID for provisioning clusters.
	AWSAccessKeyID = "moa.awsAccessKey"

	// AWSSecretAccessKey for provisioning clusters.
	AWSSecretAccessKey = "moa.awsSecretAccessKey"

	// AWSRegion for provisioning clusters.
	AWSRegion = "moa.awsRegion"

	// MachineCIDR is the CIDR to use for machines.
	MachineCIDR = "moa.machineCIDR"

	// ServiceCIDR is the CIDR to use for services.
	ServiceCIDR = "moa.serviceCIDR"

	// PodCIDR is the CIDR to use for pods.
	PodCIDR = "moa.podCIDR"

	// ComputeMachineType is instance size of the compute nodes in a cluster.
	ComputeMachineType = "moa.computeMachineType"

	// ComputeNodes is number of compute nodes in a cluster.
	ComputeNodes = "moa.computeNodes"

	// HostPrefix is the prefix for the hosts produced by MOA.
	HostPrefix = "moa.hostPrefix"
)

func init() {
	// ----- MOA -----
	viper.SetDefault(Env, "prod")
	viper.BindEnv(Env, "MOA_ENV")

	viper.BindEnv(AWSAccessKeyID, "MOA_AWS_ACCESS_KEY_ID")
	config.RegisterSecret(AWSAccessKeyID, "moa-aws-access-key")

	viper.BindEnv(AWSSecretAccessKey, "MOA_AWS_SECRET_ACCESS_KEY")
	config.RegisterSecret(AWSSecretAccessKey, "moa-aws-secret-access-key")

	viper.BindEnv(AWSRegion, "MOA_AWS_REGION")
	config.RegisterSecret(AWSRegion, "moa-aws-region")

	viper.BindEnv(MachineCIDR, "MOA_MACHINE_CIDR")

	viper.BindEnv(ServiceCIDR, "MOA_SERVICE_CIDR")

	viper.BindEnv(PodCIDR, "MOA_POD_CIDR")

	viper.BindEnv(ComputeMachineType, "MOA_COMPUTE_MACHINE_TYPE")

	viper.BindEnv(ComputeNodes, "MOA_COMPUTE_NODES")

	viper.BindEnv(HostPrefix, "MOA_HOST_PREFIX")
}
