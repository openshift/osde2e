// Package moaprovider will allow the provisioning of clusters through moa.
package moaprovider

import (
	"fmt"

	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/spf13/viper"
)

func init() {
	spi.RegisterProvider("moa", func() (spi.Provider, error) { return New() })
}

// MOAProvider will provision clusters via MOA.
type MOAProvider struct {
	ocmProvider *ocmprovider.OCMProvider
}

// New will create a new MOAProvider.
func New() (*MOAProvider, error) {
	ocmProvider, err := ocmprovider.NewWithEnv(viper.GetString(Env))

	if err != nil {
		return nil, fmt.Errorf("error creating OCM provider for MOA provider: %v", err)
	}

	return &MOAProvider{
		ocmProvider: ocmProvider,
	}, nil
}

// Type returns the provisioner type: moa
func (m *MOAProvider) Type() string {
	return "moa"
}
