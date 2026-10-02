package httpapi

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/redis/go-redis/v9"
	"logiflows/backend/internal/config"
	"logiflows/backend/internal/middleware"
)

type ServerDeps struct {
	Config  *config.Config
	Logger  *slog.Logger
	DB      *sql.DB
	Redis   *redis.Client
	Version string
}

func BuildRouter(deps ServerDeps) http.Handler {
	r := chi.NewRouter()

	// Global Core Middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RequestLogger(deps.Logger))
	r.Use(middleware.Recoverer(deps.Logger))
	r.Use(chimw.RealIP)

	// Cross-Origin Resource Sharing
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   deps.Config.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID", "X-Tenant-ID"},
		ExposedHeaders:   []string{"Link", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	healthHandler := NewHealthHandler(deps.DB, deps.Redis, deps.Config.AIServiceURL, deps.Version)

	// Base root route
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		RespondJSON(w, r, http.StatusOK, map[string]string{
			"service": "LogiFlows Authoritative Backend",
			"status":  "operational",
			"api_docs": "/api/v1/openapi.yaml",
		}, nil)
	})

	// API v1 Namespace
	r.Route("/api/v1", func(v1 chi.Router) {
		v1.Get("/healthz", healthHandler.Healthz)
		v1.Get("/readyz", healthHandler.Readyz)

		// Foundation API Ping
		v1.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			RespondJSON(w, r, http.StatusOK, map[string]string{
				"message": "pong",
				"system":  "LogiFlows Core Logistics Platform",
			}, nil)
		})
	})

	return r
}
