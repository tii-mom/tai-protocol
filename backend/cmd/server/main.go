package main

import (
	"context"
	"log"

	"github.com/tii-mom/tai-protocol/backend/ent"
	"github.com/tii-mom/tai-protocol/backend/internal/config"
	"github.com/tii-mom/tai-protocol/backend/internal/server"

	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Connect to PostgreSQL
	db, err := ent.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Auto-migrate schema (dev only; production uses explicit migrations)
	if err := db.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed to run schema migration: %v", err)
	}

	log.Println("Database connected and schema synced")

	srv := server.New(cfg, db)
	if err := srv.Run(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
