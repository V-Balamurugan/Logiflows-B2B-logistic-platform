package httpapi

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/redis/go-redis/v9"
	"logiflows/backend/internal/auth"
	"logiflows/backend/internal/config"
	"logiflows/backend/internal/middleware"
	"logiflows/backend/internal/swagger"
)

type ServerDeps struct {
	Config      *config.Config
	Logger      *slog.Logger
	DB          *sql.DB
	Redis       *redis.Client
	AuthService *auth.AuthService
	Version     string
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
	authHandler := NewAuthHandler(deps.AuthService)

	// Interactive Swagger UI & OpenAPI Specification routes
	r.Get("/swagger", swagger.UIHandler)
	r.Get("/swagger/*", swagger.UIHandler)
	r.Get("/docs", swagger.UIHandler)
	r.Get("/openapi.yaml", swagger.SpecHandler)

	// Base root route
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		RespondJSON(w, r, http.StatusOK, map[string]string{
			"service":  "LogiFlows Authoritative Backend",
			"status":   "operational",
			"version":  deps.Version,
			"swagger":  "/swagger",
			"docs":     "/docs",
			"api_docs": "/api/v1/openapi.yaml",
		}, nil)
	})

	// API v1 Namespace
	r.Route("/api/v1", func(v1 chi.Router) {
		// OpenAPI Spec & Documentation inside /api/v1
		v1.Get("/openapi.yaml", swagger.SpecHandler)
		v1.Get("/docs", swagger.UIHandler)
		v1.Get("/swagger", swagger.UIHandler)
		// System Health
		v1.Get("/healthz", healthHandler.Healthz)
		v1.Get("/readyz", healthHandler.Readyz)

		// Foundation API Ping
		v1.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			RespondJSON(w, r, http.StatusOK, map[string]string{
				"message": "pong",
				"system":  "LogiFlows Core Logistics Platform",
			}, nil)
		})

		// Public Authentication
		v1.Route("/auth", func(authRouter chi.Router) {
			authRouter.Post("/register", authHandler.Register)
			authRouter.Post("/login", authHandler.Login)
			authRouter.Post("/refresh", authHandler.Refresh)

			// Authenticated Auth Endpoints
			authRouter.Group(func(protected chi.Router) {
				protected.Use(middleware.Authenticate(deps.Config))
				protected.Post("/logout", authHandler.Logout)
				protected.Get("/me", authHandler.Me)
			})
		})

		// RBAC and Multi-Tenant Protected Test Verification Endpoints
		v1.Group(func(protected chi.Router) {
			protected.Use(middleware.Authenticate(deps.Config))

			// Endpoint requiring platform admin role
			protected.With(middleware.RequireRole(auth.RolePlatformAdmin)).Get("/admin/system-check", func(w http.ResponseWriter, r *http.Request) {
				RespondJSON(w, r, http.StatusOK, map[string]string{
					"message": "Platform admin access granted",
				}, nil)
			})

			// Endpoint requiring parcel.create permission
			protected.With(middleware.RequirePermission(auth.PermParcelCreate)).Get("/parcels/permission-check", func(w http.ResponseWriter, r *http.Request) {
				RespondJSON(w, r, http.StatusOK, map[string]string{
					"message": "User has parcel.create permission",
				}, nil)
			})

			// Endpoint verifying tenant isolation boundary
			protected.Route("/tenants/{tenant_id}", func(tenantRouter chi.Router) {
				tenantRouter.Use(middleware.EnforceTenantIsolation("tenant_id"))
				tenantRouter.Get("/boundary-check", func(w http.ResponseWriter, r *http.Request) {
					routeTenantID := chi.URLParam(r, "tenant_id")
					RespondJSON(w, r, http.StatusOK, map[string]string{
						"message":   "Tenant boundary access verified",
						"tenant_id": routeTenantID,
					}, nil)
				})
			})
		})
	})

	return r
}
