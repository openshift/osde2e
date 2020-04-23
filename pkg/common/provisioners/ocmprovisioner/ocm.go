// Package ocmprovisioner allows for the creation and management of OSD clusters through OCM.
package ocmprovisioner

import (
	"fmt"
	"sync"

	"github.com/openshift/osde2e/pkg/common/spi"

	ocm "github.com/openshift-online/ocm-sdk-go"
	accounts "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
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

// OCMProvisioner will provision clusters using the OCM API.
type OCMProvisioner struct {
	conn *ocm.Connection

	// Since getting versions is a noisy operation, we'll just cache the version retrieval.
	// This changes rarely and we only ever look at it once at the start of time, so it's not
	// expected to meaningfully change over the course of a run.
	versionCacheOnce sync.Once
	versionCache     []spi.Version
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

// New returns a new OCMProvisioner.
func New(token string, env string, debug bool) (*OCMProvisioner, error) {
	conn, err := OCMConnection(token, env, debug)

	if err != nil {
		return nil, err
	}
	return &OCMProvisioner{
		conn:             conn,
		versionCacheOnce: sync.Once{},
	}, nil
}

// CurrentAccount returns the current account being used.
func (o *OCMProvisioner) CurrentAccount() (*accounts.Account, error) {
	var act *accounts.CurrentAccountGetResponse

	err := retryer().Do(func() error {
		var err error
		act, err = o.conn.AccountsMgmt().V1().CurrentAccount().Get().Send()

		if err != nil {
			return err
		}

		if act != nil && act.Error() != nil {
			return errResp(act.Error())
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error getting current account: %v", err)
	} else if act == nil {
		return nil, fmt.Errorf("account can't be nil")
	}

	return act.Body(), err
}

// ErrResp takes an OCM error and converts it into a regular Golang error.
func errResp(resp *ocmerr.Error) error {
	if resp != nil {
		return fmt.Errorf("api error: %s", resp.Reason())
	}
	return nil
}
