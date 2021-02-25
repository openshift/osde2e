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
func (pdc Config) FireAlert(pipeline, details string) error {
	event := pd.V2Event{
		RoutingKey: pdc.IntegrationKey,
		Action:     "trigger",
		Payload: &pd.V2Payload{
			Summary:  pipeline + " failed",
			Severity: "info",
			Source:   pipeline,
			Details: map[string]string{
				"details": details,
			},
		},
	}
	_, err := pd.ManageEvent(event)
	if err != nil {
		return fmt.Errorf("failed firing alert: %w", err)
	}
	return nil
}
