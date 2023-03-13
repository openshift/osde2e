package ocmprovider

const (
	integration = "int"
	stage       = "stage"
	prod        = "prod"

	govintegration = "govint"
	govstage       = "govstage"
	govprod        = "govprod"

	integrationURL = "https://api.integration.openshift.com"
	stageURL       = "https://api.stage.openshift.com"
	prodURL        = "https://api.openshift.com"

	govintegrationURL = "https://api.int.openshiftusgov.com"
	govstageURL       = "https://api.stage.openshiftusgov.com"
	govprodURL        = "https://api.openshiftusgov.com"
)

// Environments are known instance of OSD.
var Environments = environments{
	// default to using integration environment
	"": integration,

	// environments available
	integration:    integrationURL,
	stage:          stageURL,
	prod:           prodURL,
	govintegration: govintegrationURL,
	govstage:       govstageURL,
	govprod:        govprodURL,
}

type environments map[string]string

// Choose returns the endpoint for the desired OSD environment. If desired is URL, it will be returned as the endpoint.
func (e environments) Choose(desired string) string {
	if val, ok := e[desired]; !ok || desired == val {
		return desired
	} else {
		return e.Choose(val)
	}
}
