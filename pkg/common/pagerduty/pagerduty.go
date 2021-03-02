package pagerduty

import (
	"fmt"

	pd "github.com/PagerDuty/go-pagerduty"
)

// Config holds state necessary to open pagerduty incidents using an integration key.
type Config struct {
	IntegrationKey string
}

// FireIncident attempts to create an alert indicating a failure in the provided
// pipeline.
func (pdc Config) FireAlert(details pd.V2Payload) (*pd.V2EventResponse, error) {
	event := pd.V2Event{
		RoutingKey: pdc.IntegrationKey,
		Action:     "trigger",
		Payload:    &details,
	}
	e, err := pd.ManageEvent(event)
	if err != nil {
		return e, fmt.Errorf("failed firing alert: %w", err)
	}
	return e, nil
}
