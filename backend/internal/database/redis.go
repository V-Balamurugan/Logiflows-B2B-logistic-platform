package database

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"logiflows/backend/internal/config"
)

func NewRedisClient(cfg *config.Config, logger *slog.Logger) *redis.Client {
	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		logger.Warn("Failed to parse Redis URL; using fallback default options",
			slog.String("redis_url", cfg.RedisURL),
			slog.String("error", err.Error()),
		)
		opts = &redis.Options{
			Addr: "localhost:6379",
		}
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logger.Warn("Redis ping failed during startup; caching and realtime features will degrade gracefully",
			slog.String("error", err.Error()),
		)
	} else {
		logger.Info("Connected to Redis successfully")
	}

	return client
}
