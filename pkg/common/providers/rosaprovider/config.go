package rosaprovider

import (
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
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

	// ComputeMachineType is instance size of the compute nodes in a cluster.
	ComputeMachineTypeRegex = "rosa.computeMachineTypeRegex"

	// Replicas is number of compute nodes in a cluster.
	Replicas = "rosa.replicas"

	// HostPrefix is the prefix for the hosts produced by ROSA.
	HostPrefix = "rosa.hostPrefix"

	// STS is a boolean tracking whether or not this cluster should be provisioned using the STS workflow
	STS = "rosa.STS"

	// SubnetIDs is comma-separated list of strings to specify the subnets for cluster provision
	SubnetIDs = "rosa.subnetIDs"
)

func init() {
	// ----- ROSA -----
	viper.SetDefault(Env, "prod")
	viper.BindEnv(Env, "ROSA_ENV")

	viper.BindEnv(AWSAccessKeyID, "ROSA_AWS_ACCESS_KEY_ID", "AWS_ACCESS_KEY_ID")
	config.RegisterSecret(AWSAccessKeyID, "rosa-aws-access-key")

	viper.BindEnv(AWSSecretAccessKey, "ROSA_AWS_SECRET_ACCESS_KEY", "AWS_SECRET_ACCESS_KEY")
	config.RegisterSecret(AWSSecretAccessKey, "rosa-aws-secret-access-key")

	viper.BindEnv(AWSRegion, "ROSA_AWS_REGION", "AWS_REGION")
	config.RegisterSecret(AWSRegion, "rosa-aws-region")

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

	viper.BindEnv(SubnetIDs, "ROSA_SUBNET_IDS", "SUBNET_IDS")
	config.RegisterSecret(SubnetIDs, "subnet-ids")
}
