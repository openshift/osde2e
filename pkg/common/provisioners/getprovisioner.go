package provisioners

import (
	"fmt"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/provisioners/ocmprovisioner"
	"github.com/openshift/osde2e/pkg/common/spi"
)

// ClusterProvisioner returns the provisioner configured by the config object.
func ClusterProvisioner() (spi.Provisioner, error) {
	switch config.Instance.Provisioner {
	case "ocm":
		return ocmprovisioner.New(config.Instance.OCM.Token, config.Instance.OCM.Env, config.Instance.OCM.Debug)
	default:
		return nil, fmt.Errorf("unrecognized provisioner: %s", config.Instance.Provisioner)
	}
}

// ClusterProvisionerForProduction returns the provisioner configured by the config object using the production environment.
func ClusterProvisionerForProduction() (spi.Provisioner, error) {
	switch config.Instance.Provisioner {
	case "ocm":
		return ocmprovisioner.New(config.Instance.OCM.Token, "prod", config.Instance.OCM.Debug)
	default:
		return nil, fmt.Errorf("unrecognized provisioner: %s", config.Instance.Provisioner)
	}
}
