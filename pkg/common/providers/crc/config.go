package crc

import (
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
)

const (
	// CRCPullSecretFile is a file containing your pull secret
	CRCPullSecretFile = "crc.pull_secret_file"

	// CRCPullSecret is a string containing your pull secret
	CRCPullSecret = "crc.pull_secret"
)

func init() {
	// ----- CRC -----

	viper.BindEnv(CRCPullSecretFile, "CRC_PULL_SECRET_FILE")

	viper.BindEnv(CRCPullSecret, "CRC_PULL_SECRET")
}
