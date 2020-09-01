package ocmprovider

import (
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/spf13/viper"
)

const (
	// Token is used to authenticate with OCM.
	Token = "ocm.token"

	// Env is the OpenShift Dedicated environment used to provision clusters.
	Env = "ocm.env"

	// Debug shows debug level messages when enabled.
	Debug = "ocm.debug"

	// NumRetries is the number of times to retry each OCM call.
	NumRetries = "ocm.numRetries"

	// ComputeMachineType is the specific cloud machine type to use for compute nodes.
	// This will accept a comma-delimited list. If said-list is present, it will select
	// an entry at random and use that machine type.
	ComputeMachineType = "ocm.computeMachineType"

	// UserOverride will hard set the user assigned to the "owner" tag by the OCM provider.
	UserOverride = "ocm.userOverride"
)

func init() {
	// ----- OCM -----
	viper.BindEnv(Token, "OCM_TOKEN")
	config.RegisterSecret(Token, "ocm-refresh-token")

	viper.SetDefault(Env, "prod")
	viper.BindEnv(Env, "OSD_ENV")

	viper.SetDefault(Debug, false)
	viper.BindEnv(Debug, "DEBUG_OSD")

	viper.SetDefault(NumRetries, 3)
	viper.BindEnv(NumRetries, "NUM_RETRIES")

	viper.SetDefault(ComputeMachineType, "")
	viper.BindEnv(ComputeMachineType, "OCM_COMPUTE_MACHINE_TYPE")

	viper.BindEnv(UserOverride, "OCM_USER_OVERRIDE")
}
