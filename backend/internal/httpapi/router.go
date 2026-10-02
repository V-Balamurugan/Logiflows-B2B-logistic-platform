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
	"logiflows/backend/internal/tenancy"
)

type ServerDeps struct {
	Config         *config.Config
	Logger         *slog.Logger
	DB             *sql.DB
	Redis          *redis.Client
	AuthService    *auth.AuthService
	TenancyService tenancy.Service
	Version        string
}

func BuildRouter(deps ServerDeps) http.Handler {
	if deps.Logger == nil {
		deps.Logger = slog.Default()
	}
	if deps.Config == nil {
		deps.Config = &config.Config{
			JWTSecret: "test_secret_key_at_least_32_bytes_long",
		}
	}

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
	var tenancyHandler *TenancyHandler
	if deps.TenancyService != nil {
		tenancyHandler = NewTenancyHandler(deps.TenancyService)
	}

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

		// Public or Authenticated Geospatial Serviceability Check
		if tenancyHandler != nil {
			v1.Post("/serviceability/check", tenancyHandler.CheckServiceability)
			v1.Get("/serviceability/check", tenancyHandler.CheckServiceability)
		}

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

		// Protected Operations (RBAC & Tenant Isolation)
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

			// Tenants & Branches Route Group
			if tenancyHandler != nil {
				protected.Route("/tenants", func(tr chi.Router) {
					// Tenant listing & creation
					tr.With(middleware.RequireRole(auth.RolePlatformAdmin)).Post("/", tenancyHandler.CreateTenant)
					tr.With(middleware.RequirePermission(auth.PermTenantRead)).Get("/", tenancyHandler.ListTenants)

					// Scoped to specific tenant
					tr.Route("/{tenant_id}", func(singleTenant chi.Router) {
						singleTenant.Use(middleware.EnforceTenantIsolation("tenant_id"))
						singleTenant.With(middleware.RequirePermission(auth.PermTenantRead)).Get("/", tenancyHandler.GetTenant)
						singleTenant.With(middleware.RequirePermission(auth.PermTenantUpdate)).Put("/", tenancyHandler.UpdateTenant)

						// Geospatial Coverage GeoJSON FeatureCollection
						singleTenant.Get("/serviceability/coverage", tenancyHandler.GetCoverage)

						// Branch Management (Strictly Tenant-isolated)
						singleTenant.Route("/branches", func(br chi.Router) {
							br.Get("/", tenancyHandler.ListBranches)
							br.With(middleware.RequirePermission(auth.PermBranchManage)).Post("/", tenancyHandler.CreateBranch)
							br.Get("/{branch_id}", tenancyHandler.GetBranch)
							br.With(middleware.RequirePermission(auth.PermBranchManage)).Put("/{branch_id}", tenancyHandler.UpdateBranch)
							br.With(middleware.RequirePermission(auth.PermBranchManage)).Delete("/{branch_id}", tenancyHandler.DeleteBranch)
						})

						// Boundary verification
						singleTenant.Get("/boundary-check", func(w http.ResponseWriter, r *http.Request) {
							routeTenantID := chi.URLParam(r, "tenant_id")
							RespondJSON(w, r, http.StatusOK, map[string]string{
								"message":   "Tenant boundary access verified",
								"tenant_id": routeTenantID,
							}, nil)
						})
					})
				})
			}
		})
	})

	return r
}
