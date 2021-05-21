package db

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	migratepgx "github.com/golang-migrate/migrate/v4/database/pgx"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:generate go run github.com/kyleconroy/sqlc/cmd/sqlc generate

//go:embed migrations
var migrations embed.FS

// WithDB opens a connection to the postgres url specified and runs a function with access to
// the resulting database connection. Any error returned during the opening or closing of the
// database connection will be returned, as will any error from `f`. Errors from `f` supercede
// errors in closing the connection.
func WithDB(url string, f func(*sql.DB) error) (err error) {
	db, err := sql.Open("pgx", url)
	if err != nil {
		return fmt.Errorf("failed opening db: %w", err)
	}
	defer func() {
		if e := db.Close(); err == nil && e != nil {
			err = e
		}
	}()
	return f(db)
}

// WithMigrator uses the provided database, creates a Migrate type, and runs a function
// with that migrate type available. Any error during the connection to the database, creation
// of the Migrate, or returned by `f` will be returned.
func WithMigrator(db *sql.DB, f func(*migrate.Migrate) error) (err error) {
	sd, err := iofs.New(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("failed creating migration source from embedded migrations: %w", err)
	}
	driver, err := migratepgx.WithInstance(db, &migratepgx.Config{
		MigrationsTable:       migratepgx.DefaultMigrationsTable,
		MultiStatementMaxSize: migratepgx.DefaultMultiStatementMaxSize,
	})
	if err != nil {
		return fmt.Errorf("failed wrapping db conn in migration driver: %w", err)
	}
	migrater, err := migrate.NewWithInstance("embedded-migrations", sd, "pgx", driver)
	if err != nil {
		return fmt.Errorf("failed creating migrator: %w", err)
	}
	defer func() {
		if se := sd.Close(); se != nil && err == nil {
			err = se
		}
	}()
	return f(migrater)
}
