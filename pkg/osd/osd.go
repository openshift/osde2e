// Package osd allows for the creation and management of OSD clusters.
package osd

import (
	"errors"
	"fmt"
	uhc "github.com/openshift-online/uhc-sdk-go/pkg/client"
	accounts "github.com/openshift-online/uhc-sdk-go/pkg/client/accountsmgmt/v1"
	clusters "github.com/openshift-online/uhc-sdk-go/pkg/client/clustersmgmt/v1"
	uhcerr "github.com/openshift-online/uhc-sdk-go/pkg/client/errors"
)

const (
	// APIVersion is the version of the OSD API to use.
	APIVersion = "v1"

	// TokenURL specifies the endpoint used to create access tokens.
	TokenURL = "https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token"

	// ClientID is used to identify the client to OSD.
	ClientID = "cloud-services"
)

// New setups a client to connect to OSD.
func New(token, env string, debug bool) (*OSD, error) {
	logger, err := uhc.NewGoLoggerBuilder().
		Debug(debug).
		Build()
	if err != nil {
		return nil, fmt.Errorf("couldn't build logger: %v", err)
	}

	// select correct environment
	url := Environments.Choose(env)

	builder := uhc.NewConnectionBuilder().
		URL(url).
		TokenURL(TokenURL).
		Client(ClientID, "").
		Logger(logger).
		Tokens(token)

	conn, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("couldn't setup connection: %v", err)
	}

	return &OSD{
		conn: conn,
	}, nil
}

// OSD acts as a client to manage an instance.
type OSD struct {
	conn *uhc.Connection
}

// CurrentAccount returns the current account being used.
func (u *OSD) CurrentAccount() (*accounts.Account, error) {
	act, err := u.conn.AccountsMgmt().V1().CurrentAccount().Get().Send()
	if err == nil && act != nil {
		err = errResp(act.Error())
	} else if act == nil {
		return nil, errors.New("account can't be nil")
	}
	return act.Body(), err
}

// clusters returns a client used to perform cluster operations.
func (u *OSD) clusters() *clusters.ClustersClient {
	return u.conn.ClustersMgmt().V1().Clusters()
}

// cluster returns the client for a specific cluster
func (u *OSD) cluster(clusterID string) *clusters.ClusterClient {
	return u.clusters().Cluster(clusterID)
}

// versions returns a client used to retrieve versions currently offered by OSD.
func (u *OSD) versions() *clusters.VersionsClient {
	return u.conn.ClustersMgmt().V1().Versions()
}

func errResp(resp *uhcerr.Error) error {
	if resp != nil {
		return fmt.Errorf("api error: %s", resp.Reason())
	}
	return nil
}
