package ocmprovider

import (
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
)

const (
	// Token is used to authenticate with OCM.
	Token = "ocm.token"

	// ClientID is used to authenticate with OCM.
	ClientID = "ocm.clientID"

	// ClientSecret is used to authenticate with OCM.
	ClientSecret = "ocm.clientSecret"

	// Env is the OpenShift Dedicated environment used to provision clusters.
	Env = "ocm.env"

	// Debug shows debug level messages when enabled.
	Debug = "ocm.debug"

	// NumRetries is the number of times to retry each OCM call.
	NumRetries = "ocm.numRetries"

	// ComputeMachineType is the specific cloud machine type to use for compute nodes.
	ComputeMachineType = "ocm.computeMachineType"

	// ComputeMachineTypeRegex is the regex for cloud machine type to use for compute nodes.
	ComputeMachineTypeRegex = "ocm.computeMachineTypeRegex"

	// UserOverride will hard set the user assigned to the "owner" tag by the OCM provider.
	UserOverride = "ocm.userOverride"

	// Flavour is an OCM cluster descriptor for cluster defaults
	Flavour = "ocm.flavour"

	// SKU Rule ID is an identifier for a SKU that OCM can provision
	Sku = "ocm.skuRule"

	// AdditionalLabels is used to add more specific labels to a cluster in OCM.
	AdditionalLabels = "ocm.additionalLabels"

	// CCS defines whether the cluster should expect cloud credentials or not
	CCS = "ocm.ccs"

	// HTTPS_PROXY - Currently only used for FedRamp
	HTTPSProxy = "ocm.https_proxy"
)

func init() {
	// ----- OCM -----
	_ = viper.BindEnv(Token, "OCM_TOKEN")
	config.RegisterSecret(Token, "ocm-refresh-token")

	_ = viper.BindEnv(ClientID, "OCM_CLIENT_ID")
	config.RegisterSecret(ClientID, "ocm-client-id")

	_ = viper.BindEnv(ClientSecret, "OCM_CLIENT_SECRET")
	config.RegisterSecret(ClientSecret, "ocm-client-secret")

	viper.SetDefault(Env, "stage")
	_ = viper.BindEnv(Env, "OSD_ENV")

	viper.SetDefault(Debug, false)
	_ = viper.BindEnv(Debug, "DEBUG_OSD")

	viper.SetDefault(NumRetries, 3)
	_ = viper.BindEnv(NumRetries, "NUM_RETRIES")

	viper.SetDefault(ComputeMachineType, "")
	_ = viper.BindEnv(ComputeMachineType, "OCM_COMPUTE_MACHINE_TYPE", "INSTANCE_TYPE")

	viper.SetDefault(ComputeMachineTypeRegex, "")
	_ = viper.BindEnv(ComputeMachineTypeRegex, "OCM_COMPUTE_MACHINE_TYPE_REGEX", "INSTANCE_TYPE")

	_ = viper.BindEnv(UserOverride, "OCM_USER_OVERRIDE")

	viper.SetDefault(Flavour, "osd-4")
	_ = viper.BindEnv(Flavour, "OCM_FLAVOUR")

	viper.SetDefault(Sku, "")
	_ = viper.BindEnv(Sku, "OCM_SKU")

	_ = viper.BindEnv(AdditionalLabels, "OCM_ADDITIONAL_LABELS")

	viper.SetDefault(CCS, false)
	_ = viper.BindEnv(CCS, "OCM_CCS", "CCS")

	// ----- FedRamp -----
	viper.SetDefault(HTTPSProxy, "")
	_ = viper.BindEnv(HTTPSProxy, "HTTPS_PROXY")

	config.RegisterSecret(HTTPSProxy, "https-proxy")
}
