package db

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"runtime"

	"finance_tracker/internal/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file" // включает поддержку file://

	_ "github.com/jackc/pgx/v5/stdlib"
)

func ConnectDB(cfg config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("cannot open db: %w", err)
	}
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("cannot ping db: %w", err)
	}
	return db, nil
}

func RunMigrations(cfg config.Config, db *sql.DB) error {
	absPath, err := filepath.Abs(cfg.MigrationsPath)
	if err != nil {
		return fmt.Errorf("abs migrations path: %w", err)
	}
	// golang-migrate expects file URLs with forward slashes, including on Windows.
	absPath = filepath.ToSlash(absPath)
	migrationURL := "file:///" + absPath
	// On Windows, file:///D:/... may be interpreted as /D:/... by url parsing.
	// golang-migrate's file source expects file://D:/... (two slashes) for drive paths.
	if runtime.GOOS == "windows" && filepath.VolumeName(absPath) != "" {
		migrationURL = "file://" + absPath
	}

	// Создаём драйвер БД
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("postgres driver: %w", err)
	}

	// Создаём экземпляр мигратора
	m, err := migrate.NewWithDatabaseInstance(migrationURL, "postgres", driver)
	if err != nil {
		return fmt.Errorf("migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration up: %w", err)
	}
	return nil
}
