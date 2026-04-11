package cli

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"booker/config"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/urfave/cli/v2"
)

func MigrateUp(c *cli.Context) error {
	configPath := c.String("config")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	db, err := sql.Open("pgx", cfg.Database.URI)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	migrationsDir := filepath.Join("migrations")
	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	fmt.Println("Migrations completed successfully")
	return nil
}

func MigrateDown(c *cli.Context) error {
	configPath := c.String("config")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	db, err := sql.Open("pgx", cfg.Database.URI)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	migrationsDir := filepath.Join("migrations")
	if err := goose.Down(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	fmt.Println("Migration rollback completed successfully")
	return nil
}

func MigrateStatus(c *cli.Context) error {
	configPath := c.String("config")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	db, err := sql.Open("pgx", cfg.Database.URI)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	migrationsDir := filepath.Join("migrations")
	if err := goose.Status(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	return nil
}
