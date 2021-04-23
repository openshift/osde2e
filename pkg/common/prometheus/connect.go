package prometheus

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/prometheus/client_golang/api"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
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

func CreateClusterClient(h *helper.H) (api.Client, error) {

	promHost, err := getClusterPrometheusHost(h)
	if err != nil {
		return nil, err
	}
	clusterBearerToken, err := getClusterPrometheusToken(h)
	if err != nil {
		return nil, err
	}

	return api.NewClient(api.Config{
		Address:      *promHost,
		RoundTripper: createRoundTripper(*clusterBearerToken),
	})
}

func getClusterPrometheusHost(h *helper.H) (*string, error) {
	route, err := h.Route().RouteV1().Routes("openshift-monitoring").Get(context.TODO(), "prometheus-k8s", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	hostUrl := "https://" + route.Spec.Host
	return &hostUrl, nil
}

func getClusterPrometheusToken(h *helper.H) (*string, error) {
	sa, err := h.Kube().CoreV1().ServiceAccounts("openshift-monitoring").Get(context.TODO(), "prometheus-k8s", metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("Unable to fetch prometheus-k8s service account: %s", err)
	}

	tokenSecret := ""
	for _, secret := range sa.Secrets {
		if strings.HasPrefix(secret.Name, "prometheus-k8s-token") {
			tokenSecret = secret.Name
		}
	}
	if len(tokenSecret) == 0 {
		return nil, fmt.Errorf("Failed to find token secret for prometheus-k8s SA")
	}

	secret, err := h.Kube().CoreV1().Secrets("openshift-monitoring").Get(context.TODO(), tokenSecret, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("Unable to fetch secret %s: %s", tokenSecret, err)
	}

	token := secret.Data[corev1.ServiceAccountTokenKey]
	stringToken := string(token)

	return &stringToken, nil
}
