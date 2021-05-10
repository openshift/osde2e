package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	pd "github.com/PagerDuty/go-pagerduty"
	"github.com/fatih/color"
	"github.com/openshift/osde2e/pkg/common/pagerduty"
)

func AllIncidents(c *pd.Client) ([]pd.Incident, error) {
	var incidents []pd.Incident
	options := pd.ListIncidentsOptions{
		ServiceIDs: []string{"P7VT2V5"},
		Statuses:   []string{"triggered", "acknowledged"},
		APIListObject: pd.APIListObject{
			Limit: 100,
		},
	}
	if err := pagerduty.Incidents(c, options, func(incident pd.Incident) error {
		incidents = append(incidents, incident)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed listing incidents: %w", err)
	}
	return incidents, nil
}

func AllNotes(c *pd.Client, incident pd.Incident) ([]pd.IncidentNote, error) {
	return c.ListIncidentNotes(incident.Id)
}

func PrintIncident(incident pd.Incident, alerts []pd.IncidentAlert, notes []pd.IncidentNote) {
	if len(alerts) < 2 {
		return
	}
	status := incident.Status
	if status == "acknowledged" {
		status = color.GreenString(status)
	} else if status == "triggered" {
		status = color.MagentaString(status)
	}
	fmt.Printf("%-3d %10s %s\n", len(alerts), status, incident.Title)
	segments := make([]map[string]struct{}, 0)
	for _, alert := range alerts {
		url := alert.Body["details"].(map[string]interface{})["details"].(string)
		parts := strings.Split(url, "/")
		build := parts[len(parts)-2]
		if strings.HasPrefix(build, "release-") {
			continue
		}
		for i, segment := range strings.Split(build, "-") {
			if i >= len(segments) {
				segments = append(segments, make(map[string]struct{}))
			}
			segments[i][segment] = struct{}{}
		}
	}
	for i, value := range segments {
		if i == 0 {
			continue
		}
		values := []string{}
		for k := range value {
			values = append(values, k)
		}
		if len(values) == 1 {
			fmt.Print(color.RedString(fmt.Sprint(values)))
		} else {
			fmt.Print(values)
		}
	}
	fmt.Print("\n")
	fmt.Println(incident.HTMLURL)
	for _, note := range notes {
		fmt.Println(color.CyanString(note.Content))
	}
	fmt.Print("\n")
}

func run() error {
	client := pd.NewClient(os.Getenv("PAGERDUTY_TOKEN"))
	incidents, err := AllIncidents(client)
	if err != nil {
		return fmt.Errorf("failed listing incidents: %w", err)
	}

	alertsForIncident := make(map[string][]pd.IncidentAlert)
	notesForIncident := make(map[string][]pd.IncidentNote)

	for _, incident := range incidents {
		if err := pagerduty.Alerts(client, incident, pd.ListIncidentAlertsOptions{}, func(alert pd.IncidentAlert) error {
			alertsForIncident[incident.Id] = append(alertsForIncident[incident.Id], alert)
			return nil
		}); err != nil {
			return fmt.Errorf("failed listing alerts: %w", err)
		}
		if notes, err := AllNotes(client, incident); err != nil {
			return fmt.Errorf("failed listing notes: %w", err)
		} else {
			notesForIncident[incident.Id] = notes
		}
	}

	sort.Slice(incidents, func(i, j int) bool {
		a := incidents[i].Id
		b := incidents[j].Id
		return len(alertsForIncident[a]) > len(alertsForIncident[b])
	})

	for _, incident := range incidents {
		alerts := alertsForIncident[incident.Id]
		notes := notesForIncident[incident.Id]
		PrintIncident(incident, alerts, notes)
	}

	return nil
}

func main() {
	switch err := run().(type) {
	default:
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		os.Exit(0)
	}
}
