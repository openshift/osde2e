package pagerduty

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

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

// ProcessCICDIncidents merges all incidents for the CI Watcher by their title and then closes
// any incidents that haven't had recent alerts.
func ProcessCICDIncidents(client *pd.Client) error {
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
	if err := ResolveOldIncidents(client, incidents); err != nil {
		return fmt.Errorf("failed resolving incidents: %w", err)
	}
	return nil
}

// ResolveOldIncidents automatically resolves any incident whose most recent alert is
// older than 30 hours.
func ResolveOldIncidents(c *pd.Client, incidents []pd.Incident) error {
	now := time.Now()
	changes := []pd.ManageIncidentsOptions{}
	for _, i := range incidents {
		newest := time.Time{}
		if err := Alerts(c, i, pd.ListIncidentAlertsOptions{}, func(a pd.IncidentAlert) error {
			t, err := time.Parse(time.RFC3339, a.CreatedAt)
			if err != nil {
				return fmt.Errorf("Unable to parse time %v: %w", a.CreatedAt, err)
			}
			if t.After(newest) {
				newest = t
			}
			return nil
		}); err != nil {
			return fmt.Errorf("failed listing alerts for incident %s: %w", i.Id, err)
		}
		if age := now.Sub(newest); newest != (time.Time{}) && age > time.Hour*30 {
			log.Printf("Resolving incident %v because it is %v old.", i.Id, age)
			changes = append(changes, pd.ManageIncidentsOptions{
				ID:     i.Id,
				Type:   i.Type,
				Status: "resolved",
			})
		} else {
			log.Printf("Not resolving %v, newest alert is %v old", i.Id, age)
		}
	}
	if len(changes) > 0 {
		if _, err := c.ManageIncidents("", changes); err != nil {
			if err != nil {
				return fmt.Errorf("failed auto-resolving incidents: %w", err)
			}
		}
	}
	return nil
}

// MergeIncidentsByTitle will combine all incidents in the provided slice that share the
// same title into a single incident with multiple alerts.
func MergeIncidentsByTitle(c *pd.Client, incidents []pd.Incident) error {
	titleToIncident := make(map[string][]pd.Incident)

	for _, incident := range incidents {
		title := incident.Title
		title = strings.TrimPrefix(title, "[install] ")
		title = strings.TrimPrefix(title, "[upgrade] ")
		titleToIncident[title] = append(titleToIncident[title], incident)
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
