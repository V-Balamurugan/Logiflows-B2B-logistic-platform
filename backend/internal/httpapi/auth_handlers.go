package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"logiflows/backend/internal/auth"
	"logiflows/backend/internal/middleware"
)

type AuthHandler struct {
	authService *auth.AuthService
}

func NewAuthHandler(authService *auth.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req auth.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Failed to parse request JSON", nil)
		return
	}

	ip := r.RemoteAddr
	userAgent := r.UserAgent()

	user, tokens, err := h.authService.Register(r.Context(), req, ip, userAgent)
	if err != nil {
		if errors.Is(err, auth.ErrEmailExists) {
			RespondError(w, r, http.StatusConflict, "EMAIL_EXISTS", err.Error(), nil)
			return
		}
		if errors.Is(err, auth.ErrUnauthorizedRole) || errors.Is(err, auth.ErrWeakPassword) {
			RespondError(w, r, http.StatusBadRequest, "INVALID_REGISTRATION", err.Error(), nil)
			return
		}
		RespondError(w, r, http.StatusInternalServerError, "REGISTRATION_FAILED", err.Error(), nil)
		return
	}

	permissions := auth.RolePermissions[user.Role]

	RespondJSON(w, r, http.StatusCreated, map[string]interface{}{
		"user":        user,
		"tokens":      tokens,
		"permissions": permissions,
	}, nil)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Failed to parse request JSON", nil)
		return
	}

	ip := r.RemoteAddr
	userAgent := r.UserAgent()

	user, tokens, err := h.authService.Login(r.Context(), req.Email, req.Password, ip, userAgent)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			RespondError(w, r, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password", nil)
			return
		}
		if errors.Is(err, auth.ErrUserInactive) {
			RespondError(w, r, http.StatusForbidden, "ACCOUNT_INACTIVE", "Your account has been deactivated", nil)
			return
		}
		RespondError(w, r, http.StatusInternalServerError, "LOGIN_FAILED", "Authentication encountered an error", nil)
		return
	}

	permissions := auth.RolePermissions[user.Role]

	RespondJSON(w, r, http.StatusOK, map[string]interface{}{
		"user":        user,
		"tokens":      tokens,
		"permissions": permissions,
	}, nil)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		RespondError(w, r, http.StatusBadRequest, "INVALID_REQUEST_BODY", "refresh_token is required", nil)
		return
	}

	ip := r.RemoteAddr
	userAgent := r.UserAgent()

	user, tokens, err := h.authService.RefreshToken(r.Context(), req.RefreshToken, ip, userAgent)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidToken) || errors.Is(err, auth.ErrTokenRevoked) {
			RespondError(w, r, http.StatusUnauthorized, "INVALID_TOKEN", err.Error(), nil)
			return
		}
		RespondError(w, r, http.StatusInternalServerError, "TOKEN_REFRESH_FAILED", err.Error(), nil)
		return
	}

	permissions := auth.RolePermissions[user.Role]

	RespondJSON(w, r, http.StatusOK, map[string]interface{}{
		"user":        user,
		"tokens":      tokens,
		"permissions": permissions,
	}, nil)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	userID := middleware.GetUserID(r.Context())
	_ = h.authService.Logout(r.Context(), req.RefreshToken, userID)

	RespondJSON(w, r, http.StatusOK, map[string]string{
		"message": "logged out successfully",
	}, nil)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		RespondError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required", nil)
		return
	}

	user, err := h.authService.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		RespondError(w, r, http.StatusNotFound, "USER_NOT_FOUND", "User profile not found", nil)
		return
	}

	permissions := auth.RolePermissions[user.Role]

	RespondJSON(w, r, http.StatusOK, map[string]interface{}{
		"user":        user,
		"permissions": permissions,
	}, nil)
}
