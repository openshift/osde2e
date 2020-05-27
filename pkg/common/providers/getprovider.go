package providers

import (
	"fmt"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/providers/crc"
	"github.com/openshift/osde2e/pkg/common/providers/mock"
	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/spf13/viper"
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
	provider := viper.GetString(config.Provider)
	switch provider {
	case OCM:
		return ocmprovider.New()
	case Mock:
		return mock.New()
	case CRC:
		return crc.New()
	default:
		return nil, fmt.Errorf("unrecognized provisioner: %s", provider)
	}
}
