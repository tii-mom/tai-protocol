package main

import (
	"log"

	"github.com/tii-mom/tai-protocol/backend/internal/config"
	"github.com/tii-mom/tai-protocol/backend/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	srv := server.New(cfg)
	if err := srv.Run(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
