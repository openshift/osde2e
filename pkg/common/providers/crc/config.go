package crc

import (
	"github.com/spf13/viper"
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