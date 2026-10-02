package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"logiflows/backend/internal/auth"
	"logiflows/backend/internal/config"
	"logiflows/backend/internal/database"
	"logiflows/backend/internal/httpapi"
	"logiflows/backend/internal/logging"
	"logiflows/backend/internal/migrations"
	"logiflows/backend/internal/tenancy"
)

const AppVersion = "0.2.0-tenants-branches"

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error loading configuration: %v\n", err)
		os.Exit(1)
	}

	logLevel := "info"
	if cfg.Debug {
		logLevel = "debug"
	}
	logger := logging.NewLogger(logLevel, cfg.AppEnv == "production")
	slog.SetDefault(logger)

	logger.Info("Starting LogiFlows Authoritative Backend",
		slog.String("version", AppVersion),
		slog.String("env", cfg.AppEnv),
		slog.Int("port", cfg.AppPort),
	)

	// Initialize Database Connection
	db, err := database.NewPostgresDB(cfg, logger)
	if err != nil {
		logger.Error("Failed to initialize database", slog.String("error", err.Error()))
	} else if db != nil {
		// Run initial and auth migrations
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		if err := migrations.RunMigrations(ctx, db, logger); err != nil {
			logger.Warn("Auto-migration skipped or encountered error; verify schema", slog.String("error", err.Error()))
		}
		cancel()
	}

	// Initialize Redis Client
	redisClient := database.NewRedisClient(cfg, logger)

	// Initialize Core Auth Service
	authService := auth.NewAuthService(db, cfg, logger)

	// Initialize Tenancy and Geospatial Serviceability Service
	tenancyRepo := tenancy.NewRepository(db)
	tenancyService := tenancy.NewService(tenancyRepo, logger)

	// Build Router
	router := httpapi.BuildRouter(httpapi.ServerDeps{
		Config:         cfg,
		Logger:         logger,
		DB:             db,
		Redis:          redisClient,
		AuthService:    authService,
		TenancyService: tenancyService,
		Version:        AppVersion,
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.AppPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info(fmt.Sprintf("Server listening on port %d", cfg.AppPort))
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			logger.Error("Server encountered fatal error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	case sig := <-shutdown:
		logger.Info("Shutdown signal received", slog.String("signal", sig.String()))

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Graceful shutdown failed, forcing close", slog.String("error", err.Error()))
			_ = server.Close()
		}

		if db != nil {
			_ = db.Close()
		}
		if redisClient != nil {
			_ = redisClient.Close()
		}

		logger.Info("LogiFlows backend exited cleanly")
	}
}
