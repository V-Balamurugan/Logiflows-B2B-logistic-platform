package middleware

import (
	"net/http"

	"logiflows/backend/internal/auth"
)

func RequireRole(roles ...auth.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetUserClaims(r.Context())
			if claims == nil {
				respondJSONError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required", nil)
				return
			}

			allowed := false
			for _, role := range roles {
				if claims.Role == role {
					allowed = true
					break
				}
			}

			if !allowed {
				respondJSONError(w, r, http.StatusForbidden, "FORBIDDEN_ROLE", "You do not have the required role to access this resource", map[string]interface{}{
					"required_roles": roles,
					"user_role":      claims.Role,
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequirePermission(perm auth.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetUserClaims(r.Context())
			if claims == nil {
				respondJSONError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required", nil)
				return
			}

			if !auth.HasPermission(claims.Role, perm) {
				respondJSONError(w, r, http.StatusForbidden, "PERMISSION_DENIED", "You lack the authoritative permission required for this action", map[string]interface{}{
					"required_permission": perm,
					"user_role":           claims.Role,
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
