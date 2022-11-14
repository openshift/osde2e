// Package rosaprovider will allow the provisioning of clusters through rosa.
package rosaprovider

import (
	"fmt"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
	"github.com/openshift/osde2e/pkg/common/spi"
)

func init() {
	spi.RegisterProvider("rosa", func() (spi.Provider, error) { return New() })
}

// ROSAProvider will provision clusters via ROSA.
type ROSAProvider struct {
	ocmProvider *ocmprovider.OCMProvider
}

// New will create a new ROSAProvider.
func New() (*ROSAProvider, error) {
	ocmProvider, err := ocmprovider.NewWithEnv(viper.GetString(Env))
	if err != nil {
		return nil, fmt.Errorf("error creating OCM provider for ROSA provider: %v", err)
	}

	return &ROSAProvider{
		ocmProvider: ocmProvider,
	}, nil
}

// Type returns the provisioner type: rosa
func (m *ROSAProvider) Type() string {
	return "rosa"
}
