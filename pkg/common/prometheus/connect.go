package prometheus

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/prometheus/client_golang/api"
	"github.com/spf13/viper"
)

// CreateClient will create a Prometheus client.
// If no arguments are supplied, the global config will be used.
// If one argument is supplied, it will be used as the address for Prometheus, but will use the global config for the bearer token.
// If two arguments are supplied, the first will be used as the address for Prometheus and the second will be used as the bearer token.
func CreateClient(args ...string) (api.Client, error) {
	numArgs := len(args)
	if numArgs > 2 {
		return nil, fmt.Errorf("unexpected number of arguments, only 2 are supported")
	}

	var address, bearerToken string

	if numArgs == 0 {
		address = viper.GetString(config.Prometheus.Address)
		bearerToken = viper.GetString(config.Prometheus.BearerToken)
	} else if numArgs == 1 {
		address = args[0]
		bearerToken = viper.GetString(config.Prometheus.BearerToken)
	} else if numArgs == 2 {
		address = args[0]
		bearerToken = args[1]
	}

	return api.NewClient(api.Config{
		Address:      address,
		RoundTripper: createRoundTripper(bearerToken),
	})
}

// createRoundTripper will create a RoundTripper like api.DefaultRoundTripper with an added stripping
// of cert verification and adding the bearer token to the HTTP request
func createRoundTripper(bearerToken string) http.RoundTripper {
	return &http.Transport{
		Proxy: func(request *http.Request) (*url.URL, error) {
			request.Header.Add("Authorization", "Bearer "+bearerToken)
			return http.ProxyFromEnvironment(request)
		},
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		TLSHandshakeTimeout: 10 * time.Second,
	}
}
