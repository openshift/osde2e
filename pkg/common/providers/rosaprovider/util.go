package rosaprovider

import (
	"fmt"
	"log"
	"os"
	"strings"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
)

// At the moment, rosa requires AWS sessions to be set globally. To get around that, we'll use this
// helper method here so that we can set environment variables and restore them before returning from the function.
func callAndSetAWSSession(f func() error) error {
	var env []string
	defer func() {
		os.Clearenv()
		for _, envVar := range env {
			keyAndValue := strings.SplitN(envVar, "=", 2)
			os.Setenv(keyAndValue[0], keyAndValue[1])
		}
	}()

	envVarCheck := func(envVars map[string]string) bool {
		error := false
		for key, value := range envVars {
			os.Setenv(key, viper.GetString(value))

			if os.Getenv(key) == "" {
				log.Printf("%s is not set", key)
				error = true
			}
		}
		return error
	}

	env = os.Environ()

	accessKeyError := envVarCheck(
		map[string]string{
			"AWS_ACCESS_KEY_ID":     config.AWSAccessKey,
			"AWS_SECRET_ACCESS_KEY": config.AWSSecretAccessKey,
		},
	)
	regionError := envVarCheck(
		map[string]string{"AWS_REGION": config.AWSRegion},
	)
	profileError := envVarCheck(
		map[string]string{"AWS_PROFILE": config.AWSProfile},
	)

	if (!accessKeyError && !regionError) || (!profileError && !regionError) {
		return f()
	}

	return fmt.Errorf("aws variables were not set (access key id, secret access key, region) or (aws profile, region)")
}
