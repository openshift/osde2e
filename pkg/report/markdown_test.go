package report

import (
	"bytes"
	"testing"
	"time"

	"k8s.io/test-infra/testgrid/metadata/junit"
)

func TestTemplateReport(t *testing.T) {
	now := time.Now().UTC()
	rng := TimeRange{
		Start: now.Add(-5 * time.Minute),
		End:   now,
	}

	report := &Report{
		Title: "osde2e Triage",
		Range: rng,
		Envs: []Env{
			{
				Name: "int",
			},
			{
				Name: "stage",
				Jobs: []Job{
					{
						Name: "4.1",
						Runs: []Run{
							{
								BuildNum: 331,
								Failures: []Failure{
									{
										junit.Result{
											Name: "BeforeSuite",
										},
									},
								},
							},
							{
								BuildNum: 334,
								Failures: []Failure{
									{
										junit.Result{
											Name: "BeforeSuite",
										},
									},
								},
							},
							{
								BuildNum: 335,
								Failures: []Failure{
									{
										junit.Result{
											Name: "BeforeSuite",
										},
									},
								},
							},
						},
					},
					{
						Name: "4.2",
					},
					{
						Name: "upgrade 4.1-4.1",
						Runs: []Run{
							{
								BuildNum: 3,
								Failures: []Failure{
									{
										junit.Result{
											Name: "BeforeSuite",
										},
									},
								},
							},
							{
								BuildNum: 4,
								Failures: []Failure{
									{
										junit.Result{
											Name: "BeforeSuite",
										},
									},
								},
							},
							{
								BuildNum: 6,
								Failures: []Failure{
									{
										junit.Result{
											Name: "BeforeSuite",
										},
									},
								},
							},
							{
								BuildNum: 7,
								Failures: []Failure{
									{
										junit.Result{
											Name: "BeforeSuite",
										},
									},
								},
							},
						},
					},
				},
			},
			{
				Name: "prod",
			},
		},
	}

	var buf bytes.Buffer
	if err := report.Markdown(&buf); err != nil {
		t.Fatal(err)
	}
	t.Log(buf.String())
}
