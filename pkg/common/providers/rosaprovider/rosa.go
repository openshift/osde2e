// Package rosaprovider will allow the provisioning of clusters through rosa.
package rosaprovider

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	ocmclient "github.com/openshift/osde2e-common/pkg/clients/ocm"
	rosaprovider "github.com/openshift/osde2e-common/pkg/openshift/rosa"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"

	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
	"github.com/openshift/osde2e/pkg/common/spi"
	"k8s.io/klog/v2/textlogger"
)

var rosaProvider *rosaprovider.Provider

func init() {
	spi.RegisterProvider("rosa", func() (spi.Provider, error) { return New() })
}

// ROSAProvider will provision clusters via ROSA.
type ROSAProvider struct {
	ocmProvider      *ocmprovider.OCMProvider
	provider         *rosaprovider.Provider
	awsRegion        string
	versionGateLabel string
}

// New will create a new ROSAProvider.
func New() (*ROSAProvider, error) {
	fedramp := viper.GetBool(config.Cluster.FedRamp)
	rosaEnv := viper.GetString(Env)
	var ocmEnv ocmclient.Environment

	ocmProvider, err := ocmprovider.NewWithEnv(rosaEnv)
	if err != nil {
		return nil, fmt.Errorf("error creating OCM provider for ROSA provider: %v", err)
	}

	switch rosaEnv {
	case "prod":
		if fedramp {
			ocmEnv = ocmclient.FedRampProduction
		} else {
			ocmEnv = ocmclient.Production
		}
	case "stage":
		if fedramp {
			ocmEnv = ocmclient.FedRampStage
		} else {
			ocmEnv = ocmclient.Stage
		}
	case "int":
		if fedramp {
			ocmEnv = ocmclient.FedRampIntegration
		} else {
			ocmEnv = ocmclient.Integration
		}
	default:
		return nil, fmt.Errorf("error selecting ocm environment for %s", rosaEnv)
	}

	if rosaProvider == nil {
		// TODO: Revisit logger
		err = callAndSetAWSSession(func() error {
			ctx := context.Background()
			rosaProvider, err = rosaprovider.New(
				ctx,
				viper.GetString("ocm.token"),
				viper.GetString("ocm.clientID"),
				viper.GetString("ocm.clientSecret"),
				ocmEnv,
				textlogger.NewLogger(textlogger.NewConfig()),
			)
			return err
		})
		if err != nil {
			return nil, err
		}
	}

	versionGateLabel := "api.openshift.com/gate-ocp"
	if viper.GetBool(STS) {
		versionGateLabel = "api.openshift.com/gate-sts"
	}

	viper.Set(config.CloudProvider.Region, rosaProvider.AWSRegion)
	viper.Set(config.AWSRegion, rosaProvider.AWSRegion)

	return &ROSAProvider{
		ocmProvider:      ocmProvider,
		awsRegion:        rosaProvider.AWSRegion,
		versionGateLabel: versionGateLabel,
		provider:         rosaProvider,
	}, nil
}

// Type returns the provisioner type: rosa
func (m *ROSAProvider) Type() string {
	return "rosa"
}

// VersionGateLabel returns the provider version gate label
func (m *ROSAProvider) VersionGateLabel() string {
	return m.versionGateLabel
}

// callAndSetAWSSession sets the aws credentials as environment variables
// and runs the function provided
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
