// Package main allows for the creation and management of OSD clusters through OCM.
package main

import (
	"fmt"
	"sync"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/spi"

	ocm "github.com/openshift-online/ocm-sdk-go"
	ocmerr "github.com/openshift-online/ocm-sdk-go/errors"
)

const (
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
	prodProvider spi.Provider

	// Since getting versions is a noisy operation, we'll just cache the version retrieval.
	// This changes rarely and we only ever look at it once at the start of time, so it's not
	// expected to meaningfully change over the course of a run.
	versionCacheOnce sync.Once
	versionCache     *spi.VersionList
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

	connection, err := builder.Build()

	if err != nil {
		connectionCache[cacheKey] = nil
		return nil, err
	}

	connectionCache[cacheKey] = connection
	return connection, nil
}

// ProviderName is the name of the provider.
func ProviderName() string {
	return "ocm"
}

// New returns a new OCMProvisioner.
func New(cfg *config.Config) (spi.Provider, error) {
	conn, err := OCMConnection(cfg.OCM.Token, cfg.OCM.Env, cfg.OCM.Debug)

	if err != nil {
		return nil, err
	}

	var prodProvider spi.Provider = nil

	// For integration/stage environments, we need a connection to production so that we're
	// able to get the default version in production. This will allow us to make relative version
	// upgrades by measuring against the current production default.
	if cfg.OCM.Env != prod {
		prodConfig := *cfg
		prodConfig.OCM.Env = prod
		prodProvider, err = New(&prodConfig)

		if err != nil {
			return nil, err
		}
	}

	return &OCMProvider{
		env:              cfg.OCM.Env,
		conn:             conn,
		prodProvider:     prodProvider,
		versionCacheOnce: sync.Once{},
	}, nil
}

// Environment simply returns the environment this OCMProvider is pointed to.
func (o *OCMProvider) Environment() string {
	return o.env
}

// UpgradeSource indicates that for stage/production clusters, we should use Cincinnati.
// For integration clusters, we should use the release controller.
func (o *OCMProvider) UpgradeSource() spi.UpgradeSource {
	if o.env != integration {
		return spi.CincinnatiSource
	}

	return spi.ReleaseControllerSource
}

// CincinnatiChannel returns a "fast" channel for stage and a "stable" channel for prod. This
// channel won't be used for integration since the upgrade source for integration is the release
// controller and not Cincinnati.
func (o *OCMProvider) CincinnatiChannel() spi.CincinnatiChannel {
	if o.env == stage {
		return spi.CincinnatiFastChannel
	}

	return spi.CincinnatiStableChannel
}

// ErrResp takes an OCM error and converts it into a regular Golang error.
func errResp(resp *ocmerr.Error) error {
	if resp != nil {
		return fmt.Errorf("api error: %s", resp.Reason())
	}
	return nil
}
