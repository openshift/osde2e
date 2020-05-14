// Package moaprovider will allow the provisioning of clusters through moa.
package moaprovider

import (
	"fmt"

	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
)

// MOAProvider will provision clusters via MOA.
type MOAProvider struct {
	ocmProvider *ocmprovider.OCMProvider
}

// New will create a new MOAProvider.
func New(token string, env string, debug bool) (*MOAProvider, error) {
	ocmProvider, err := ocmprovider.New(token, env, debug)

	if err != nil {
		return nil, fmt.Errorf("error creating OCM provider for MOA provider: %v", err)
	}

	return &MOAProvider{
		ocmProvider: ocmProvider,
	}, nil
}
