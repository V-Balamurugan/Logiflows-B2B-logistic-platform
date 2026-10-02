package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"logiflows/backend/internal/config"
	"logiflows/backend/internal/database"
	"logiflows/backend/internal/logging"
	"logiflows/backend/internal/migrations"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		os.Exit(1)
	}

	logger := logging.NewLogger("info", false)
	logger.Info("Executing database migrations CLI tool")

	db, err := database.NewPostgresDB(cfg, logger)
	if err != nil || db == nil {
		logger.Error("Database connection failed", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := migrations.RunMigrations(ctx, db, logger); err != nil {
		logger.Error("Migrations failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("All database migrations applied successfully")
}
