package ocmprovider

import (
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
)

func init() {
	// ----- OCM -----
	viper.BindEnv(Token, "OCM_TOKEN")

	viper.SetDefault(Env, "prod")
	viper.BindEnv(Env, "OSD_ENV")

	viper.SetDefault(Debug, false)
	viper.BindEnv(Debug, "DEBUG_OSD")

	viper.SetDefault(NumRetries, 3)
	viper.BindEnv(NumRetries, "NUM_RETRIES")
}
