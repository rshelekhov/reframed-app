package main

//
// A small CLI utility for running database migrations
//

import (
	"errors"
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rshelekhov/remedi/internal/config"
)

func main() {
	var migrationsPath string

	cfg := config.MustLoad()

	flag.StringVar(&migrationsPath, "migrations-path", "./migrations", "path to migrations")
	flag.Parse()

	if migrationsPath == "" {
		// I'm fine with panic for now, as it's an auxiliary utility.
		panic("migrations-path is required")
	}

	// Create a migrator object by passing the credentials to our database
	m, err := migrate.New(
		"file://"+migrationsPath,
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			cfg.Postgres.User,
			cfg.Postgres.Password,
			cfg.Postgres.Host,
			cfg.Postgres.Port,
			cfg.Postgres.DBName,
			cfg.Postgres.SSLMode),
	)
	if err != nil {
		panic(err)
	}

	// Migrate to the latest version
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")

			return
		}

		panic(err)
	}
}
