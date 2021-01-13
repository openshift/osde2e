package rosaprovider

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
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

	env = os.Environ()
	os.Setenv("AWS_ACCESS_KEY_ID", viper.GetString(AWSAccessKeyID))
	os.Setenv("AWS_SECRET_ACCESS_KEY", viper.GetString(AWSSecretAccessKey))
	os.Setenv("AWS_REGION", viper.GetString(AWSRegion))
	error := false
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" {
		log.Println("AWS_ACCESS_KEY_ID is empty")
		error = true
	}
	if os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
		log.Println("AWS_SECRET_ACCESS_KEY is empty")
		error = true
	}
	if os.Getenv("AWS_REGION") == "" {
		log.Println("AWS_REGION is empty")
		error = true
	}

	if !error {
		return f()
	}

	return fmt.Errorf("one or more required AWS variables were not set")
}
