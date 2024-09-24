// GCP config provides the gcp configuration for tests run as part of the osde2e suite.
package config

import viper "github.com/openshift/osde2e/pkg/common/concurrentviper"

var (
	// GCPCredsJSON GCP CCS Credential json
	// Env: GCP_CREDS_JSON
	GCPCredsJSON = "gcp.credsJSON"
	// GCPCredsType GCP creds json internals
	GCPCredsType               = "gcp.credsType"
	GCPProjectID               = "gcp.projectID"
	GCPPrivateKey              = "gcp.privateKey"
	GCPPrivateKeyID            = "gcp.privateKeyID"
	GCPClientEmail             = "gcp.clientEmail"
	GCPClientID                = "gcp.clientID"
	GCPAuthURI                 = "gcp.authURI"
	GCPTokenURI                = "gcp.tokenURI"
	GCPAuthProviderX509CertURL = "gcp.authProviderX509CertURL"
	GCPClientX509CertURL       = "gcp.clientX509CertURL"
)

func InitGCPViper() {
	_ = viper.BindEnv(GCPCredsJSON, "GCP_CREDS_JSON")
	_ = viper.BindEnv(GCPCredsType, "GCP_CREDS_TYPE")
	_ = viper.BindEnv(GCPProjectID, "GCP_PROJECT_ID")
	_ = viper.BindEnv(GCPPrivateKey, "GCP_PRIVATE_KEY")
	_ = viper.BindEnv(GCPPrivateKeyID, "GCP_PRIVATE_KEY_ID")
	_ = viper.BindEnv(GCPClientEmail, "GCP_CLIENT_EMAIL")
	_ = viper.BindEnv(GCPClientID, "GCP_CLIENT_ID")
	_ = viper.BindEnv(GCPAuthURI, "GCP_AUTH_URI")
	_ = viper.BindEnv(GCPTokenURI, "GCP_TOKEN_URI")
	_ = viper.BindEnv(GCPAuthProviderX509CertURL, "GCP_AUTH_PROVIDER_X509_CERT_URL")
	_ = viper.BindEnv(GCPClientX509CertURL, "GCP_CLIENT_X509_CERT_URL")

	RegisterSecret(GCPCredsJSON, "gcp-creds.json")

	RegisterSecret(GCPCredsType, "gcp-creds-type")
	RegisterSecret(GCPProjectID, "gcp-project-id")
	RegisterSecret(GCPPrivateKey, "gcp-private-key")
	RegisterSecret(GCPPrivateKeyID, "gcp-private-key-id")
	RegisterSecret(GCPClientEmail, "gcp-client-email")
	RegisterSecret(GCPClientID, "gcp-client-id")
	RegisterSecret(GCPAuthURI, "gcp-auth-uri")
	RegisterSecret(GCPTokenURI, "gcp-token-uri")
	RegisterSecret(GCPAuthProviderX509CertURL, "gcp-auth-provider-x509-cert-url")
	RegisterSecret(GCPClientX509CertURL, "gcp-client-x509-cert-url")
}
