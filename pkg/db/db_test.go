package db_test

import (
	"database/sql"
	"fmt"
	"log"
	"testing"

	"github.com/openshift/osde2e/pkg/db"
	"github.com/ory/dockertest"
)

var dbPool *dockertest.Pool = func() *dockertest.Pool {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	dbPool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
	return dbPool
}()

func getDBURL(t *testing.T) (url string, cleanup func()) {
	const password = "secret"
	// pulls an image, creates a container based on it and runs it
	resource, err := dbPool.Run("postgres", "13", []string{"POSTGRES_PASSWORD=" + password})
	if err != nil {
		t.Fatalf("Could not start resource: %s", err)
	}
	cleanup = func() {
		if err := dbPool.Purge(resource); err != nil {
			log.Printf("Could not purge resource: %s", err)
		}
	}

	url = fmt.Sprintf("postgres://postgres:%s@127.0.0.1:%s/postgres?sslmode=disable", password, resource.GetPort("5432/tcp"))

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := dbPool.Retry(func() error {
		db, err := sql.Open("postgres", url)
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		cleanup()
		t.Fatalf("Could not connect to postgres: %s", err)
	}
	return url, cleanup
}

func TestNew(t *testing.T) {
	url, cleanup := getDBURL(t)
	t.Cleanup(cleanup)
	t.Log(url)
	d, err := db.New(url)
	if err != nil {
		t.Fatalf("expected to succeed creating db, got %v", err)
	}
	if d == nil {
		t.Fatalf("expected non-nil db")
	}
}
