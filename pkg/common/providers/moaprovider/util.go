package moaprovider

import (
	"os"
	"strings"

	"github.com/spf13/viper"
)

// At the moment, moactl requires AWS sessions to be set globally. To get around that, we'll use this
// helper method here so that we can set environment variables and restore them before returning from the function.
func callAndSetAWSSession(f func()) {
	var env []string
	defer func() {
		os.Clearenv()
		for _, envVar := range env {
			keyAndValue := strings.SplitN(envVar, "=", 2)
			os.Setenv(keyAndValue[0], keyAndValue[1])
		}
	}()

	env = os.Environ()
	os.Setenv("AWS_ACCESS_KEY_ID", viper.GetString(AWSAccessKeyID))
	os.Setenv("AWS_SECRET_ACCESS_KEY", viper.GetString(AWSSecretAccessKey))
	os.Setenv("AWS_REGION", viper.GetString(AWSRegion))

	f()
}
