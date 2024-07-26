package ocmprovider

import (
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
)

const (
	integration   = "int"
	stage         = "stage"
	prod          = "prod"
	frIntegration = "frInt"
	frStage       = "frStage"
	frProd        = "frProd"

	integrationURL   = "https://api.integration.openshift.com"
	stageURL         = "https://api.stage.openshift.com"
	prodURL          = "https://api.openshift.com"
	frIntegrationURL = "https://api.int.openshiftusgov.com"
	frStageURL       = "https://api.stage.openshiftusgov.com"
	frProdURL        = "https://api.openshiftusgov.com"
)

// Environments are known instance of OSD.
var Environments = environments{
	// default to using integration environment
	"": integration,

	// environments available
	integration:   integrationURL,
	stage:         stageURL,
	prod:          prodURL,
	frIntegration: frIntegrationURL,
	frStage:       frStageURL,
	frProd:        frProdURL,
}

type environments map[string]string

// Choose returns the endpoint for the desired OSD environment. If desired is URL, it will be returned as the endpoint.
func (e environments) Choose(desired string) string {
	if viper.GetBool(config.Cluster.FedRamp) {
		switch desired {
		case integration:
			desired = frIntegration
		case stage:
			desired = frStage
		case prod:
			desired = frProd
		}
	}
	if val, ok := e[desired]; !ok || desired == val {
		return desired
	} else {
		return e.Choose(val)
	}
}
