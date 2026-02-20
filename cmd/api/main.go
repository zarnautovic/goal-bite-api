package main

import (
	"log/slog"
	"os"

	"nutrition/internal/app"
	"nutrition/internal/config"

	_ "nutrition/docs/swagger"
)

// @title Nutrition API
// @version 1.0
// @description Nutrition tracking API.
// @description
// @description Error code catalog:
// @description - See docs/openapi-error-codes.md in repository root.
// @BasePath /api/v1
// @schemes http https
func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	application, err := app.New(cfg)
	if err != nil {
		slog.Error("failed to initialize app", "error", err)
		os.Exit(1)
	}

	if err := application.Run(); err != nil {
		slog.Error("application stopped with error", "error", err)
		os.Exit(1)
	}
}
