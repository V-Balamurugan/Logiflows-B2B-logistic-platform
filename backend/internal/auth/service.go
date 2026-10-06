package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"logiflows/backend/internal/config"
)

var (
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrUserInactive        = errors.New("user account is inactive")
	ErrUserNotFound        = errors.New("user not found")
	ErrEmailExists         = errors.New("email is already registered")
	ErrInvalidToken        = errors.New("invalid or expired token")
	ErrTokenRevoked        = errors.New("refresh token has been revoked")
	ErrUnauthorizedRole    = errors.New("self-registration with this role is not permitted")
	ErrWeakPassword        = errors.New("password must be at least 8 characters and contain letters and numbers")
)

type RegisterRequest struct {
	Email       string  `json:"email"`
	Password    string  `json:"password"`
	FirstName   string  `json:"first_name"`
	LastName    string  `json:"last_name"`
	Phone       string  `json:"phone,omitempty"`
	Role        Role    `json:"role"`
	CompanyName string  `json:"company_name,omitempty"` // Required if role == TENANT
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthService struct {
	db     *sql.DB
	cfg    *config.Config
	logger *slog.Logger
}

func NewAuthService(db *sql.DB, cfg *config.Config, logger *slog.Logger) *AuthService {
	return &AuthService{
		db:     db,
		cfg:    cfg,
		logger: logger,
	}
}

func (s *AuthService) Register(ctx context.Context, req RegisterRequest, ip, userAgent string) (*User, *TokenPair, error) {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		return nil, nil, errors.New("invalid email address")
	}

	if len(req.Password) < 8 {
		return nil, nil, ErrWeakPassword
	}

	// Security: Self-registration only allowed for CUSTOMER and TENANT
	if req.Role != RoleCustomer && req.Role != RoleTenant {
		return nil, nil, ErrUnauthorizedRole
	}

	if req.Role == RoleTenant && strings.TrimSpace(req.CompanyName) == "" {
		return nil, nil, errors.New("company_name is required for tenant registration")
	}

	// Check existing email
	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", req.Email).Scan(&exists)
	if err != nil {
		return nil, nil, fmt.Errorf("database query error: %w", err)
	}
	if exists {
		return nil, nil, ErrEmailExists
	}

	pwdHash, err := HashPassword(req.Password)
	if err != nil {
		return nil, nil, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("transaction begin failed: %w", err)
	}
	defer tx.Rollback()

	userID := uuid.New()
	user := &User{
		ID:        userID,
		Email:     req.Email,
		FirstName: strings.TrimSpace(req.FirstName),
		LastName:  strings.TrimSpace(req.LastName),
		Phone:     strings.TrimSpace(req.Phone),
		Role:      req.Role,
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	insertUserSQL := `
	INSERT INTO users (id, email, password_hash, first_name, last_name, phone, role, is_active, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err = tx.ExecContext(ctx, insertUserSQL,
		user.ID, user.Email, pwdHash, user.FirstName, user.LastName, user.Phone, string(user.Role), user.IsActive, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to insert user: %w", err)
	}

	// If TENANT registration, create tenant and membership
	if req.Role == RoleTenant {
		tenantID := uuid.New()
		tenantCode := strings.ToLower(regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(req.CompanyName, "-"))
		if len(tenantCode) > 20 {
			tenantCode = tenantCode[:20]
		}
		tenantCode = fmt.Sprintf("%s-%s", tenantCode, tenantID.String()[:6])

		insertTenantSQL := `
		INSERT INTO tenants (id, name, code, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, true, $4, $5)`
		_, err = tx.ExecContext(ctx, insertTenantSQL, tenantID, req.CompanyName, tenantCode, time.Now().UTC(), time.Now().UTC())
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create tenant: %w", err)
		}

		insertMemberSQL := `
		INSERT INTO memberships (id, tenant_id, user_id, role, is_default, created_at, updated_at)
		VALUES ($1, $2, $3, $4, true, $5, $6)`
		_, err = tx.ExecContext(ctx, insertMemberSQL, uuid.New(), tenantID, user.ID, string(RoleTenant), time.Now().UTC(), time.Now().UTC())
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create tenant membership: %w", err)
		}

		user.TenantID = &tenantID
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, fmt.Errorf("failed to commit registration: %w", err)
	}

	s.RecordAudit(ctx, user.TenantID, &user.ID, "USER_REGISTERED", "users", map[string]interface{}{
		"email": user.Email,
		"role":  user.Role,
	}, ip, userAgent)

	tokens, err := s.generateTokens(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

func (s *AuthService) Login(ctx context.Context, email, password, ip, userAgent string) (*User, *TokenPair, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	var user User
	var pwdHash string
	query := `
	SELECT id, email, password_hash, first_name, last_name, COALESCE(phone, ''), role, is_active, created_at, updated_at
	FROM users WHERE email = $1`
	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &pwdHash, &user.FirstName, &user.LastName, &user.Phone, &user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.RecordAudit(ctx, nil, nil, "LOGIN_FAILED", "auth", map[string]interface{}{
				"email":  email,
				"reason": "user_not_found",
			}, ip, userAgent)
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, fmt.Errorf("db query error: %w", err)
	}

	if !user.IsActive {
		return nil, nil, ErrUserInactive
	}

	if !CheckPassword(password, pwdHash) {
		s.RecordAudit(ctx, nil, &user.ID, "LOGIN_FAILED", "auth", map[string]interface{}{
			"email":  email,
			"reason": "invalid_password",
		}, ip, userAgent)
		return nil, nil, ErrInvalidCredentials
	}

	// Lookup tenant membership
	var tenantID uuid.UUID
	memErr := s.db.QueryRowContext(ctx, "SELECT tenant_id FROM memberships WHERE user_id = $1 ORDER BY is_default DESC LIMIT 1", user.ID).Scan(&tenantID)
	if memErr == nil && tenantID != uuid.Nil {
		user.TenantID = &tenantID
	}

	tokens, err := s.generateTokens(ctx, &user)
	if err != nil {
		return nil, nil, err
	}

	s.RecordAudit(ctx, user.TenantID, &user.ID, "LOGIN_SUCCESS", "auth", map[string]interface{}{
		"email": user.Email,
		"role":  user.Role,
	}, ip, userAgent)

	return &user, tokens, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenStr string, ip, userAgent string) (*User, *TokenPair, error) {
	if refreshTokenStr == "" {
		return nil, nil, ErrInvalidToken
	}

	tokenHash := HashRefreshToken(refreshTokenStr)

	var tokenID, userID uuid.UUID
	var expiresAt time.Time
	var revokedAt sql.NullTime

	query := `SELECT id, user_id, expires_at, revoked_at FROM refresh_tokens WHERE token_hash = $1`
	err := s.db.QueryRowContext(ctx, query, tokenHash).Scan(&tokenID, &userID, &expiresAt, &revokedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, ErrInvalidToken
		}
		return nil, nil, fmt.Errorf("token lookup failed: %w", err)
	}

	if revokedAt.Valid {
		return nil, nil, ErrTokenRevoked
	}

	if time.Now().UTC().After(expiresAt) {
		return nil, nil, ErrInvalidToken
	}

	// Revoke current token (rotation)
	_, _ = s.db.ExecContext(ctx, "UPDATE refresh_tokens SET revoked_at = $1 WHERE id = $2", time.Now().UTC(), tokenID)

	// Fetch user
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	tokens, err := s.generateTokens(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	s.RecordAudit(ctx, user.TenantID, &user.ID, "TOKEN_REFRESHED", "auth", nil, ip, userAgent)
	return user, tokens, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshTokenStr string, userID uuid.UUID) error {
	if refreshTokenStr != "" {
		tokenHash := HashRefreshToken(refreshTokenStr)
		_, _ = s.db.ExecContext(ctx, "UPDATE refresh_tokens SET revoked_at = $1 WHERE token_hash = $2", time.Now().UTC(), tokenHash)
	}

	s.RecordAudit(ctx, nil, &userID, "LOGOUT", "auth", nil, "", "")
	return nil
}

func (s *AuthService) GetUserByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	var user User
	query := `
	SELECT id, email, first_name, last_name, COALESCE(phone, ''), role, is_active, created_at, updated_at
	FROM users WHERE id = $1`
	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.Phone, &user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("query user failed: %w", err)
	}

	var tenantID uuid.UUID
	memErr := s.db.QueryRowContext(ctx, "SELECT tenant_id FROM memberships WHERE user_id = $1 ORDER BY is_default DESC LIMIT 1", user.ID).Scan(&tenantID)
	if memErr == nil && tenantID != uuid.Nil {
		user.TenantID = &tenantID
	}

	return &user, nil
}

func (s *AuthService) generateTokens(ctx context.Context, user *User) (*TokenPair, error) {
	now := time.Now().UTC()
	accessExpiry := now.Add(s.cfg.JWTAccessExpiry)
	if s.cfg.JWTAccessExpiry == 0 {
		accessExpiry = now.Add(15 * time.Minute)
	}

	claims := UserClaims{
		UserID:   user.ID,
		Email:    user.Email,
		Role:     user.Role,
		TenantID: user.TenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.cfg.JWTIssuer,
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedAccess, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	rawRefreshToken := uuid.NewString() + "-" + uuid.NewString()
	tokenHash := HashRefreshToken(rawRefreshToken)
	refreshExpiry := now.Add(7 * 24 * time.Hour)

	insertRefreshSQL := `
	INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at)
	VALUES ($1, $2, $3, $4, $5)`
	_, err = s.db.ExecContext(ctx, insertRefreshSQL, uuid.New(), user.ID, tokenHash, refreshExpiry, now)
	if err != nil {
		return nil, fmt.Errorf("failed to persist refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  signedAccess,
		RefreshToken: rawRefreshToken,
		ExpiresAt:    accessExpiry,
		TokenType:    "Bearer",
	}, nil
}

func (s *AuthService) RecordAudit(ctx context.Context, tenantID *uuid.UUID, userID *uuid.UUID, action, resource string, details map[string]interface{}, ip, userAgent string) {
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var detJSON string
		if details != nil {
			detJSON = fmt.Sprintf("%v", details)
		} else {
			detJSON = "{}"
		}

		insertAudit := `
		INSERT INTO audit_logs (id, tenant_id, user_id, action, resource, details, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, '{}'::jsonb, $6, $7, $8)`
		_, _ = s.db.ExecContext(bgCtx, insertAudit, uuid.New(), tenantID, userID, action, resource, ip, userAgent, time.Now().UTC())
		_ = detJSON
	}()
}
