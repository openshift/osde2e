package metrics

import (
	"testing"

	"github.com/Masterminds/semver"
	"github.com/prometheus/common/model"
)

func TestExtractMetricFromSample(t *testing.T) {
	tests := []struct {
		name           string
		sample         *model.SampleStream
		metricName     model.LabelName
		expectedOutput string
	}{
		{
			name: "extract metric from sample",
			sample: &model.SampleStream{
				Metric: map[model.LabelName]model.LabelValue{
					"blah": "value",
				},
			},
			metricName:     "blah",
			expectedOutput: "value",
		},
		{
			name: "extract non-existent metric from sample",
			sample: &model.SampleStream{
				Metric: map[model.LabelName]model.LabelValue{
					"blah": "value",
				},
			},
			metricName:     "doesnt-exist",
			expectedOutput: "",
		},
	}

	for _, test := range tests {
		retrievedMetric := extractMetricFromSample(test.sample, test.metricName)

		if retrievedMetric != test.expectedOutput {
			t.Errorf("test %s failed as retrieved metric value %s did not match the expected metric value %s", test.name, retrievedMetric, test.expectedOutput)
		}
	}
}

func TestExtractInstallAndUpgradeVersionsFromSample(t *testing.T) {
	tests := []struct {
		name                   string
		sample                 *model.SampleStream
		expectedInstallVersion *semver.Version
		expectedUpgradeVersion *semver.Version
		expectError            bool
	}{
		{
			name: "extract install and upgrade versions from sample",
			sample: &model.SampleStream{
				Metric: map[model.LabelName]model.LabelValue{
					"install_version": "openshift-v4.1.0",
					"upgrade_version": "openshift-v4.1.2",
				},
			},
			expectedInstallVersion: semver.MustParse("4.1.0"),
			expectedUpgradeVersion: semver.MustParse("4.1.2"),
			expectError:            false,
		},
		{
			name: "extract install and empty upgrade versions from sample",
			sample: &model.SampleStream{
				Metric: map[model.LabelName]model.LabelValue{
					"install_version": "openshift-v4.1.0",
					"upgrade_version": "",
				},
			},
			expectedInstallVersion: semver.MustParse("4.1.0"),
			expectedUpgradeVersion: nil,
			expectError:            false,
		},
		{
			name: "extract empty install and empty upgrade versions from sample",
			sample: &model.SampleStream{
				Metric: map[model.LabelName]model.LabelValue{
					"install_version": "",
					"upgrade_version": "",
				},
			},
			expectError: true,
		},
		{
			name: "extract malformed install versions",
			sample: &model.SampleStream{
				Metric: map[model.LabelName]model.LabelValue{
					"install_version": "un-parseable",
					"upgrade_version": "",
				},
			},
			expectError: true,
		},
		{
			name: "extract malformed upgrade versions",
			sample: &model.SampleStream{
				Metric: map[model.LabelName]model.LabelValue{
					"install_version": "openshift-v4.1.0",
					"upgrade_version": "un-parseable",
				},
			},
			expectError: true,
		},
	}

	for _, test := range tests {
		installVersion, upgradeVersion, err := extractInstallAndUpgradeVersionsFromSample(test.sample)

		if (err != nil) != test.expectError {
			t.Errorf("test %s failed as error %v was found and expected error setting is %t", test.name, err, test.expectError)
		}

		if !test.expectError {
			if !installVersion.Equal(test.expectedInstallVersion) {
				t.Errorf("test %s failed as produced install version %v does not match expected install version %v", test.name, installVersion, test.expectedInstallVersion)
			}

			if (upgradeVersion != nil && !upgradeVersion.Equal(test.expectedUpgradeVersion)) || (upgradeVersion == nil && upgradeVersion != test.expectedUpgradeVersion) {
				t.Errorf("test %s failed as produced upgrade version %v does not match expected upgrade version %v", test.name, upgradeVersion, test.expectedUpgradeVersion)
			}
		}
	}
}

func TestStringToResult(t *testing.T) {
	tests := []struct {
		name           string
		stringToParse  string
		expectedResult Result
	}{
		{
			name:           "passed",
			stringToParse:  "passed",
			expectedResult: Passed,
		},
		{
			name:           "passed case insensitive",
			stringToParse:  "PaSsEd",
			expectedResult: Passed,
		},
		{
			name:           "skipped",
			stringToParse:  "skipped",
			expectedResult: Skipped,
		},
		{
			name:           "skipped case insensitive",
			stringToParse:  "SkIpPeD",
			expectedResult: Skipped,
		},
		{
			name:           "failed",
			stringToParse:  "failed",
			expectedResult: Failed,
		},
		{
			name:           "failed case insensitive",
			stringToParse:  "FaIlEd",
			expectedResult: Failed,
		},
		{
			name:           "unknown",
			stringToParse:  "something else",
			expectedResult: UnknownResult,
		},
	}

	for _, test := range tests {
		parsedResult := stringToResult(test.stringToParse)

		if parsedResult != test.expectedResult {
			t.Errorf("test %s failed as the parsed result %s did not match the expected result %s", test.name, parsedResult, test.expectedResult)
		}
	}
}

func TestStringToPhase(t *testing.T) {
	tests := []struct {
		name          string
		stringToParse string
		expectedPhase Phase
	}{
		{
			name:          "install",
			stringToParse: "install",
			expectedPhase: Install,
		},
		{
			name:          "install case insensitive",
			stringToParse: "InStAlL",
			expectedPhase: Install,
		},
		{
			name:          "upgrade",
			stringToParse: "upgrade",
			expectedPhase: Upgrade,
		},
		{
			name:          "upgrade case insensitive",
			stringToParse: "UpGrAdE",
			expectedPhase: Upgrade,
		},
		{
			name:          "unknown",
			stringToParse: "something else",
			expectedPhase: UnknownPhase,
		},
	}

	for _, test := range tests {
		parsedPhase := stringToPhase(test.stringToParse)

		if parsedPhase != test.expectedPhase {
			t.Errorf("test %s failed as the parsed phase %s did not match the expected result %s", test.name, parsedPhase, test.expectedPhase)
		}
	}
}

func TestEscapeQuotes(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedOutput string
	}{
		{
			name:           "no quotes",
			input:          `some string`,
			expectedOutput: `some string`,
		},
		{
			name:           "some quotes",
			input:          `some "string"`,
			expectedOutput: `some \"string\"`,
		},
		{
			name:           "all quotes",
			input:          `""""""""`,
			expectedOutput: `\"\"\"\"\"\"\"\"`,
		},
	}

	for _, test := range tests {
		escapedString := escapeQuotes(test.input)

		if escapedString != test.expectedOutput {
			t.Errorf("test %s failed as the escaped string %s did not match the expected result %s", test.name, escapedString, test.expectedOutput)
		}
	}
}

func TestAverageValues(t *testing.T) {
	tests := []struct {
		name           string
		samplePairs    []model.SamplePair
		expectedOutput float64
	}{
		{
			name: "average a single value",
			samplePairs: []model.SamplePair{
				{
					Timestamp: 1,
					Value:     16,
				},
			},
			expectedOutput: 16,
		},
		{
			name: "average a multiple values",
			samplePairs: []model.SamplePair{
				{
					Timestamp: 1,
					Value:     16,
				},
				{
					Timestamp: 1,
					Value:     32,
				},
				{
					Timestamp: 1,
					Value:     48,
				},
			},
			expectedOutput: 32,
		},
		{
			name:           "average a zero values",
			samplePairs:    []model.SamplePair{},
			expectedOutput: 0,
		},
	}

	for _, test := range tests {
		average := averageValues(test.samplePairs)

		if average != test.expectedOutput {
			t.Errorf("test %s failed as the produced average %f did not match the expected average %f", test.name, average, test.expectedOutput)
		}
	}
}

func TestPickFirstTimestamp(t *testing.T) {
	tests := []struct {
		name           string
		samplePairs    []model.SamplePair
		expectedOutput int64
	}{
		{
			name: "pick a single value",
			samplePairs: []model.SamplePair{
				{
					Timestamp: 1,
					Value:     16,
				},
			},
			expectedOutput: 1,
		},
		{
			name: "pick from multiple v alues",
			samplePairs: []model.SamplePair{
				{
					Timestamp: 1,
					Value:     16,
				},
				{
					Timestamp: 2,
					Value:     32,
				},
				{
					Timestamp: 3,
					Value:     48,
				},
			},
			expectedOutput: 1,
		},
		{
			name:           "pick from no values",
			samplePairs:    []model.SamplePair{},
			expectedOutput: 0,
		},
	}

	for _, test := range tests {
		timestamp := pickFirstTimestamp(test.samplePairs)

		if timestamp != test.expectedOutput {
			t.Errorf("test %s failed as the produced timestamp %d did not match the expected timestamp %d", test.name, timestamp, test.expectedOutput)
		}
	}
}
