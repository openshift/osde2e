// Package aggregator provides metadata extraction functionality
// for osde2e CI artifacts.
package aggregator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/go-logr/logr"
)

// extractMetadata extracts all metadata from collected log entries
func extractMetadata(logs []LogEntry, logger logr.Logger) map[string]any {
	metadata := map[string]any{
		"Environment": "ci", // Default
	}

	// Create quick lookup map
	fileMap := make(map[string]string)
	for _, log := range logs {
		fileMap[filepath.Base(log.Source)] = log.Source
	}

	// Extract from prowjob.json
	if path, exists := fileMap["prowjob.json"]; exists {
		extractProwJobData(path, metadata, logger)
	}

	// Extract timing data
	extractTimingData(fileMap["started.json"], fileMap["finished.json"], metadata, logger)

	// Extract cluster info from logs
	extractClusterData(logs, metadata, logger)

	// Extract environment details
	extractEnvironmentData(logs, metadata, logger)

	logger.Info("extracted metadata", "fields", len(metadata))
	return metadata
}

// extractProwJobData extracts from prowjob.json
func extractProwJobData(path string, metadata map[string]any, logger logr.Logger) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var prowJob map[string]any
	if json.Unmarshal(data, &prowJob) != nil {
		return
	}

	// Extract spec fields
	if spec, ok := prowJob["spec"].(map[string]any); ok {
		metadata["JobName"] = spec["job"]
		metadata["JobType"] = spec["type"]
		metadata["BuildCluster"] = spec["cluster"]
		metadata["Context"] = spec["context"]

		// Extract refs
		if refs, ok := spec["refs"].(map[string]any); ok {
			metadata["Repository"] = refs["org"].(string) + "/" + refs["repo"].(string)
			metadata["BaseRef"] = refs["base_ref"]
			metadata["BaseSHA"] = refs["base_sha"]

			// Extract pull request info
			if pulls, ok := refs["pulls"].([]any); ok && len(pulls) > 0 {
				if pull, ok := pulls[0].(map[string]any); ok {
					metadata["PullNumber"] = pull["number"]
					metadata["PullAuthor"] = pull["author"]
					metadata["PullTitle"] = pull["title"]
					metadata["PullSHA"] = pull["sha"]
					metadata["PullLink"] = pull["link"]
					metadata["PullBranch"] = pull["head_ref"]
					metadata["AuthorLink"] = pull["author_link"]
				}
			}
		}
	}

	// Extract status fields
	if status, ok := prowJob["status"].(map[string]any); ok {
		metadata["JobID"] = status["build_id"]
		metadata["JobState"] = status["state"]
		metadata["Description"] = status["description"]
		metadata["JobURL"] = status["url"]
	}

	logger.Info("extracted prowjob metadata")
}

// extractTimingData extracts from started.json and finished.json
func extractTimingData(startedPath, finishedPath string, metadata map[string]any, logger logr.Logger) {
	// Extract start time
	if startedPath != "" {
		if data, err := os.ReadFile(startedPath); err == nil {
			var started map[string]any
			if json.Unmarshal(data, &started) == nil {
				if ts, ok := started["timestamp"].(float64); ok {
					startTime := time.Unix(int64(ts), 0)
					metadata["StartTime"] = startTime.Format(time.RFC3339)
					metadata["StartTimestamp"] = int64(ts)
				}
				if pull, ok := started["pull"].(string); ok && pull != "" {
					metadata["PullRequest"] = pull
				}
			}
		}
	}

	// Extract completion time and result
	if finishedPath != "" {
		if data, err := os.ReadFile(finishedPath); err == nil {
			var finished map[string]any
			if json.Unmarshal(data, &finished) == nil {
				if ts, ok := finished["timestamp"].(float64); ok {
					completionTime := time.Unix(int64(ts), 0)
					metadata["CompletionTime"] = completionTime.Format(time.RFC3339)
					metadata["CompletionTimestamp"] = int64(ts)

					// Calculate duration
					if startTs, exists := metadata["StartTimestamp"].(int64); exists {
						duration := completionTime.Sub(time.Unix(startTs, 0))
						metadata["Duration"] = duration.String()
						metadata["DurationMinutes"] = int(duration.Minutes())
					}
				}
				metadata["JobPassed"] = finished["passed"]
				metadata["JobResult"] = finished["result"]

				// Extract nested metadata
				if meta, ok := finished["metadata"].(map[string]any); ok {
					metadata["WorkNamespace"] = meta["work-namespace"]
					metadata["PodName"] = meta["pod"]
				}
			}
		}
	}

	logger.Info("extracted timing metadata")
}

// extractClusterData extracts cluster info from log files
func extractClusterData(logs []LogEntry, metadata map[string]any, logger logr.Logger) {
	for _, log := range logs {
		fileName := filepath.Base(log.Source)

		// Only check relevant log files
		if !strings.Contains(fileName, "ci-operator.log") &&
			!strings.Contains(fileName, "build-log.txt") &&
			!strings.Contains(fileName, "test_output.log") {
			continue
		}

		data, err := os.ReadFile(log.Source)
		if err != nil {
			continue
		}
		content := string(data)

		// Use very specific patterns that match real formats
		if region := regexp.MustCompile(`AWS_REGION[=:\s]+([a-z]+-[a-z]+-\d+)`).FindStringSubmatch(content); len(region) > 1 {
			metadata["Region"] = region[1]
			metadata["Provider"] = "aws"
		}

		if region := regexp.MustCompile(`gcp[_-]region[=:\s]+([a-z]+-[a-z]+\d*-[a-z])`).FindStringSubmatch(content); len(region) > 1 {
			metadata["Region"] = region[1]
			metadata["Provider"] = "gcp"
		}

		if version := regexp.MustCompile(`version[=:\s]+([0-9]+\.[0-9]+\.[0-9]+)`).FindStringSubmatch(content); len(version) > 1 {
			metadata["Version"] = version[1]
		}

		// Look for actual cluster name patterns (osde2e-xxxxx format)
		if name := regexp.MustCompile(`osde2e-[a-z0-9]{5,}`).FindString(content); name != "" {
			metadata["ClusterName"] = name
		}
	}

	// Check terraform state for provider info
	for _, log := range logs {
		if strings.Contains(log.Source, "terraform.tfstate") {
			if data, err := os.ReadFile(log.Source); err == nil {
				var tfState map[string]any
				if json.Unmarshal(data, &tfState) == nil {
					if resources, ok := tfState["resources"].([]any); ok {
						for _, resource := range resources {
							if resMap, ok := resource.(map[string]any); ok {
								if provider, ok := resMap["provider"].(string); ok {
									if strings.Contains(provider, "aws") {
										metadata["Provider"] = "aws"
									} else if strings.Contains(provider, "google") {
										metadata["Provider"] = "gcp"
									}
								}
							}
						}
					}
				}
			}
			break
		}
	}

	logger.Info("extracted cluster metadata")
}

// extractEnvironmentData extracts environment and additional metadata
func extractEnvironmentData(logs []LogEntry, metadata map[string]any, logger logr.Logger) {
	// Check artifacts metadata
	for _, log := range logs {
		if strings.Contains(log.Source, "artifacts/metadata.json") {
			if data, err := os.ReadFile(log.Source); err == nil {
				var artifactsMetadata map[string]any
				if json.Unmarshal(data, &artifactsMetadata) == nil {
					if workNamespace, ok := artifactsMetadata["work-namespace"].(string); ok {
						metadata["WorkNamespace"] = workNamespace
					}
					if pod, ok := artifactsMetadata["pod"].(string); ok {
						metadata["Pod"] = pod
					}
				}
			}
			break
		}
	}

	// Determine environment from job name
	if jobName, exists := metadata["JobName"]; exists {
		jobNameStr := jobName.(string)
		if strings.Contains(jobNameStr, "prod") {
			metadata["Environment"] = "production"
		} else if strings.Contains(jobNameStr, "stage") {
			metadata["Environment"] = "staging"
		} else if strings.Contains(jobNameStr, "int") {
			metadata["Environment"] = "integration"
		}
	}

	// Set failure time if job failed
	if jobState, exists := metadata["JobState"]; exists && jobState == "failure" {
		if completionTime, exists := metadata["CompletionTime"]; exists {
			metadata["FailureTime"] = completionTime
		}
	}

	logger.Info("extracted environment metadata")
}
