package db

import (
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations
var migrations embed.FS

type DB struct {
}

func New(url string) (db *DB, err error) {
	sd, err := iofs.New(migrations, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed creating migration source from embedded migrations: %w", err)
	}
	migrater, err := migrate.NewWithSourceInstance("embedded-migrations", sd, url)
	if err != nil {
		return nil, fmt.Errorf("failed creating migrator: %w", err)
	}
	defer func() {
		if se := sd.Close(); se != nil && err == nil {
			err = se
		}
		if se, de := migrater.Close(); se != nil && err == nil {
			err = se
		} else if de != nil && err == nil {
			err = de
		}
	}()
	if err := migrater.Up(); err != nil {
		return nil, fmt.Errorf("failed running DB migrations: %w", err)
	}
	return nil, nil
}
