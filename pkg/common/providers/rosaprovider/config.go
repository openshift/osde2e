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

	// MintMode is a boolean tracking whether or not this cluster should be provisioned using the non-STS(Mint) workflow
	MintMode = "rosa.mintMode"

	// OIDCConfigID is the ID of the oidc-config created through ROSA CLI (required for HCP)
	OIDCConfigID = "rosa.oidcConfigID"

	// STSUseDefaultAccountRolesPrefix controls whether to use the default
	// "Managed-*" account roles or create unique roles based on the cluster
	// name
	STSUseDefaultAccountRolesPrefix = "rosa.stsUseDefaultAccountRolesPrefix"

	// BillingAccountID is the billing account ID used for Hosted Control Plane clusters.
	BillingAccountID = "rosa.billingAccountID"
)

func init() {
	// ----- ROSA -----
	viper.SetDefault(Env, "prod")
	_ = viper.BindEnv(Env, "ROSA_ENV")

	_ = viper.BindEnv(MachineCIDR, "ROSA_MACHINE_CIDR")

	_ = viper.BindEnv(ServiceCIDR, "ROSA_SERVICE_CIDR")

	_ = viper.BindEnv(PodCIDR, "ROSA_POD_CIDR")

	_ = viper.BindEnv(ComputeMachineType, "ROSA_COMPUTE_MACHINE_TYPE", "INSTANCE_TYPE")
	_ = viper.BindEnv(ComputeMachineTypeRegex, "ROSA_COMPUTE_MACHINE_TYPE", "INSTANCE_TYPE")

	_ = viper.BindEnv(Replicas, "ROSA_REPLICAS")
	viper.SetDefault(Replicas, 2)

	_ = viper.BindEnv(HostPrefix, "ROSA_HOST_PREFIX")
	viper.SetDefault(HostPrefix, 0)

	_ = viper.BindEnv(STS, "ROSA_STS")
	viper.SetDefault(STS, false)
	config.RegisterSecret(STS, "rosa-sts")

	_ = viper.BindEnv(MintMode, "ROSA_MINT_MODE")
	viper.SetDefault(MintMode, false)
	config.RegisterSecret(MintMode, "rosa-mint-mode")

	viper.SetDefault(OIDCConfigID, "")
	_ = viper.BindEnv(OIDCConfigID, "ROSA_OIDC_CONFIG_ID")
	config.RegisterSecret(OIDCConfigID, "rosa-oidc-config-id")

	_ = viper.BindEnv(STSUseDefaultAccountRolesPrefix, "ROSA_STS_USE_DEFAULT_ACCOUNT_ROLES_PREFIX")
	viper.SetDefault(STSUseDefaultAccountRolesPrefix, true)

	_ = viper.BindEnv(BillingAccountID, "ROSA_BILLING_ACCOUNT_ID")
	viper.SetDefault(BillingAccountID, "")
}
