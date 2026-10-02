package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/lib/pq"
	"logiflows/backend/internal/config"
)

func NewPostgresDB(cfg *config.Config, logger *slog.Logger) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(cfg.DBConnMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		logger.Warn("Postgres connection ping failed during startup; continuing with degraded state",
			slog.String("error", err.Error()),
		)
		return db, nil // Return instance so health/ready checks can report status
	}

	logger.Info("Connected to PostgreSQL successfully")
	return db, nil
}
