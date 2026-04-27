package main

import (
	"finance_tracker/internal/config"
	"log"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Database path: %s", cfg.DBPath)
	log.Printf("Server will listen on port: %s", cfg.ServerPort)

}
