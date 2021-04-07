package pagerduty

import (
	"fmt"
	"log"
	"sort"

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

// MergeCICDIncidents merges all incidents for the CI Watcher by their title.
func MergeCICDIncidents(client *pd.Client) error {
	options := pd.ListIncidentsOptions{
		ServiceIDs: []string{"P7VT2V5"},
		Statuses:   []string{"triggered", "acknowledged"},
		APIListObject: pd.APIListObject{
			Limit: 100,
		},
	}
	var incidents []pd.Incident
	if err := Incidents(client, options, func(i pd.Incident) error {
		incidents = append(incidents, i)
		return nil
	}); err != nil {
		return fmt.Errorf("failed collecting incidents: %w", err)
	}
	if err := MergeIncidentsByTitle(client, incidents); err != nil {
		return fmt.Errorf("failed merging incidents: %w", err)
	}
	return nil
}

// MergeIncidentsByTitle will combine all incidents in the provided slice that share the
// same title into a single incident with multiple alerts.
func MergeIncidentsByTitle(c *pd.Client, incidents []pd.Incident) error {
	titleToIncident := make(map[string][]pd.Incident)

	for _, incident := range incidents {
		titleToIncident[incident.Title] = append(titleToIncident[incident.Title], incident)
	}

	for _, incidents := range titleToIncident {
		sort.Slice(incidents, func(i, j int) bool {
			return incidents[i].Id < incidents[j].Id
		})
		if len(incidents) < 2 {
			continue
		}
		first := incidents[0]
		mergeOptions := []pd.MergeIncidentsOptions{}
		for _, toMerge := range incidents[1:] {
			mergeOptions = append(mergeOptions, pd.MergeIncidentsOptions{
				ID:   toMerge.Id,
				Type: toMerge.APIObject.Type,
			})
		}
		log.Printf("Merging into %s: %v", first.Id, mergeOptions)
		_, err := c.MergeIncidents("", first.Id, mergeOptions)
		if err != nil {
			return fmt.Errorf("Failed merging %d incidents into %s: %w", len(incidents)-1, first.Id, err)
		}
	}
	return nil
}

// Incidents uses the provided client to retrieve all Incidents matching the provided
// list options and calls the handler function on each one.
func Incidents(c *pd.Client, options pd.ListIncidentsOptions, handler func(pd.Incident) error) error {
	var (
		il          = new(pd.ListIncidentsResponse)
		err         error
		previousLen int
	)
	firstIteration := true
	for il.APIListObject.More || firstIteration {
		firstIteration = false
		options.APIListObject.Offset = il.APIListObject.Offset + uint(previousLen)
		il, err = c.ListIncidents(options)
		if err != nil {
			return fmt.Errorf("failed listing incidents: %w", err)
		}
		previousLen = len(il.Incidents)
		for _, incident := range il.Incidents {
			if err := handler(incident); err != nil {
				return fmt.Errorf("handler failed: %w", err)
			}
		}
	}
	return nil
}

// Alerts uses the provided client to retrieve all Alerts associated with the provided
// incident, calling the provided handler function on each alert.
func Alerts(c *pd.Client, incident pd.Incident, options pd.ListIncidentAlertsOptions, handler func(pd.IncidentAlert) error) error {
	var (
		il          = new(pd.ListAlertsResponse)
		err         error
		previousLen int
	)
	firstIteration := true
	for il.APIListObject.More || firstIteration {
		firstIteration = false
		options.APIListObject.Offset = il.APIListObject.Offset + uint(previousLen)
		il, err = c.ListIncidentAlertsWithOpts(incident.Id, options)
		if err != nil {
			return fmt.Errorf("failed listing alerts: %w", err)
		}
		previousLen = len(il.Alerts)
		for _, alert := range il.Alerts {
			if err := handler(alert); err != nil {
				return fmt.Errorf("handler failed: %w", err)
			}
		}
	}
	return nil
}
