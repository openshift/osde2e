package providers

import (
	"fmt"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/providers/moaprovider"
	"github.com/openshift/osde2e/pkg/common/providers/mock"
	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
	"github.com/openshift/osde2e/pkg/common/spi"
)

const (
	// OCM provider.
	OCM = "ocm"

	// MOA provider.
	MOA = "moa"

	// Mock provider.
	Mock = "mock"
)

// ClusterProvider returns the provisioner configured by the config object.
func ClusterProvider() (spi.Provider, error) {
	switch config.Instance.Provider {
	case OCM:
		return ocmprovider.New(config.Instance.OCM.Token, config.Instance.OCM.Env, config.Instance.OCM.Debug)
	case MOA:
		return moaprovider.New(config.Instance.OCM.Token, config.Instance.OCM.Env, config.Instance.OCM.Debug)
	case Mock:
		return mock.New(config.Instance.OCM.Env)
	default:
		return nil, fmt.Errorf("unrecognized provisioner: %s", config.Instance.Provider)
	}
}

// ClusterProviderForProduction returns the provisioner configured by the config object using the production environment.
func ClusterProviderForProduction() (spi.Provider, error) {
	switch config.Instance.Provider {
	case OCM:
		return ocmprovider.New(config.Instance.OCM.Token, "prod", config.Instance.OCM.Debug)
	case Mock:
		return mock.New("prod")
	default:
		return nil, fmt.Errorf("unrecognized provisioner: %s", config.Instance.Provider)
	}
}
