package providers

import (
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/spf13/viper"
)

// ClusterProvider returns the provisioner configured by the config object.
func ClusterProvider() (spi.Provider, error) {
	provider := viper.GetString(config.Provider)
	return spi.GetProvider(provider)
}
