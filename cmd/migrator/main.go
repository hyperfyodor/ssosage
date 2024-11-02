package main

import (
	"errors"
	"flag"
	"fmt"
	"ssosage/internal/config/migrator"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {

	var configPath string
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.Parse()

	cfg := migrator.MustLoad(configPath)

	if cfg.StoragePath == "" {
		panic("storage path is required")
	}

	if cfg.MigrationsPath == "" {
		panic("migrations path is required")
	}

	m, err := migrate.New(
		"file://"+cfg.MigrationsPath,
		fmt.Sprintf("sqlite3://%s?x-migrations-table=%s", cfg.StoragePath, cfg.MigrationsTable),
	)
	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")

			return
		}

		panic(err)
	}

}
