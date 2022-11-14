package metrics

import (
	"testing"
	"time"

	"github.com/Masterminds/semver"
	"github.com/prometheus/common/model"
)

func TestSampleToJUnitResult(t *testing.T) {
	tests := []struct {
		name           string
		sample         *model.SampleStream
		expectedOutput JUnitResult
	}{
		{
			name: "regular parse",
			sample: &model.SampleStream{
				Metric: map[model.LabelName]model.LabelValue{
					"install_version": "openshift-v4.1.0",
					"upgrade_version": "",
					"cloud_provider":  "test",
					"environment":     "prod",
					"suite":           "test-suite",
					"testname":        "test-name",
					"result":          "passed",
					"cluster_id":      "1234567",
					"phase":           "install",
					"job":             "test-job1",
					"job_id":          "9999",
				},
				Values: []model.SamplePair{
					{
						Timestamp: 1,
						Value:     10,
					},
				},
			},
			expectedOutput: JUnitResult{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: nil,
				CloudProvider:  "test",
				Environment:    "prod",
				Suite:          "test-suite",
				TestName:       "test-name",
				Result:         Passed,
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Phase:          Install,
				Duration:       10 * time.Second,
				Timestamp:      1,
			},
		},
	}

	for _, test := range tests {
		jUnitResult, err := sampleToJUnitResult(test.sample)
		if err != nil {
			t.Errorf("test %s failed while converting the sample to JUnit result: %v", test.name, err)
		}

		if !jUnitResult.Equal(test.expectedOutput) {
			t.Errorf("test %s failed because the produced JUnit result %v does not match the expected output %v", test.name, jUnitResult, test.expectedOutput)
		}
	}
}

func TestCalculatePassRates(t *testing.T) {
	tests := []struct {
		name              string
		jUnitResults      []JUnitResult
		expectedPassRates map[string]float64
	}{
		{
			name: "one job pass rate all success",
			jUnitResults: []JUnitResult{
				makeJUnitResult("job1", Passed),
				makeJUnitResult("job1", Passed),
				makeJUnitResult("job1", Passed),
				makeJUnitResult("job1", Passed),
			},
			expectedPassRates: map[string]float64{
				"job1": 1.0,
			},
		},
		{
			name: "one job pass rate all success ignore skips",
			jUnitResults: []JUnitResult{
				makeJUnitResult("job1", Passed),
				makeJUnitResult("job1", Passed),
				makeJUnitResult("job1", Passed),
				makeJUnitResult("job1", Passed),
				makeJUnitResult("job1", Skipped),
				makeJUnitResult("job1", Skipped),
				makeJUnitResult("job1", Skipped),
			},
			expectedPassRates: map[string]float64{
				"job1": 1.0,
			},
		},
		{
			name: "one job pass rate all success ignore log metrics",
			jUnitResults: []JUnitResult{
				makeJUnitResult("job1", Passed),
				makeJUnitResult("job1", Passed),
				makeJUnitResult("job1", Passed),
				makeJUnitResult("job1", Passed),
				makeLogMetricResult("job1", Failed),
				makeLogMetricResult("job1", Failed),
			},
			expectedPassRates: map[string]float64{
				"job1": 1.0,
			},
		},
		{
			name: "one job pass rate partial success",
			jUnitResults: []JUnitResult{
				makeJUnitResult("job1", Failed),
				makeJUnitResult("job1", Passed),
				makeJUnitResult("job1", Passed),
				makeJUnitResult("job1", Passed),
			},
			expectedPassRates: map[string]float64{
				"job1": 0.75,
			},
		},
		{
			name: "one job pass rate partial success with upgrade failure",
			jUnitResults: []JUnitResult{
				makeJUnitResult("job1", Failed),
				makeJUnitResult("job1", Passed),
				makeJUnitResult("job1", Passed),
				makeJUnitResult("job1", Passed),
				makeUpgradeFailure("job1"),
			},
			expectedPassRates: map[string]float64{
				"job1": 0.375,
			},
		},
		{
			name: "one job pass rate partial success with upgrade failure but upgrade tests run anyway",
			jUnitResults: []JUnitResult{
				makeJUnitResult("job1", Failed),
				makeJUnitResult("job1", Passed),
				makeJUnitResult("job1", Passed),
				makeJUnitResult("job1", Passed),
				makeUpgradeFailure("job1"),
				makeUpgradeTest("job1", Passed),
				makeUpgradeTest("job1", Passed),
				makeUpgradeTest("job1", Failed),
				makeUpgradeTest("job1", Failed),
			},
			expectedPassRates: map[string]float64{
				"job1": 5.0 / 9.0,
			},
		},
		{
			name: "one job pass rate all failure",
			jUnitResults: []JUnitResult{
				makeJUnitResult("job1", Failed),
				makeJUnitResult("job1", Failed),
				makeJUnitResult("job1", Failed),
				makeJUnitResult("job1", Failed),
			},
			expectedPassRates: map[string]float64{
				"job1": 0,
			},
		},
		{
			name: "multiple jobs pass rate partial success with upgrade failures",
			jUnitResults: []JUnitResult{
				makeJUnitResult("job1", Failed),
				makeJUnitResult("job1", Passed),
				makeJUnitResult("job1", Passed),
				makeJUnitResult("job1", Passed),
				makeUpgradeFailure("job1"),
				makeJUnitResult("job2", Failed),
				makeJUnitResult("job2", Failed),
				makeJUnitResult("job2", Passed),
				makeJUnitResult("job2", Passed),
				makeUpgradeFailure("job2"),
			},
			expectedPassRates: map[string]float64{
				"job1": 0.375,
				"job2": 0.25,
			},
		},
		{
			name: "all skipped results",
			jUnitResults: []JUnitResult{
				makeJUnitResult("job1", Skipped),
				makeJUnitResult("job1", Skipped),
				makeJUnitResult("job1", Skipped),
				makeJUnitResult("job1", Skipped),
				makeJUnitResult("job2", Skipped),
				makeJUnitResult("job2", Skipped),
				makeJUnitResult("job2", Skipped),
				makeJUnitResult("job2", Skipped),
			},
			expectedPassRates: map[string]float64{
				"job1": 0,
				"job2": 0,
			},
		},
		{
			name:         "no results",
			jUnitResults: []JUnitResult{},
			expectedPassRates: map[string]float64{
				"job1": 0,
				"job2": 0,
			},
		},
	}

	for _, test := range tests {
		passRates := calculatePassRates(test.jUnitResults)

		for jobName, expectedPassRate := range test.expectedPassRates {
			if passRates[jobName] != expectedPassRate {
				t.Errorf("test %s failed because the produced pass rates %v does not match the expected output %v", test.name, passRates, test.expectedPassRates)
			}
		}
	}
}

func makeJUnitResult(jobName string, result Result) JUnitResult {
	return JUnitResult{
		InstallVersion: semver.MustParse("4.1.0"),
		UpgradeVersion: nil,
		CloudProvider:  "test",
		Environment:    "prod",
		Suite:          "test-suite",
		TestName:       "test-name",
		Result:         result,
		ClusterID:      "1234567",
		JobName:        jobName,
		JobID:          9999,
		Phase:          Install,
		Duration:       10 * time.Second,
		Timestamp:      1,
	}
}

func makeLogMetricResult(jobName string, result Result) JUnitResult {
	jUnitResult := makeJUnitResult(jobName, result)

	jUnitResult.TestName = "[Log Metrics] test-name"

	return jUnitResult
}

func makeUpgradeFailure(jobName string) JUnitResult {
	jUnitResult := makeJUnitResult(jobName, Failed)

	jUnitResult.TestName = "[upgrade] BeforeSuite"

	return jUnitResult
}

func makeUpgradeTest(jobName string, result Result) JUnitResult {
	jUnitResult := makeJUnitResult(jobName, result)

	jUnitResult.TestName = "[upgrade] test-name"

	return jUnitResult
}
