// Package store provides a SQLite-backed persistence layer for pipeline results.
// It is written to by the SQS consumer (incremental) and the backfill job (bulk),
// and read by the dashboard HTTP handlers for sub-millisecond page loads.
package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite" // pure-Go SQLite driver, no CGO required

	"github.com/openshift/osde2e/pkg/dashboard/models"
)

const schema = `
PRAGMA journal_mode=WAL;
PRAGMA foreign_keys=ON;

-- Latest result per (operator, env) — used by the Pipelines overview table.
CREATE TABLE IF NOT EXISTS pipeline_latest (
    operator_name TEXT NOT NULL,
    env           TEXT NOT NULL,
    version       TEXT NOT NULL DEFAULT 'unknown',
    status        TEXT NOT NULL DEFAULT 'unknown',
    passed        INTEGER NOT NULL DEFAULT 0,
    failed        INTEGER NOT NULL DEFAULT 0,
    total         INTEGER NOT NULL DEFAULT 0,
    job_id        TEXT NOT NULL DEFAULT '',
    last_run      DATETIME NOT NULL,
    log_url       TEXT NOT NULL DEFAULT '',
    junit_url     TEXT NOT NULL DEFAULT '',
    failed_tests  TEXT NOT NULL DEFAULT '[]', -- JSON []FailedTestCase
    llm_analysis  TEXT NOT NULL DEFAULT '',   -- JSON LLMAnalysis or empty
    PRIMARY KEY (operator_name, env)
);

-- Every individual run — used by the pipeline-detail history page.
CREATE TABLE IF NOT EXISTS pipeline_runs (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    operator_name TEXT NOT NULL,
    env           TEXT NOT NULL,
    version       TEXT NOT NULL DEFAULT 'unknown',
    status        TEXT NOT NULL DEFAULT 'unknown',
    passed        INTEGER NOT NULL DEFAULT 0,
    failed        INTEGER NOT NULL DEFAULT 0,
    total         INTEGER NOT NULL DEFAULT 0,
    job_id        TEXT NOT NULL DEFAULT '',
    date          TEXT NOT NULL DEFAULT '',
    last_run      DATETIME NOT NULL,
    log_url       TEXT NOT NULL DEFAULT '',
    junit_url     TEXT NOT NULL DEFAULT '',
    failed_tests  TEXT NOT NULL DEFAULT '[]', -- JSON []FailedTestCase
    llm_analysis  TEXT NOT NULL DEFAULT '',   -- JSON LLMAnalysis or empty
    UNIQUE (operator_name, env, job_id)       -- deduplicate on re-process
);

CREATE INDEX IF NOT EXISTS idx_runs_operator ON pipeline_runs (operator_name, last_run DESC);

-- Migration: add llm_analysis column to existing DBs that predate this field.
-- SQLite ignores "duplicate column" errors but this pattern avoids them.
`

// Store wraps the SQLite database connection and provides typed query methods.
type Store struct {
	db *sql.DB
}

// Open opens (or creates) the SQLite database at path and applies the schema.
// Use ":memory:" for an in-memory database (useful for tests).
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite %s: %w", path, err)
	}

	// SQLite performs best with a single writer connection.
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}

	// Best-effort migrations for existing databases missing the llm_analysis column.
	for _, tbl := range []string{"pipeline_latest", "pipeline_runs"} {
		_, _ = db.Exec(`ALTER TABLE ` + tbl + ` ADD COLUMN llm_analysis TEXT NOT NULL DEFAULT ''`)
	}

	log.Printf("Store: opened SQLite at %s", path)
	return &Store{db: db}, nil
}

// Close closes the underlying database connection.
func (s *Store) Close() error { return s.db.Close() }

// RunRecord is the flat struct used when writing to the store.
type RunRecord struct {
	OperatorName string
	Env          string
	Version      string
	Status       string
	Passed       int
	Failed       int
	Total        int
	JobID        string
	Date         string
	LastRun      time.Time
	LogURL       string
	JUnitURL     string
	FailedTests  []models.FailedTestCase
	LLMAnalysis  *models.LLMAnalysis
}

// UpsertRun inserts or updates both pipeline_latest and pipeline_runs for one run result.
func (s *Store) UpsertRun(r RunRecord) error {
	ft, err := json.Marshal(r.FailedTests)
	if err != nil {
		return fmt.Errorf("marshal failed_tests: %w", err)
	}

	llmStr := ""
	if r.LLMAnalysis != nil {
		b, err := json.Marshal(r.LLMAnalysis)
		if err != nil {
			return fmt.Errorf("marshal llm_analysis: %w", err)
		}
		llmStr = string(b)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Upsert pipeline_latest — only overwrite if this run is newer.
	_, err = tx.Exec(`
		INSERT INTO pipeline_latest
			(operator_name, env, version, status, passed, failed, total, job_id, last_run, log_url, junit_url, failed_tests, llm_analysis)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(operator_name, env) DO UPDATE SET
			version      = excluded.version,
			status       = excluded.status,
			passed       = excluded.passed,
			failed       = excluded.failed,
			total        = excluded.total,
			job_id       = excluded.job_id,
			last_run     = excluded.last_run,
			log_url      = excluded.log_url,
			junit_url    = excluded.junit_url,
			failed_tests = excluded.failed_tests,
			llm_analysis = excluded.llm_analysis
		WHERE excluded.last_run > pipeline_latest.last_run
	`,
		r.OperatorName, r.Env, r.Version, r.Status,
		r.Passed, r.Failed, r.Total,
		r.JobID, r.LastRun, r.LogURL, r.JUnitURL,
		string(ft), llmStr,
	)
	if err != nil {
		return fmt.Errorf("upsert pipeline_latest: %w", err)
	}

	// Insert pipeline_runs — ignore duplicate job_id.
	_, err = tx.Exec(`
		INSERT OR IGNORE INTO pipeline_runs
			(operator_name, env, version, status, passed, failed, total, job_id, date, last_run, log_url, junit_url, failed_tests, llm_analysis)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		r.OperatorName, r.Env, r.Version, r.Status,
		r.Passed, r.Failed, r.Total,
		r.JobID, r.Date, r.LastRun, r.LogURL, r.JUnitURL,
		string(ft), llmStr,
	)
	if err != nil {
		return fmt.Errorf("insert pipeline_runs: %w", err)
	}

	return tx.Commit()
}

// GetLatest returns all rows from pipeline_latest as []models.OperatorStatus,
// grouped by operator name (one entry per operator, results keyed by env).
func (s *Store) GetLatest() ([]models.OperatorStatus, error) {
	rows, err := s.db.Query(`
		SELECT operator_name, env, version, status, passed, failed, total,
		       job_id, last_run, log_url, junit_url, failed_tests, llm_analysis
		FROM pipeline_latest
		ORDER BY operator_name, env
	`)
	if err != nil {
		return nil, fmt.Errorf("query pipeline_latest: %w", err)
	}
	defer rows.Close()

	index := make(map[string]*models.OperatorStatus)
	var order []string

	for rows.Next() {
		var (
			name, env, ver, status  string
			passed, failed, total   int
			jobID, logURL, junitURL string
			lastRun                 time.Time
			ftJSON, llmJSON         string
		)
		if err := rows.Scan(&name, &env, &ver, &status, &passed, &failed, &total,
			&jobID, &lastRun, &logURL, &junitURL, &ftJSON, &llmJSON); err != nil {
			return nil, fmt.Errorf("scan pipeline_latest: %w", err)
		}

		var failedTests []models.FailedTestCase
		_ = json.Unmarshal([]byte(ftJSON), &failedTests)

		var llm *models.LLMAnalysis
		if llmJSON != "" {
			llm = &models.LLMAnalysis{}
			if err := json.Unmarshal([]byte(llmJSON), llm); err != nil {
				llm = nil
			}
		}

		er := &models.EnvironmentResult{
			Version:     ver,
			Status:      status,
			Passed:      passed,
			Failed:      failed,
			Total:       total,
			JobID:       jobID,
			LastRun:     lastRun,
			LogURL:      logURL,
			JUnitURL:    junitURL,
			FailedTests: failedTests,
			LLMAnalysis: llm,
		}

		op, ok := index[name]
		if !ok {
			op = &models.OperatorStatus{
				Name:    name,
				Results: make(map[string]*models.EnvironmentResult),
			}
			index[name] = op
			order = append(order, name)
		}
		op.Results[env] = er
		if lastRun.After(op.LastUpdated) {
			op.LastUpdated = lastRun
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := make([]models.OperatorStatus, 0, len(order))
	for _, name := range order {
		result = append(result, *index[name])
	}
	return result, nil
}

// GetHistory returns all pipeline_runs for a given operator, newest first.
func (s *Store) GetHistory(operatorName string) (*models.PipelineHistory, error) {
	rows, err := s.db.Query(`
		SELECT env, version, status, passed, failed, total,
		       job_id, date, last_run, log_url, junit_url, failed_tests, llm_analysis
		FROM pipeline_runs
		WHERE operator_name = ?
		ORDER BY last_run DESC
	`, operatorName)
	if err != nil {
		return nil, fmt.Errorf("query pipeline_runs: %w", err)
	}
	defer rows.Close()

	var runs []models.PipelineRun
	for rows.Next() {
		var (
			env, ver, status        string
			passed, failed, total   int
			jobID, date             string
			logURL, junitURL        string
			lastRun                 time.Time
			ftJSON, llmJSON         string
		)
		if err := rows.Scan(&env, &ver, &status, &passed, &failed, &total,
			&jobID, &date, &lastRun, &logURL, &junitURL, &ftJSON, &llmJSON); err != nil {
			return nil, fmt.Errorf("scan pipeline_runs: %w", err)
		}

		var failedTests []models.FailedTestCase
		_ = json.Unmarshal([]byte(ftJSON), &failedTests)

		var llm *models.LLMAnalysis
		if llmJSON != "" {
			llm = &models.LLMAnalysis{}
			if err := json.Unmarshal([]byte(llmJSON), llm); err != nil {
				llm = nil
			}
		}

		runs = append(runs, models.PipelineRun{
			Env:         env,
			Version:     ver,
			Status:      status,
			Passed:      passed,
			Total:       total,
			JobID:       jobID,
			Date:        date,
			LastRun:     lastRun,
			LogURL:      logURL,
			JUnitURL:    junitURL,
			Failed:      failedTests,
			LLMAnalysis: llm,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Group runs by version, preserving newest-first order per version.
	// For each version we keep the newest run per env (int takes precedence over stage over prod).
	type versionKey = string
	versionOrder := []versionKey{}
	versionMap := make(map[versionKey]*models.VersionPipeline)

	for i := range runs {
		run := &runs[i]
		vp, exists := versionMap[run.Version]
		if !exists {
			vp = &models.VersionPipeline{
				Version: run.Version,
				Date:    run.Date,
				LastRun: run.LastRun,
				EnvRuns: make(map[string]*models.PipelineRun),
			}
			versionMap[run.Version] = vp
			versionOrder = append(versionOrder, run.Version)
		}
		// Keep newest run per env (runs are already newest-first so first wins).
		if _, seen := vp.EnvRuns[run.Env]; !seen {
			vp.EnvRuns[run.Env] = run
		}
		if run.LastRun.After(vp.LastRun) {
			vp.LastRun = run.LastRun
			vp.Date = run.Date
		}
	}

	versions := make([]models.VersionPipeline, 0, len(versionOrder))
	for _, ver := range versionOrder {
		versions = append(versions, *versionMap[ver])
	}

	return &models.PipelineHistory{
		OperatorName: operatorName,
		Runs:         runs,
		Versions:     versions,
	}, nil
}

// OperatorNames returns a sorted list of all distinct operator names in the store.
func (s *Store) OperatorNames() ([]string, error) {
	rows, err := s.db.Query(`SELECT DISTINCT operator_name FROM pipeline_latest ORDER BY operator_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		names = append(names, n)
	}
	return names, rows.Err()
}
