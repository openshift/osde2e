package osd

import (
	"fmt"

	uhc "github.com/openshift-online/uhc-sdk-go/pkg/client"
)

const (
	StagingURL = "https://api.stage.openshift.com"
	APIPrefix  = "/api/clusters_mgmt"
	APIVersion = "v1"
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
