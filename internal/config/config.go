package config

import (
	"errors"
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBPath     string
	JWTSecret  string
	ServerPort string
}

func LoadConfig() (*Config, error) {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	var cfg Config

	flag.StringVar(&cfg.DBPath, "db", "", "Path to PostgreSQL database file")
	flag.StringVar(&cfg.JWTSecret, "jwt-secret", "", "Secret key for JWT signing")
	flag.StringVar(&cfg.ServerPort, "port", "", "HTTP server port")
	flag.Parse()

	if cfg.DBPath == "" {
		cfg.DBPath = os.Getenv("DB_PATH")
	}
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = os.Getenv("JWT_SECRET")
	}
	if cfg.ServerPort == "" {
		cfg.ServerPort = os.Getenv("SERVER_PORT")
	}

	if cfg.DBPath == "" {
		cfg.DBPath = "./finance.db"
	}
	if cfg.ServerPort == "" {
		cfg.ServerPort = "8080"
	}
	if cfg.JWTSecret == "" {
		return nil, errors.New("JWT secret is required: set via -jwt-secret flag or JWT_SECRET env variable")
	}

	return &Config{}, nil
}
