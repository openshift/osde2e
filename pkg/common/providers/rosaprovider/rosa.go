// Package rosaprovider will allow the provisioning of clusters through rosa.
package rosaprovider

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/openshift/osde2e/pkg/common/aws"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
	"github.com/openshift/osde2e/pkg/common/spi"
)

func init() {
	spi.RegisterProvider("rosa", func() (spi.Provider, error) { return New() })
}

// ROSAProvider will provision clusters via ROSA.
type ROSAProvider struct {
	ocmProvider      *ocmprovider.OCMProvider
	awsCredentials   *credentials.Value
	awsRegion        string
	versionGateLabel string
}

// New will create a new ROSAProvider.
func New() (*ROSAProvider, error) {
	ocmProvider, err := ocmprovider.NewWithEnv(viper.GetString(Env))
	if err != nil {
		return nil, fmt.Errorf("error creating OCM provider for ROSA provider: %v", err)
	}

	awsCredentials, err := aws.CcsAwsSession.GetCredentials()
	if err != nil {
		return nil, fmt.Errorf("error creating aws session: %v", err)
	}

	region := *aws.CcsAwsSession.GetRegion()
	if region == "" {
		return nil, fmt.Errorf("aws region is undefined")
	}

	versionGateLabel := "api.openshift.com/gate-ocp"
	if viper.GetBool(STS) {
		versionGateLabel = "api.openshift.com/gate-sts"
	}

	return &ROSAProvider{
		ocmProvider:      ocmProvider,
		awsCredentials:   awsCredentials,
		awsRegion:        region,
		versionGateLabel: versionGateLabel,
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
