package config

import (
	"errors"
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// internal/config/config.go
type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	JWTSecret      string
	ServerPort     string
	MigrationsPath string
}

func LoadConfig() (*Config, error) {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	var cfg Config

	flag.StringVar(&cfg.DBHost, "db-host", "", "Database host")
	flag.StringVar(&cfg.DBPort, "db-port", "", "Database port")
	flag.StringVar(&cfg.DBUser, "db-user", "", "Database user")
	flag.StringVar(&cfg.DBPassword, "db-password", "", "Database password")
	flag.StringVar(&cfg.DBName, "db-name", "", "Database name")
	flag.StringVar(&cfg.DBSSLMode, "db-sslmode", "", "Database SSL mode")
	flag.StringVar(&cfg.JWTSecret, "jwt-secret", "", "Secret key for JWT signing")
	flag.StringVar(&cfg.ServerPort, "port", "", "HTTP server port")
	flag.StringVar(&cfg.MigrationsPath, "migrations-path", "", "Path to migrations directory")
	flag.Parse()

	if cfg.DBHost == "" {
		cfg.DBHost = os.Getenv("DB_HOST")
	}
	if cfg.DBPort == "" {
		cfg.DBPort = os.Getenv("DB_PORT")
	}
	if cfg.DBUser == "" {
		cfg.DBUser = os.Getenv("DB_USER")
	}
	if cfg.DBPassword == "" {
		cfg.DBPassword = os.Getenv("DB_PASSWORD")
	}
	if cfg.DBName == "" {
		cfg.DBName = os.Getenv("DB_NAME")
	}
	if cfg.DBSSLMode == "" {
		cfg.DBSSLMode = os.Getenv("DB_SSLMODE")
	}
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = os.Getenv("JWT_SECRET")
	}
	if cfg.ServerPort == "" {
		cfg.ServerPort = os.Getenv("SERVER_PORT")
	}
	if cfg.MigrationsPath == "" {
		cfg.MigrationsPath = os.Getenv("MIGRATIONS_PATH")
	}

	if cfg.DBHost == "" {
		cfg.DBHost = "localhost"
	}
	if cfg.DBPort == "" {
		cfg.DBPort = "5442"
	}
	if cfg.DBSSLMode == "" {
		cfg.DBSSLMode = "disable"
	}
	if cfg.DBName == "" {
		cfg.DBName = "finance_db"
	}
	if cfg.DBUser == "" {
		cfg.DBUser = "postgres"
	}
	if cfg.ServerPort == "" {
		cfg.ServerPort = "8081"
	}
	if cfg.MigrationsPath == "" {
		// значение по умолчанию относительно корня проекта
		cfg.MigrationsPath = "internal/repository/migrations"
	}

	if cfg.DBPassword == "" {
		return nil, errors.New("database password is required: set via -db-password flag or DB_PASSWORD env variable")
	}
	if cfg.JWTSecret == "" {
		return nil, errors.New("JWT secret is required: set via -jwt-secret flag or JWT_SECRET env variable")
	}

	return &cfg, nil
}
