package state

import "testing"

func TestFindCriticalAlerts(t *testing.T) {
	tests := []struct {
		Name        string
		Results     []result
		Provider    string
		Environment string
		Expected    bool
	}{
		{
			Name: "found critical",
			Results: []result{
				{
					Metric: metric{
						AlertName: "alert1",
						Severity:  "info",
					},
				},
				{
					Metric: metric{
						AlertName: "alert2",
						Severity:  "warning",
					},
				},
				{
					Metric: metric{
						AlertName: "alert3",
						Severity:  "critical",
					},
				},
			},
			Provider:    "ocm",
			Environment: "prod",
			Expected:    true,
		},
		{
			Name: "no critical",
			Results: []result{
				{
					Metric: metric{
						AlertName: "alert1",
						Severity:  "info",
					},
				},
				{
					Metric: metric{
						AlertName: "alert2",
						Severity:  "warning",
					},
				},
				{
					Metric: metric{
						AlertName: "alert3",
						Severity:  "warning",
					},
				},
			},
			Provider:    "ocm",
			Environment: "prod",
			Expected:    false,
		},
		{
			Name: "ignored critical",
			Results: []result{
				{
					Metric: metric{
						AlertName: "alert1",
						Severity:  "info",
					},
				},
				{
					Metric: metric{
						AlertName: "alert2",
						Severity:  "warning",
					},
				},
				{
					Metric: metric{
						AlertName: "MetricsClientSendFailingSRE",
						Severity:  "critical",
					},
				},
			},
			Provider:    "ocm",
			Environment: "int",
			Expected:    false,
		},
		{
			Name: "found critical ignored in other environment",
			Results: []result{
				{
					Metric: metric{
						AlertName: "alert1",
						Severity:  "info",
					},
				},
				{
					Metric: metric{
						AlertName: "alert2",
						Severity:  "warning",
					},
				},
				{
					Metric: metric{
						AlertName: "MetricsClientSendFailingSRE",
						Severity:  "critical",
					},
				},
			},
			Provider:    "ocm",
			Environment: "prod",
			Expected:    true,
		},
		{
			Name: "found critical ignored in other provider",
			Results: []result{
				{
					Metric: metric{
						AlertName: "alert1",
						Severity:  "info",
					},
				},
				{
					Metric: metric{
						AlertName: "alert2",
						Severity:  "warning",
					},
				},
				{
					Metric: metric{
						AlertName: "MetricsClientSendFailingSRE",
						Severity:  "critical",
					},
				},
			},
			Provider:    "other-provider",
			Environment: "int",
			Expected:    true,
		},
	}

	for _, test := range tests {
		if findCriticalAlerts(test.Results, test.Provider, test.Environment) != test.Expected {
			t.Errorf("Test %s did not produce expected result (%t) for finding critical alerts", test.Name, test.Expected)
		}
	}
}
