package db_test

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/openshift/osde2e/pkg/db"
	"github.com/ory/dockertest"
)

var dbPool *dockertest.Pool = func() *dockertest.Pool {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	dbPool, err := dockertest.NewPool("")
	if err != nil {
		log.Printf("Could not connect to docker: %s", err)
		return nil
	} else if os.Getenv("FORCE_REAL_DB_TESTS") == "1" {
		return nil
	}
	return dbPool
}()

// dbConfig is a lightweight type that constructs connection strings. It isn't very
// smart, but is much easier to use than the smart options in other libraries. If
// it ever needs to be smarter, we should switch to those types instead.
type dbConfig struct {
	Host, Port, User, Pass, Database, Params string
}

// URL constructs a url from the config.
func (d dbConfig) URL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?%s",
		d.User, d.Pass, d.Host, d.Port, d.Database, d.Params)
}

// getDBConfig returns the database connection configuration for a database against
// which tests can run.
func getDBConfig(t *testing.T) dbConfig {
	if dbPool == nil {
		t.Skip()
		return dbConfig{}
	}
	_, err := dbPool.Client.Version()
	if err != nil && os.Getenv("PG_HOST") == "" {
		t.Skip()
		return dbConfig{}
	}
	if dbPool != nil {
		const password = "secret"
		// pulls an image, creates a container based on it and runs it
		resource, err := dbPool.Run("postgres", "12", []string{"POSTGRES_PASSWORD=" + password})
		if err != nil {
			t.Fatalf("Could not start resource: %s", err)
		}
		t.Cleanup(func() {
			if err := dbPool.Purge(resource); err != nil {
				log.Printf("Could not purge resource: %s", err)
			}
		})
		config := dbConfig{
			User:   "postgres",
			Pass:   password,
			Host:   "127.0.0.1",
			Port:   resource.GetPort("5432/tcp"),
			Params: "sslmode=disable",
		}

		url := config.URL()

		// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
		if err := dbPool.Retry(func() error {
			db, err := sql.Open("pgx", url)
			if err != nil {
				return err
			}
			return db.Ping()
		}); err != nil {
			t.Fatalf("Could not connect to postgres: %s", err)
		}
		return config
	}
	return dbConfig{
		User: os.Getenv("PG_USER"),
		Pass: os.Getenv("PG_PASS"),
		Host: os.Getenv("PG_HOST"),
		Port: os.Getenv("PG_PORT"),
	}
}

// tableNames returns the names of all existing tables (of type BASE TABLE) in the public
// postgres schema of the connected database.
func tableNames(pg *sql.DB) ([]string, error) {
	const q = `SELECT table_name
    		  FROM information_schema.tables
    		 WHERE table_schema='public'
    		   AND table_type='BASE TABLE';`
	rows, err := pg.Query(q)
	if err != nil {
		return nil, fmt.Errorf("failed listing table names: %w", err)
	}
	defer rows.Close()
	var tableNames []string
	for i := 0; rows.Next(); i++ {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed scanning row %d: %w", i, err)
		}
		tableNames = append(tableNames, name)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("failed closing rows: %w", err)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed checking error on rows: %w", err)
	}
	return tableNames, nil
}

// columnNames returns the names of all existing columns for a specified table
// postgres schema of the connected database.
func columnNames(pg *sql.DB, table string) ([]string, error) {
	const q = "select column_name from information_schema.columns where table_name = $1;"

	rows, err := pg.Query(q, table)
	if err != nil {
		return nil, fmt.Errorf("failed listing column names: %w", err)
	}
	defer rows.Close()
	var columnNames []string
	for i := 0; rows.Next(); i++ {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed scanning row %d: %w", i, err)
		}
		columnNames = append(columnNames, name)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("failed closing rows: %w", err)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed checking error on rows: %w", err)
	}
	return columnNames, nil
}

// ensureNotTables returns a function that will fail a test if any of the provided table
// names currently exists in the connected database.
func ensureNotTables(unexpectedNames ...string) func(pg *sql.DB, t *testing.T) {
	return func(pg *sql.DB, t *testing.T) {
		names, err := tableNames(pg)
		if err != nil {
			t.Fatalf("couldn't list table rows: %v", err)
		}
		existing := make(map[string]bool)
		for _, name := range names {
			existing[name] = true
		}
		for _, name := range unexpectedNames {
			if existing[name] {
				t.Errorf("did not expect table %s to exist", name)
			}
		}
	}
}

// ensureTables returns a function that will fail a test if any of the provided table names
// currently does not exist in the connected database.
func ensureTables(expectedNames ...string) func(pg *sql.DB, t *testing.T) {
	return func(pg *sql.DB, t *testing.T) {
		names, err := tableNames(pg)
		if err != nil {
			t.Fatalf("couldn't list table rows: %v", err)
		}
		existing := make(map[string]bool)
		for _, name := range names {
			existing[name] = true
		}
		for _, name := range expectedNames {
			if !existing[name] {
				t.Errorf("expected table %s to exist", name)
			}
		}
	}
}

// ensureColumns returns a function that will fail a test if any of the provided columns
// don't exist within a specified table in the connected database.
func ensureColumns(table string, expectedColumns ...string) func(pg *sql.DB, t *testing.T) {
	return func(pg *sql.DB, t *testing.T) {
		names, err := columnNames(pg, table)
		if err != nil {
			t.Fatalf("couldn't list column rows: %v", err)
		}
		existing := make(map[string]bool)
		for _, name := range names {
			existing[name] = true
		}
		for _, name := range expectedColumns {
			if !existing[name] {
				t.Errorf("expected column %s to exist", name)
			}
		}
	}
}

// ensureNotColumns returns a function that will fail a test if any of the provided columns
// don't exist within a specified table in the connected database.
func ensureNotColumns(table string, expectedColumns ...string) func(pg *sql.DB, t *testing.T) {
	return func(pg *sql.DB, t *testing.T) {
		names, err := columnNames(pg, table)
		if err != nil {
			t.Fatalf("couldn't list column rows: %v", err)
		}
		existing := make(map[string]bool)
		for _, name := range names {
			existing[name] = true
		}
		for _, name := range expectedColumns {
			if existing[name] {
				t.Errorf("did not expect column %s to exist", name)
			}
		}
	}
}

// migrationTestCase provides hooks for testing database migrations.
// The `preup` function will be run before the up migration is applied to ensure
// that state is as expected and to provide an opportunity to seed the database
// with data.
// the `during` function will be run after the up migration has been applied to
// allow checking for the effects of the migration.
// the `postdown` function will be run after the down migration has been applied
// to ensure that the migration cleaned up after itself properly.
//
// NOTE: All three handler functions are _required_. Leaving them nil will cause
// tests to panic, and this is by design. All migrations should validate their
// behavior at each of these points in their lifecycle.
type migrationTestCase struct {
	preup, during, postdown func(pg *sql.DB, t *testing.T)
}

// migrationTests maps from migration numbers to the migrationTestCast that will
// be used to check their correctness.
var migrationTests = map[int]migrationTestCase{
	1: {
		preup:    ensureNotTables("jobs"),
		during:   ensureTables("jobs"),
		postdown: ensureNotTables("jobs"),
	},
	2: {
		preup:    ensureNotTables("testcases"),
		during:   ensureTables("testcases"),
		postdown: ensureNotTables("testcases"),
	},
	3: {
		preup:    ensureNotColumns("jobs", "upgrade_version"),
		during:   ensureColumns("jobs", "upgrade_version"),
		postdown: ensureNotColumns("jobs", "upgrade_version"),
	},
}

// TestMigrations runs all configured migrations up and down, verifying their correctness
// using the contents of `migrationTests`.
func TestMigrations(t *testing.T) {
	// if dbPool is nil, assume database connectivity is unavailable and skip the test
	if dbPool == nil {
		t.Skip()
	}
	// generate probably-unique DB name
	testDatabase := "test_db_" + fmt.Sprintf("%d", time.Now().UnixNano())
	config := getDBConfig(t)
	urlWithoutDB := config.URL()
	// create DB for testing
	if err := db.WithDB(urlWithoutDB, func(pd *sql.DB) error {
		_, err := pd.Exec("CREATE DATABASE " + testDatabase)
		if err != nil {
			return fmt.Errorf("failed creating ephemeral test database: %v", err)
		}
		return nil
	}); err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	// ensure that our test DB is destroyed later
	t.Cleanup(func() {
		dropDb := func(pd *sql.DB) error {
			_, err := pd.Exec("DROP DATABASE " + testDatabase)
			if err != nil {
				return fmt.Errorf("failed dropping ephemeral test database: %v", err)
			}
			return nil
		}
		start := time.Now()
		for err := db.WithDB(urlWithoutDB, dropDb); err != nil; err = db.WithDB(urlWithoutDB, dropDb) {
			if time.Now().Sub(start) > 2*time.Minute {
				t.Fatalf("Failed to drop test database: %v", err)
			}
			time.Sleep(time.Second * 5)
		}
	})
	// exercise our migrations against the test DB
	config.Database = testDatabase
	if err := db.WithDB(config.URL(), func(pg *sql.DB) error {
		return db.WithMigrator(pg, func(m *migrate.Migrate) error {
			// make sure we know how many migrations exist, and that they all apply cleanly
			// up and down
			if err := m.Up(); err != nil {
				t.Fatalf("Failed running all up migrations: %v", err)
			}
			maxVersion, _, err := m.Version()
			if err != nil {
				t.Fatalf("Did not expect error fetching final migration version: %v", err)
			}
			if err := m.Down(); err != nil {
				t.Fatalf("Failed running all down migrations: %v", err)
			}
			// ensure each migration passes its own tests
			for migrationNum := 1; migrationNum <= int(maxVersion); migrationNum++ {
				testcase, ok := migrationTests[migrationNum]
				if !ok {
					t.Fatalf("No test cases provided for migration number %d", migrationNum)
				}
				testcase.preup(pg, t)
				if err := m.Steps(1); err != nil {
					t.Fatalf("Failed running migration %d: %v", migrationNum, err)
				}
				version, _, err := m.Version()
				if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
					t.Fatalf("Did not expect error fetching migration version: %v", err)
				}
				if int(version) != migrationNum {
					t.Fatalf("Expected version after migration to be %d, got %d", migrationNum, version)
				}
				testcase.during(pg, t)
				if err := m.Steps(-1); err != nil {
					t.Fatalf("Failed reversing migration %d: %v", migrationNum, err)
				}
				testcase.postdown(pg, t)
				if err := m.Steps(1); err != nil {
					t.Fatalf("Failed re-applying migration %d: %v", migrationNum, err)
				}
			}
			return nil
		})
	}); err != nil {
		t.Fatalf("Expected to succeed creating db, got %v", err)
	}
}
