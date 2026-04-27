package main

import (
	"finance_tracker/internal/config"
	repository "finance_tracker/internal/repository/db"
	"log"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("DB Config: host=%s port=%s user=%s password=%s dbname=%s sslmode=%s\n",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)

	db, err := repository.ConnectDB(*cfg)

	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}
	defer db.Close()

	// 3. Запускаем миграции
	if err := repository.RunMigrations(*cfg, db); err != nil {
		log.Fatalf("Migration error: %v", err)
	}

	log.Println("Application started successfully")

	log.Printf("Database: %s", cfg)
	log.Printf("Server will listen on port: %s", cfg.ServerPort)

}
