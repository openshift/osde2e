package prometheus

import (
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/prometheus/client_golang/api"
	"github.com/spf13/viper"
)

// CreateClient will create a Prometheus client based off of the global config.
func CreateClient() (api.Client, error) {
	return api.NewClient(api.Config{
		Address:      viper.GetString(config.Prometheus.Address),
		RoundTripper: config.WeatherRoundTripper,
	})
}
