package config

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

// LogMetrics is an array of LogMetric types with an easier lookup method
type LogMetrics []LogMetric

var once = sync.Once{}

var logMetrics = LogMetrics{}

var beforeSuiteMetrics = LogMetrics{}

// GetLogMetrics will return the log metrics.
func GetLogMetrics() LogMetrics {
	once.Do(func() {
		err := viper.UnmarshalKey(fmt.Sprintf("%slogMetrics", CloudProvider.CloudProviderID), &logMetrics)
		if err != nil {
			log.Printf("Log Metric Thresholds not set for the provider %s, defaulting to AWS settings", CloudProvider.CloudProviderID)
			viper.UnmarshalKey("awslogMetrics", &logMetrics)
		}
	})
	return logMetrics
}

// GetBeforeSuiteMetrics will return the log metrics.
func GetBeforeSuiteMetrics() LogMetrics {
	err := viper.UnmarshalKey(fmt.Sprintf("%sbeforeSuiteMetrics", CloudProvider.CloudProviderID), &beforeSuiteMetrics)
	if err != nil {
		log.Printf("Before Suite Metric Thresholds not set for the provider %s, defaulting to AWS settings", CloudProvider.CloudProviderID)
		viper.UnmarshalKey("awsbeforeSuiteMetrics", &beforeSuiteMetrics)
	}
	return beforeSuiteMetrics
}

// GetMetricByName returns a pointer to a LogMetric from the array based on the name
func (metrics LogMetrics) GetMetricByName(name string) *LogMetric {
	for k := range metrics {
		if name == metrics[k].Name {
			return &metrics[k]
		}
	}

	return &LogMetric{}
}

// LogMetric lets you define a metric name and a regex to apply on the build log
// For every match in the build log, the metric with that name will increment
type LogMetric struct {
	// Name of the metric to look for
	Name string `json:"name" yaml:"name"`
	// Regex (in string form) to parse out
	RegEx string `json:"regex" yaml:"regex"`
	// IgnoreIfMatchContains will ignore a match if the match contains any of the given strings.
	IgnoreIfMatchContains []string `json:"ignoreIfMatchContains" yaml:"ignoreIfMatchContains"`
	// High threshold before failing
	HighThreshold int `json:"highThreshold" yaml:"highThreshold" default:"9999"`
	// Low threshold before failing
	LowThreshold int `json:"lowThreshold" yaml:"lowThreshold" default:"-1"`
}

// HasMatches attempts to match the regex provided a bytearray and returns the number of matches
func (metric LogMetric) HasMatches(data []byte) int {
	regex := regexp.MustCompile(metric.RegEx)

	matches := []string{}

	buf := bytes.NewBuffer(data)
	dataReader := bufio.NewReader(buf)

	for {
		if line, _, err := dataReader.ReadLine(); err != io.EOF {
			if regex.Match(line) {
				matches = append(matches, string(line))
			}
		} else {
			break
		}
	}

	numMatches := 0
	for _, match := range matches {
		shouldCount := true
		for _, ignoreString := range metric.IgnoreIfMatchContains {
			if strings.Contains(string(match), ignoreString) {
				shouldCount = false
				break
			}
		}

		if shouldCount {
			numMatches++
		}
	}

	return numMatches
}

// IsPassing checks the current counter against the thresholds to see if this
// metric should be passing or failing via JUnit
func (metric LogMetric) IsPassing(value int) bool {
	return metric.HighThreshold > value && metric.LowThreshold < value
}
