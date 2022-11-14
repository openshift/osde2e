package ocmprovider

import (
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
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

	// AWSAccount is used in CCS clusters
	AWSAccount = "ocm.aws.account"
	// AWSAccessKey is used in CCS clusters
	AWSAccessKey = "ocm.aws.accessKey"
	// AWSSecretKey is used in CCS clusters
	AWSSecretKey = "ocm.aws.secretKey"
	// AWSVPCSubnetIDs is used in CCS clusters
	AWSVPCSubnetIDs = "ocm.aws.vpcSubnetIDs"

	// GCP CCS Credentials
	GCPCredsJSON               = "ocm.gcp.credsJSON"
	GCPCredsType               = "ocm.gcp.credsType"
	GCPProjectID               = "ocm.gcp.projectID"
	GCPPrivateKey              = "ocm.gcp.privateKey"
	GCPPrivateKeyID            = "ocm.gcp.privateKeyID"
	GCPClientEmail             = "ocm.gcp.clientEmail"
	GCPClientID                = "ocm.gcp.clientID"
	GCPAuthURI                 = "ocm.gcp.authURI"
	GCPTokenURI                = "ocm.gcp.tokenURI"
	GCPAuthProviderX509CertURL = "ocm.gcp.authProviderX509CertURL"
	GCPClientX509CertURL       = "ocm.gcp.clientX509CertURL"
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

	viper.SetDefault(ComputeMachineTypeRegex, "")
	viper.BindEnv(ComputeMachineTypeRegex, "OCM_COMPUTE_MACHINE_TYPE_REGEX")

	viper.BindEnv(UserOverride, "OCM_USER_OVERRIDE")

	viper.SetDefault(Flavour, "osd-4")
	viper.BindEnv(Flavour, "OCM_FLAVOUR")

	viper.SetDefault(Sku, "")
	viper.BindEnv(Sku, "OCM_SKU")

	viper.BindEnv(AdditionalLabels, "OCM_ADDITIONAL_LABELS")

	viper.SetDefault(CCS, false)
	viper.BindEnv(CCS, "OCM_CCS", "CCS")

	viper.BindEnv(AWSAccount, "OCM_AWS_ACCOUNT", "AWS_ACCOUNT")
	viper.BindEnv(AWSAccessKey, "OCM_AWS_ACCESS_KEY", "AWS_ACCESS_KEY_ID")
	viper.BindEnv(AWSSecretKey, "OCM_AWS_SECRET_KEY", "AWS_SECRET_ACCESS_KEY")
	viper.BindEnv(AWSVPCSubnetIDs, "OCM_AWS_VPC_SUBNET_IDS")

	config.RegisterSecret(AWSAccessKey, "aws-access-key-id")
	config.RegisterSecret(AWSSecretKey, "aws-secret-access-key")

	config.RegisterSecret(AWSAccount, "ocm-aws-account")
	config.RegisterSecret(AWSAccessKey, "ocm-aws-access-key")
	config.RegisterSecret(AWSSecretKey, "ocm-aws-secret-access-key")

	viper.BindEnv(GCPCredsType, "OCM_GCP_CREDS_TYPE", "GCP_CREDS_TYPE")
	viper.BindEnv(GCPProjectID, "OCM_GCP_PROJECT_ID", "GCP_PROJECT_ID")
	viper.BindEnv(GCPPrivateKey, "OCM_GCP_PRIVATE_KEY", "GCP_PRIVATE_KEY")
	viper.BindEnv(GCPPrivateKeyID, "OCM_GCP_PRIVATE_KEY_ID", "GCP_PRIVATE_KEY_ID")
	viper.BindEnv(GCPClientEmail, "OCM_GCP_CLIENT_EMAIL", "GCP_CLIENT_EMAIL")
	viper.BindEnv(GCPClientID, "OCM_GCP_CLIENT_ID", "GCP_CLIENT_ID")
	viper.BindEnv(GCPAuthURI, "OCM_GCP_AUTH_URI", "GCP_AUTH_URI")
	viper.BindEnv(GCPTokenURI, "OCM_GCP_TOKEN_URI", "GCP_TOKEN_URI")
	viper.BindEnv(GCPAuthProviderX509CertURL, "OCM_GCP_AUTH_PROVIDER_X509_CERT_URL", "GCP_AUTH_PROVIDER_X509_CERT_URL")
	viper.BindEnv(GCPClientX509CertURL, "OCM_GCP_CLIENT_X509_CERT_URL", "GCP_CLIENT_X509_CERT_URL")

	config.RegisterSecret(GCPCredsJSON, "ocm-gcp-creds.json")
	config.RegisterSecret(GCPCredsJSON, "gcp-creds.json")

	config.RegisterSecret(GCPCredsType, "ocm-gcp-creds-type")
	config.RegisterSecret(GCPProjectID, "ocm-gcp-project-id")
	config.RegisterSecret(GCPPrivateKey, "ocm-gcp-private-key")
	config.RegisterSecret(GCPPrivateKeyID, "ocm-gcp-private-key-id")
	config.RegisterSecret(GCPClientEmail, "ocm-gcp-client-email")
	config.RegisterSecret(GCPClientID, "ocm-gcp-client-id")
	config.RegisterSecret(GCPAuthURI, "ocm-gcp-auth-uri")
	config.RegisterSecret(GCPTokenURI, "ocm-gcp-token-uri")
	config.RegisterSecret(GCPAuthProviderX509CertURL, "ocm-gcp-auth-provider-x509-cert-url")
	config.RegisterSecret(GCPClientX509CertURL, "ocm-gcp-client-x509-cert-url")

	config.RegisterSecret(GCPCredsType, "gcp-creds-type")
	config.RegisterSecret(GCPProjectID, "gcp-project-id")
	config.RegisterSecret(GCPPrivateKey, "gcp-private-key")
	config.RegisterSecret(GCPPrivateKeyID, "ocp-private-key-id")
	config.RegisterSecret(GCPClientEmail, "gcp-client-email")
	config.RegisterSecret(GCPClientID, "gcp-client-id")
	config.RegisterSecret(GCPAuthURI, "gcp-auth-uri")
	config.RegisterSecret(GCPTokenURI, "gcp-token-uri")
	config.RegisterSecret(GCPAuthProviderX509CertURL, "gcp-auth-provider-x509-cert-url")
	config.RegisterSecret(GCPClientX509CertURL, "gcp-client-x509-cert-url")
}
