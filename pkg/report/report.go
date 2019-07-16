package report

import (
	"fmt"
	"path/filepath"
	"time"

	testgrid "k8s.io/test-infra/testgrid/metadata"
	"k8s.io/test-infra/testgrid/metadata/junit"

	"github.com/openshift/osde2e/pkg/config"
)

// Report shows the results of jobs across multiple environments.
type Report struct {
	Config Config
	Title  string
	Range  TimeRange
	Envs   []Env
}

// Update refreshes the data of a report within rng. It
func (r *Report) Update(cfg *config.Config, rng TimeRange) error {
	if rng.Start.Before(r.Range.Start) {
		return fmt.Errorf("requested range %v is before report range %v", rng, r.Range)
	} else if r.Range.Start == r.Range.End {
		r.Range = rng
	}

	for _, envCfg := range r.Config.Envs {
		envPos := r.EnvPos(envCfg.Name)

		// create if environment doesn't exist in report
		if envPos < 0 {
			r.Envs = append(r.Envs, Env{
				Name: envCfg.Name,
			})
			envPos = len(r.Envs) - 1
		}

		for _, jobCfg := range r.Config.Jobs {
			if envCfg.SkipJob(jobCfg.Name) {
				continue
			}

			fullJobName := fmt.Sprintf("%s-%s-%s", jobCfg.Name, envCfg.Name, jobCfg.Version)

			// create if job doesn't exist in report
			jobPos := r.Envs[envPos].JobPos(fullJobName)
			if jobPos < 0 {
				r.Envs[envPos].Jobs = append(r.Envs[envPos].Jobs, Job{
					Name: fullJobName,
				})
				jobPos = len(r.Envs[envPos].Jobs) - 1
			}

			// determine prefix for test results
			r.Envs[envPos].Jobs[jobPos].Prefix = filepath.Join(cfg.TestGridPrefix, fullJobName)

			// if more recent failure than earliest use then as earliest requested
			earliest := r.Range.Start
			if len(r.Envs[envPos].Jobs[jobPos].Runs) != 0 {
				lastStart := time.Unix(r.Envs[envPos].Jobs[jobPos].Runs[0].Started.Timestamp, 0)
				if lastStart.After(earliest) {
					earliest = lastStart.Add(1)
				}
			}

			// update failures for job
			failures := r.GetRuns(r.Envs[envPos].Jobs[jobPos].Prefix, earliest, cfg)
			r.Envs[envPos].Jobs[jobPos].Runs = append(failures, r.Envs[envPos].Jobs[jobPos].Runs...)
		}
	}
	return nil
}

// EnvPos returns the position in a report for the given envName. Returns -1 if not found.
func (r Report) EnvPos(envName string) int {
	for i, env := range r.Envs {
		if env.Name == envName {
			return i
		}
	}
	return -1
}

// TimeRange is a period being reported on.
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// Env contains the results for a specific environment.
type Env struct {
	Name string
	Jobs []Job
}

// JobPos returns the position in an environment for the given jobName. Returns -1 if not found.
func (e Env) JobPos(jobName string) int {
	for i, job := range e.Jobs {
		if job.Name == jobName {
			return i
		}
	}
	return -1
}

// Job contains the results for a specific job.
type Job struct {
	Name   string
	Prefix string
	Runs   []Run
}

// Run contains the results for a run within a specific job.
type Run struct {
	BuildNum   int
	HiveLogURL string

	Started  testgrid.Started
	Finished testgrid.Finished

	Failures []Failure
}

// Failure contains an individual failing test.
type Failure struct {
	junit.Result
}
