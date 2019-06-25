// Package osd allows for the creation and management of OSD clusters.
package osd

import (
	"fmt"

	uhc "github.com/openshift-online/uhc-sdk-go/pkg/client"
	"github.com/openshift-online/uhc-sdk-go/pkg/client/clustersmgmt/v1"
	"github.com/openshift-online/uhc-sdk-go/pkg/client/errors"
)

const (
	// StagingURL is the API endpoint for the OSD staging environment.
	StagingURL = "https://api.stage.openshift.com"

	// APIVersion is the version of the OSD API to use.
	APIVersion = "v1"

	// TokenURL specifies the endpoint used to create access tokens.
	TokenURL = "https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token"

	// ClientID is used to identify the client to OSD.
	ClientID = "cloud-services"
)

// New setups a client to connect to OSD.
func New(token string, staging, debug bool) (*OSD, error) {
	logger, err := uhc.NewGoLoggerBuilder().
		Debug(debug).
		Build()
	if err != nil {
		return nil, fmt.Errorf("couldn't build logger: %v", err)
	}

	builder := uhc.NewConnectionBuilder().
		TokenURL(TokenURL).
		Client(ClientID, "").
		Logger(logger).
		Tokens(token)

	if staging {
		builder.URL(StagingURL)
	}

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

// clusters returns a client used to perform cluster operations.
func (u *OSD) clusters() *v1.ClustersClient {
	return u.conn.ClustersMgmt().V1().Clusters()
}

// cluster returns the client for a specific cluster
func (u *OSD) cluster(clusterID string) *v1.ClusterClient {
	return u.clusters().Cluster(clusterID)
}

// versions returns a client used to retrieve versions currently offered by OSD.
func (u *OSD) versions() *v1.VersionsClient {
	return u.conn.ClustersMgmt().V1().Versions()
}

func errResp(resp *errors.Error) error {
	if resp != nil {
		return fmt.Errorf("api error: %s", resp.Reason())
	}
	return nil
}
