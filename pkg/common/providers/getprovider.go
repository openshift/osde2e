package providers

import (
	"log"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/spf13/viper"
)

// ClusterProvider returns the provisioner configured by the config object.
func ClusterProvider() (spi.Provider, error) {
	log.Println("Entered clusterprovider()")
	provider := viper.GetString(config.Provider)
	log.Printf("The provider is - %v", provider)
	return spi.GetProvider(provider)
}
