package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	pd "github.com/PagerDuty/go-pagerduty"
	"github.com/fatih/color"
	"github.com/openshift/osde2e/pkg/common/pagerduty"
)

// AllIncidents returns the unresolved incidents for the CI Watcher.
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

// AllNotes returns the pagerduty notes for a given incident.
func AllNotes(c *pd.Client, incident pd.Incident) ([]pd.IncidentNote, error) {
	return c.ListIncidentNotes(incident.Id)
}

// PrintIncident pretty-prints an incident with data from its alerts and notes.
func PrintIncident(incident pd.Incident, alerts []pd.IncidentAlert, notes []pd.IncidentNote) {
	status := incident.Status
	if status == "acknowledged" {
		status = color.GreenString(status)
	} else if status == "triggered" {
		status = color.MagentaString(status)
	}
	fmt.Printf("%s\n%-3d %10s\n", incident.Title, len(alerts), status)
	jobNames := make(map[string]struct{}, 0)
	for _, alert := range alerts {
		data, ok := alert.Body["details"].(map[string]interface{})
		if !ok {
			continue
		}
		current, ok := data["current"].(map[string]interface{})
		if !ok {
			continue
		}
		name, ok := current["job_name"].(string)
		if !ok {
			continue
		}
		if strings.HasPrefix(name, "release-") || name == "" {
			continue
		}
		jobNames[name] = struct{}{}
	}
	for name := range jobNames {
		fmt.Println(name)
	}
	fmt.Println(color.YellowString(incident.HTMLURL))
	for _, note := range notes {
		fmt.Println(color.CyanString(note.Content))
	}
	fmt.Print("\n")
}

// run examines pagerduty and prints a report
func run() error {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), `Usage of %[1]s:

%[1]s

You must set the $PAGERDUTY_TOKEN environment variable to your
personal pagerduty token in order for the report to be generated.
`, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
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
