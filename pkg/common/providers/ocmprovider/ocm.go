// Package ocmprovider allows for the creation and management of OSD clusters through OCM.
package ocmprovider

import (
	"fmt"

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

	// ClientID is used to identify the client to OSD.
	ClientID = "cloud-services"
)

type ocmConnectionKey struct {
	token string
	env   string
	debug bool
}

var connectionCache = map[ocmConnectionKey]*ocm.Connection{}

// OCMProvider will provision clusters using the OCM API.
type OCMProvider struct {
	env          string
	conn         *ocm.Connection
	prodProvider *OCMProvider

	clusterCache    map[string]*spi.Cluster
	credentialCache map[string]string
}

func init() {
	spi.RegisterProvider("ocm", func() (spi.Provider, error) { return New() })
}

// OCMConnection returns a raw OCM connection.
func OCMConnection(token, env string, debug bool) (*ocm.Connection, error) {
	cacheKey := ocmConnectionKey{
		token: token,
		env:   env,
		debug: debug,
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

	builder := ocm.NewConnectionBuilder().
		URL(url).
		TokenURL(TokenURL).
		Client(ClientID, "").
		Logger(logger).
		Tokens(token)

	if env == crc {
		builder = builder.Insecure(true)
	}

	connection, err := builder.Build()
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
	debug := viper.GetBool(Debug)

	conn, err := OCMConnection(token, env, debug)
	if err != nil {
		return nil, err
	}

	var prodProvider *OCMProvider = nil

	// For integration/stage environments, we need a connection to production so that we're
	// able to get the default version in production. This will allow us to make relative version
	// upgrades by measuring against the current production default.
	if env != prod {
		prodProvider, err = NewWithEnv(prod)

		if err != nil {
			return nil, err
		}
	}

	return &OCMProvider{
		env:             env,
		conn:            conn,
		prodProvider:    prodProvider,
		clusterCache:    make(map[string]*spi.Cluster),
		credentialCache: make(map[string]string),
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
