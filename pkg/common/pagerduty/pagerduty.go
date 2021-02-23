package pagerduty

import (
	"fmt"

	pd "github.com/PagerDuty/go-pagerduty"
)

// Config holds state necessary to open pagerduty incidents against a particular
// service and associated with a particular escalation policy.
type Config struct {
	Token, ServiceID, PolicyID string
}

// FireIncident attempts to create an incident with the provided title and body that
// is associated with the service and policy IDs stored in the PDConfig
func (pdc Config) FireIncident(title, body string) error {
	client := pd.NewClient(pdc.Token)
	service, err := pdService(client, pdc.ServiceID)
	if err != nil {
		return err
	}
	policy, err := pdPolicy(client, pdc.PolicyID)
	if err != nil {
		return err
	}
	_, err = client.CreateIncident(title, &pd.CreateIncidentOptions{
		Title: title,
		Type:  "incident",
		Body: &pd.APIDetails{
			Type:    "incident_body",
			Details: body,
		},
		Urgency:          "low",
		Service:          &service,
		EscalationPolicy: &policy,
	})
	if err != nil {
		return err
	}
	return nil
}

func pdService(client *pd.Client, id string) (pd.APIReference, error) {
	s, err := client.GetService(id, nil)
	if err != nil {
		return pd.APIReference{}, fmt.Errorf("failed looking up service %s: %w", id, err)
	}
	return pdRef(s.APIObject), nil
}

func pdPolicy(client *pd.Client, id string) (pd.APIReference, error) {
	p, err := client.GetEscalationPolicy(id, nil)
	if err != nil {
		return pd.APIReference{}, fmt.Errorf("failed looking up escalation policy %s: %w", id, err)
	}
	return pdRef(p.APIObject), nil
}

func pdRef(obj pd.APIObject) pd.APIReference {
	return pd.APIReference{
		ID:   obj.ID,
		Type: obj.Type,
	}
}
