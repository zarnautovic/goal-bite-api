package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"goal-bite-api/internal/config"
	"goal-bite-api/internal/db"
)

func main() {
	direction := flag.String("direction", "up", "migration direction: up or down")
	steps := flag.Int("steps", 1, "number of steps for down migrations; ignored for up")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	switch *direction {
	case "up":
		err = db.MigrateUp(cfg.DatabaseURL, cfg.MigrationsPath)
	case "down":
		err = db.MigrateDown(cfg.DatabaseURL, cfg.MigrationsPath, *steps)
	default:
		err = fmt.Errorf("invalid direction %q, expected up or down", *direction)
	}

	if err != nil {
		slog.Error("migration command failed", "error", err)
		os.Exit(1)
	}

	slog.Info("migration command completed", "direction", *direction)
}
