package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"logiflows/backend/internal/auth"
	"logiflows/backend/internal/config"
	"logiflows/backend/internal/logging"
)

type userContextKey string

const (
	UserClaimsKey userContextKey = "user_claims"
)

func Authenticate(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondJSONError(w, r, http.StatusUnauthorized, "AUTH_HEADER_MISSING", "Authorization header is required", nil)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				respondJSONError(w, r, http.StatusUnauthorized, "AUTH_HEADER_INVALID", "Authorization header format must be Bearer <token>", nil)
				return
			}

			tokenStr := parts[1]
			claims := &auth.UserClaims{}

			token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method")
				}
				return []byte(cfg.JWTSecret), nil
			})

			if err != nil || !token.Valid {
				respondJSONError(w, r, http.StatusUnauthorized, "TOKEN_INVALID", "Access token is invalid or expired", nil)
				return
			}

			ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
			ctx = context.WithValue(ctx, logging.UserIDKey, claims.UserID.String())
			if claims.TenantID != nil {
				ctx = context.WithValue(ctx, logging.TenantIDKey, claims.TenantID.String())
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserClaims(ctx context.Context) *auth.UserClaims {
	if ctx == nil {
		return nil
	}
	if claims, ok := ctx.Value(UserClaimsKey).(*auth.UserClaims); ok {
		return claims
	}
	return nil
}

func GetUserID(ctx context.Context) uuid.UUID {
	claims := GetUserClaims(ctx)
	if claims != nil {
		return claims.UserID
	}
	return uuid.Nil
}

func GetUserRole(ctx context.Context) auth.Role {
	claims := GetUserClaims(ctx)
	if claims != nil {
		return claims.Role
	}
	return ""
}

func GetUserTenantID(ctx context.Context) *uuid.UUID {
	claims := GetUserClaims(ctx)
	if claims != nil {
		return claims.TenantID
	}
	return nil
}
