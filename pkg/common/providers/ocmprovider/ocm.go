// Package ocmprovider allows for the creation and management of OSD clusters through OCM.
package ocmprovider

import (
	"fmt"
	"strings"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/spi"

	ocm "github.com/openshift-online/ocm-sdk-go"
	ocmerr "github.com/openshift-online/ocm-sdk-go/errors"
)

const (
	// APIVersion is the version of the OSD API to use.
	APIVersion = "v1"

	// TokenURL specifies the endpoint used to create access tokens.
	TokenURL = "https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token"

	// fRTokenURL specifies the endpoint used to create access tokens for FedRamp.
	fRTokenURL = "https://sso.int.openshiftusgov.com/realms/redhat-external/protocol/openid-connect/token"
)

type ocmConnectionKey struct {
	token        string
	clientID     string
	clientSecret string
	env          string
	debug        bool
}

var connectionCache = map[ocmConnectionKey]*ocm.Connection{}

// OCMProvider will provision clusters using the OCM API.
type OCMProvider struct {
	env              string
	conn             *ocm.Connection
	clusterCache     map[string]*spi.Cluster
	credentialCache  map[string]string
	versionGateLabel string
}

// OCMConnection returns a raw OCM connection.
func OCMConnection(token, clientID, clientSecret, env string, debug bool) (*ocm.Connection, error) {
	cacheKey := ocmConnectionKey{
		token:        token,
		clientID:     clientID,
		clientSecret: clientSecret,
		env:          env,
		debug:        debug,
	}

	// Use the cached connection if possible
	if connection, ok := connectionCache[cacheKey]; ok {
		if connection == nil {
			return nil, fmt.Errorf("unable to get OCM connection, please check logs for details")
		}
		return connection, nil
	}

	logger, err := ocm.NewGoLoggerBuilder().
		Debug(debug).
		Build()
	if err != nil {
		return nil, fmt.Errorf("couldn't build logger: %v", err)
	}

	// select correct environment
	url := Environments.Choose(env)

	connectionBuilder := ocm.NewConnectionBuilder().URL(url).Logger(logger)
	// FedRamp uses a different tokenURL, so we need to check if url contains fr
	if strings.Contains(url, "fr") {
		connectionBuilder.Client(clientID, clientSecret).TokenURL(fRTokenURL)
	} else if clientID != "" && clientSecret != "" {
		connectionBuilder.Client(clientID, clientSecret)
	} else {
		connectionBuilder.TokenURL(TokenURL).Client("cloud-services", "").Tokens(token)
	}

	connection, err := connectionBuilder.Build()
	if err != nil {
		connectionCache[cacheKey] = nil
		return nil, err
	}

	connectionCache[cacheKey] = connection
	return connection, nil
}

// New returns a new OCMProvisioner.
func New() (*OCMProvider, error) {
	return NewWithEnv(viper.GetString(Env))
}

// NewWithEnv creates a new provider with a specific environment.
func NewWithEnv(env string) (*OCMProvider, error) {
	token := viper.GetString(Token)
	clientID := viper.GetString(ClientID)
	clientSecret := viper.GetString(ClientSecret)
	debug := viper.GetBool(Debug)

	conn, err := OCMConnection(token, clientID, clientSecret, env, debug)
	if err != nil {
		return nil, err
	}

	return &OCMProvider{
		env:              env,
		conn:             conn,
		clusterCache:     make(map[string]*spi.Cluster),
		credentialCache:  make(map[string]string),
		versionGateLabel: "api.openshift.com/gate-ocp",
	}, nil
}

// Environment simply returns the environment this OCMProvider is pointed to.
func (o *OCMProvider) Environment() string {
	return o.env
}

// Metrics returns the metrics of the cluster
func (o *OCMProvider) Metrics(clusterID string) (bool, error) {
	return true, nil
}

// UpgradeSource indicates that for stage/production clusters, we should use Cincinnati.
// For integration clusters, we should use the release controller.
func (o *OCMProvider) UpgradeSource() spi.UpgradeSource {
	// TODO: Is this different for FedRamp? I think it uses a different channel since they are behind commercial - Diego S.
	if o.env == stage || o.env == prod {
		return spi.CincinnatiSource
	}

	return spi.ReleaseControllerSource
}

// CincinnatiChannel returns a "fast" channel for stage and a "stable" channel for prod. This
// channel won't be used for integration since the upgrade source for integration is the release
// controller and not Cincinnati.
func (o *OCMProvider) CincinnatiChannel() spi.CincinnatiChannel {
	if o.Environment() == "prod" {
		return spi.CincinnatiStableChannel
	}
	return spi.CincinnatiFastChannel
}

// GetConnection returns the connection used by this provider.
func (o *OCMProvider) GetConnection() *ocm.Connection {
	return o.conn
}

// ErrResp takes an OCM error and converts it into a regular Golang error.
func errResp(resp *ocmerr.Error) error {
	if resp != nil {
		return fmt.Errorf("api error: %s", resp.Reason())
	}
	return nil
}

// Type returns the provisioner type: ocm
func (o *OCMProvider) Type() string {
	return "ocm"
}

// VersionGateLabel returns the provider version gate label
func (o *OCMProvider) VersionGateLabel() string {
	return o.versionGateLabel
}
