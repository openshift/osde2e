package providers

import (
	"fmt"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
	"github.com/openshift/osde2e/pkg/common/spi"
)

// ClusterProvider returns the provisioner configured by the config object.
func ClusterProvider() (spi.Provider, error) {
	switch config.Instance.Provider {
	case "ocm":
		return ocmprovider.New(config.Instance.OCM.Token, config.Instance.OCM.Env, config.Instance.OCM.Debug)
	default:
		return nil, fmt.Errorf("unrecognized provisioner: %s", config.Instance.Provider)
	}
}

// ClusterProviderForProduction returns the provisioner configured by the config object using the production environment.
func ClusterProviderForProduction() (spi.Provider, error) {
	switch config.Instance.Provider {
	case "ocm":
		return ocmprovider.New(config.Instance.OCM.Token, "prod", config.Instance.OCM.Debug)
	default:
		return nil, fmt.Errorf("unrecognized provisioner: %s", config.Instance.Provider)
	}
}
