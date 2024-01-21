package main

import (
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"log"
)

func startDatabaseMigrations(user, password, host, port string) error {
	migrationsPath := "file:///Users/zufar/GolandProjects/Yulia-Lingo/migrations"
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/?sslmode=disable", user, password, host, port)

	migrations, err := migrate.New(migrationsPath, dbURL)
	if err != nil {
		return fmt.Errorf("failed to create migration: %w", err)
	}

	err = migrations.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Migrations applied successfully")
	return nil
}
