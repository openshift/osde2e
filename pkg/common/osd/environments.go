package osd

// Environments are known instance of OSD.
var Environments = environments{
	// default to using integration environment
	"": "int",

	// environments available
	"int":   "https://api-integration.6943.hive-integration.openshiftapps.com",
	"stage": "https://api.stage.openshift.com",
	"prod":  "https://api.openshift.com",
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
