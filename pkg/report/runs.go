package report

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/openshift/osde2e/pkg/config"
	"github.com/openshift/osde2e/pkg/testgrid"
)

const (
	hiveLogName     = "hive-log.txt"
	artifactURLTmpl = "https://storage.googleapis.com/%s/%s/%d/%s/%s"
)

// GetRuns returns TestGrid build runs starting with prefix that are after earliest.
func (r Report) GetRuns(prefix string, earliest time.Time, cfg *config.Config) (runs []Run) {
	tg, err := testgrid.NewTestGrid(cfg.TestGridBucket, prefix, cfg.TestGridServiceAccount)
	if err != nil {
		log.Fatalf("Failed to setup TestGrid support: %v", err)
	}

	ctx := context.Background()
	started, latestBuildNum, err := tg.LatestStarted(ctx)
	if err != nil {
		log.Fatalf("Couldn't get latest record: %v", err)
	}

	// report on each build
	for i := latestBuildNum; i > 0; i-- {
		if i != latestBuildNum {
			if started, err = tg.Started(ctx, i); err != nil {
				log.Printf("Error getting started for build %d: %v", i, err)
				continue
			}
		}

		// only include events in the window
		startTime := time.Unix(started.Timestamp, 0)
		if startTime.Before(earliest) {
			break
		} else if startTime.After(r.Range.End) {
			continue
		}

		finished, err := tg.Finished(ctx, i)
		if err != nil {
			log.Printf("Error getting finished for build %d: %v", i, err)
		}

		suites, err := tg.Suites(ctx, i)
		if err != nil {
			log.Printf("Couldn't get suites for build %d: %v", i, err)
			continue
		}

		// add failure report for each failure in targetTests
		var failures []Failure
		for _, suite := range suites.Suites {
			for _, result := range suite.Results {
				for _, testName := range r.Config.Tests {
					if result.Name == testName {
						failures = append(failures, Failure{
							Result: result,
						})
					}
				}
			}
		}

		// include run in report if failures occurred
		if len(failures) != 0 {
			run := Run{
				BuildNum: i,

				Started:  started,
				Finished: finished,

				Failures: failures,
			}

			// check for hive logs
			hiveLogPrefix := fmt.Sprintf("%s/%s", testgrid.ArtifactsDir, hiveLogName)
			paths, err := tg.ListFiles(ctx, i, hiveLogPrefix, "")
			if err != nil {
				log.Printf("Encountered error checking for '%s' on build %d: %v", hiveLogPrefix, i, err)
			} else if len(paths) != 0 {
				run.HiveLogURL = fmt.Sprintf(artifactURLTmpl, cfg.TestGridBucket, prefix, i, testgrid.ArtifactsDir, hiveLogName)
			}

			// add run
			runs = append(runs, run)
		}
	}
	return
}
