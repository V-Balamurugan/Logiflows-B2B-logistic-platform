package middleware

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"logiflows/backend/internal/auth"
)

// EnforceTenantIsolation ensures that if a route is scoped to a {tenant_id},
// non-Platform-Admin users cannot access data belonging to another tenant.
func EnforceTenantIsolation(paramKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetUserClaims(r.Context())
			if claims == nil {
				respondJSONError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required", nil)
				return
			}

			// Platform Admins have global cross-tenant visibility
			if claims.Role == auth.RolePlatformAdmin {
				next.ServeHTTP(w, r)
				return
			}

			routeTenantIDStr := chi.URLParam(r, paramKey)
			if routeTenantIDStr == "" {
				next.ServeHTTP(w, r)
				return
			}

			routeTenantID, err := uuid.Parse(routeTenantIDStr)
			if err != nil {
				respondJSONError(w, r, http.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant ID format", nil)
				return
			}

			if claims.TenantID == nil || *claims.TenantID != routeTenantID {
				respondJSONError(w, r, http.StatusForbidden, "CROSS_TENANT_ACCESS_DENIED", "You are not authorized to access resources belonging to another tenant", map[string]interface{}{
					"attempted_tenant": routeTenantID.String(),
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
