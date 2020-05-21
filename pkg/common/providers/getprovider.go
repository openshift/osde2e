package providers

import (
	"fmt"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/providers/crc"
	"github.com/openshift/osde2e/pkg/common/providers/mock"
	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
	"github.com/openshift/osde2e/pkg/common/spi"
)

const (
	// OCM provider.
	OCM = "ocm"

	// Mock provider.
	Mock = "mock"

	// CRC provider.
	CRC = "crc"
)

// ClusterProvider returns the provisioner configured by the config object.
func ClusterProvider() (spi.Provider, error) {
	switch config.Instance.Provider {
	case OCM:
		return ocmprovider.New(config.Instance.OCM.Token, config.Instance.OCM.Env, config.Instance.OCM.Debug)
	case Mock:
		return mock.New(config.Instance.OCM.Env)
	case CRC:
		return crc.New("crc")
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
