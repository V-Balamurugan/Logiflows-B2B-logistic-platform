package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"logiflows/backend/internal/auth"
	"logiflows/backend/internal/config"
	"logiflows/backend/internal/middleware"
)

func createTestToken(t *testing.T, userID uuid.UUID, role auth.Role, tenantID *uuid.UUID, secret string, expired bool) string {
	now := time.Now().UTC()
	expiry := now.Add(15 * time.Minute)
	if expired {
		expiry = now.Add(-15 * time.Minute)
	}

	claims := auth.UserClaims{
		UserID:   userID,
		Email:    "test@logiflows.io",
		Role:     role,
		TenantID: tenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "logiflows-auth-service",
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign test token: %v", err)
	}
	return tokenStr
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test_secret_key_at_least_32_bytes_long"}
	userID := uuid.New()
	tenantID := uuid.New()
	token := createTestToken(t, userID, auth.RoleCustomer, &tenantID, cfg.JWTSecret, false)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	var capturedClaims *auth.UserClaims
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedClaims = middleware.GetUserClaims(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.Authenticate(cfg)(testHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	if capturedClaims == nil || capturedClaims.UserID != userID {
		t.Fatalf("expected claims to be populated with user ID %s", userID)
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test_secret_key_at_least_32_bytes_long"}
	token := createTestToken(t, uuid.New(), auth.RoleCustomer, nil, cfg.JWTSecret, true)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.Authenticate(cfg)(testHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for expired token, got %d", rr.Code)
	}
}

func TestRBACMiddleware_RequireRole(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test_secret_key_at_least_32_bytes_long"}

	// Customer token
	customerToken := createTestToken(t, uuid.New(), auth.RoleCustomer, nil, cfg.JWTSecret, false)
	// Platform Admin token
	adminToken := createTestToken(t, uuid.New(), auth.RolePlatformAdmin, nil, cfg.JWTSecret, false)

	protectedHandler := middleware.Authenticate(cfg)(
		middleware.RequireRole(auth.RolePlatformAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})),
	)

	// Test 1: Customer must be rejected with 403 Forbidden
	req1 := httptest.NewRequest("GET", "/admin/only", nil)
	req1.Header.Set("Authorization", "Bearer "+customerToken)
	rr1 := httptest.NewRecorder()
	protectedHandler.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for Customer accessing admin route, got %d", rr1.Code)
	}

	// Test 2: Platform Admin must succeed with 200 OK
	req2 := httptest.NewRequest("GET", "/admin/only", nil)
	req2.Header.Set("Authorization", "Bearer "+adminToken)
	rr2 := httptest.NewRecorder()
	protectedHandler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200 for Platform Admin accessing admin route, got %d", rr2.Code)
	}
}

func TestRBACMiddleware_RequirePermission(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test_secret_key_at_least_32_bytes_long"}

	// Customer has PermParcelCreate but lacks PermTenantCreate
	customerToken := createTestToken(t, uuid.New(), auth.RoleCustomer, nil, cfg.JWTSecret, false)

	createParcelHandler := middleware.Authenticate(cfg)(
		middleware.RequirePermission(auth.PermParcelCreate)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})),
	)

	createTenantHandler := middleware.Authenticate(cfg)(
		middleware.RequirePermission(auth.PermTenantCreate)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})),
	)

	// Customer allowed to create parcel
	req1 := httptest.NewRequest("POST", "/parcels", nil)
	req1.Header.Set("Authorization", "Bearer "+customerToken)
	rr1 := httptest.NewRecorder()
	createParcelHandler.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Fatalf("expected 200 for customer creating parcel, got %d", rr1.Code)
	}

	// Customer denied creating tenant
	req2 := httptest.NewRequest("POST", "/tenants", nil)
	req2.Header.Set("Authorization", "Bearer "+customerToken)
	rr2 := httptest.NewRecorder()
	createTenantHandler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for customer creating tenant, got %d", rr2.Code)
	}
}

func TestTenantIsolationMiddleware(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test_secret_key_at_least_32_bytes_long"}

	tenantA := uuid.New()
	tenantB := uuid.New()

	tokenTenantA := createTestToken(t, uuid.New(), auth.RoleTenantAdmin, &tenantA, cfg.JWTSecret, false)
	tokenPlatformAdmin := createTestToken(t, uuid.New(), auth.RolePlatformAdmin, nil, cfg.JWTSecret, false)

	r := chi.NewRouter()
	r.Use(middleware.Authenticate(cfg))
	r.Route("/tenants/{tenant_id}", func(tr chi.Router) {
		tr.Use(middleware.EnforceTenantIsolation("tenant_id"))
		tr.Get("/data", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	// Case 1: Tenant A accesses Tenant A resource -> ALLOWED (200)
	req1 := httptest.NewRequest("GET", "/tenants/"+tenantA.String()+"/data", nil)
	req1.Header.Set("Authorization", "Bearer "+tokenTenantA)
	rr1 := httptest.NewRecorder()
	r.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Fatalf("expected 200 for accessing own tenant, got %d", rr1.Code)
	}

	// Case 2: Tenant A attempts to access Tenant B resource -> FORBIDDEN (403)
	req2 := httptest.NewRequest("GET", "/tenants/"+tenantB.String()+"/data", nil)
	req2.Header.Set("Authorization", "Bearer "+tokenTenantA)
	rr2 := httptest.NewRecorder()
	r.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for cross-tenant access, got %d", rr2.Code)
	}

	var errResp ErrorEnvelope
	_ = json.Unmarshal(rr2.Body.Bytes(), &errResp)
	if errResp.Error.Code != "CROSS_TENANT_ACCESS_DENIED" {
		t.Errorf("expected code CROSS_TENANT_ACCESS_DENIED, got %s", errResp.Error.Code)
	}

	// Case 3: Platform Admin accesses Tenant B resource -> ALLOWED (200)
	req3 := httptest.NewRequest("GET", "/tenants/"+tenantB.String()+"/data", nil)
	req3.Header.Set("Authorization", "Bearer "+tokenPlatformAdmin)
	rr3 := httptest.NewRecorder()
	r.ServeHTTP(rr3, req3)
	if rr3.Code != http.StatusOK {
		t.Fatalf("expected 200 for Platform Admin accessing tenant resource, got %d", rr3.Code)
	}
}
