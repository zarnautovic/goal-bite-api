package main

import (
	"log/slog"
	"os"

	"goal-bite-api/internal/config"
	"goal-bite-api/internal/db"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	database, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect database", "error", err)
		os.Exit(1)
	}

	if err := db.Seed(database); err != nil {
		slog.Error("failed to seed data", "error", err)
		os.Exit(1)
	}

	slog.Info("seed completed")
}
