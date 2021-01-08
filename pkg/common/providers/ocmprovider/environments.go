package ocmprovider

const (
	crc         = "crc"
	integration = "int"
	stage       = "stage"
	prod        = "prod"

	crcURL         = "https://clusters-service.apps-crc.testing"
	integrationURL = "https://api.integration.openshift.com"
	stageURL       = "https://api.stage.openshift.com"
	prodURL        = "https://api.openshift.com"
)

// Environments are known instance of OSD.
var Environments = environments{
	// default to using integration environment
	"": integration,

	// environments available
	crc:         crcURL,
	integration: integrationURL,
	stage:       stageURL,
	prod:        prodURL,
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
