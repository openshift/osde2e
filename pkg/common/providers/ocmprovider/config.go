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
	ComputeMachineType = "ocm.computeMachineType"

	// UserOverride will hard set the user assigned to the "owner" tag by the OCM provider.
	UserOverride = "ocm.userOverride"

	// Flavour is an OCM cluster descriptor for cluster defaults
	Flavour = "ocm.flavour"

	// AdditionalLabels is used to add more specific labels to a cluster in OCM.
	AdditionalLabels = "ocm.additionalLabels"

	// CCS defines whether the cluster should expect cloud credentials or not
	CCS = "ocm.ccs"

	// AWSAccount is used in CCS clusters
	AWSAccount = "ocm.aws.account"
	// AWSAccessKey is used in CCS clusters
	AWSAccessKey = "ocm.aws.accessKey"
	// AWSSecretKey is used in CCS clusters
	AWSSecretKey = "ocm.aws.secretKey"
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

	viper.SetDefault(Flavour, "osd-4")
	viper.BindEnv(Flavour, "OCM_FLAVOUR")

	viper.BindEnv(AdditionalLabels, "OCM_ADDITIONAL_LABELS")

	viper.SetDefault(CCS, false)
	viper.BindEnv(CCS, "OCM_CCS")

	viper.BindEnv(AWSAccount, "OCM_AWS_ACCOUNT")
	viper.BindEnv(AWSAccessKey, "OCM_AWS_ACCESS_KEY")
	viper.BindEnv(AWSSecretKey, "OCM_AWS_SECRET_KEY")

	config.RegisterSecret(AWSAccount, "ocm-aws-account")
	config.RegisterSecret(AWSAccessKey, "ocm-aws-access-key")
	config.RegisterSecret(AWSSecretKey, "ocm-aws-secret-access-key")
}
