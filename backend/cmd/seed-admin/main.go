package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	"logiflows/backend/internal/auth"
	"logiflows/backend/internal/config"
	"logiflows/backend/internal/database"
	"logiflows/backend/internal/logging"
	"logiflows/backend/internal/migrations"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		os.Exit(1)
	}

	logger := logging.NewLogger("info", false)
	logger.Info("Starting database seed tool")

	db, err := database.NewPostgresDB(cfg, logger)
	if err != nil || db == nil {
		logger.Error("Database connection failed", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Ensure migrations are up to date
	if err := migrations.RunMigrations(ctx, db, logger); err != nil {
		logger.Error("Migrations failed during seed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// 1. Seed Demo Tenant
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	tenantName := "Speedy Express Logistics"
	tenantCode := "speedy-express"

	_, err = db.ExecContext(ctx, `
	INSERT INTO tenants (id, name, code, is_active)
	VALUES ($1, $2, $3, true)
	ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name`,
		tenantID, tenantName, tenantCode,
	)
	if err != nil {
		logger.Error("Failed to seed tenant", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("Seeded demo tenant", slog.String("code", tenantCode))

	// 2. Seed Users
	users := []struct {
		Email     string
		Password  string
		FirstName string
		LastName  string
		Role      auth.Role
		TenantID  *uuid.UUID
	}{
		{
			Email:     "admin@logiflows.io",
			Password:  "Admin@LogiFlows2026!",
			FirstName: "Platform",
			LastName:  "SuperAdmin",
			Role:      auth.RolePlatformAdmin,
			TenantID:  nil,
		},
		{
			Email:     "tenant@speedycourier.com",
			Password:  "Tenant@Speedy2026!",
			FirstName: "Alice",
			LastName:  "Owner",
			Role:      auth.RoleTenant,
			TenantID:  &tenantID,
		},
		{
			Email:     "manager@speedycourier.com",
			Password:  "Manager@Speedy2026!",
			FirstName: "Bob",
			LastName:  "HubManager",
			Role:      auth.RoleTenantAdmin,
			TenantID:  &tenantID,
		},
		{
			Email:     "courier@speedycourier.com",
			Password:  "Courier@Speedy2026!",
			FirstName: "Charlie",
			LastName:  "Rider",
			Role:      auth.RoleEmployee,
			TenantID:  &tenantID,
		},
		{
			Email:     "customer@gmail.com",
			Password:  "Customer@LogiFlows2026!",
			FirstName: "Diana",
			LastName:  "Shopper",
			Role:      auth.RoleCustomer,
			TenantID:  nil,
		},
	}

	for _, u := range users {
		pwdHash, _ := auth.HashPassword(u.Password)
		var uid uuid.UUID
		err := db.QueryRowContext(ctx, `
		INSERT INTO users (email, password_hash, first_name, last_name, role, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
		ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash, role = EXCLUDED.role
		RETURNING id`,
			u.Email, pwdHash, u.FirstName, u.LastName, string(u.Role),
		).Scan(&uid)
		if err != nil {
			logger.Error("Failed to seed user", slog.String("email", u.Email), slog.String("error", err.Error()))
			continue
		}

		if u.TenantID != nil {
			_, _ = db.ExecContext(ctx, `
			INSERT INTO memberships (tenant_id, user_id, role, is_default)
			VALUES ($1, $2, $3, true)
			ON CONFLICT (tenant_id, user_id) DO NOTHING`,
				*u.TenantID, uid, string(u.Role),
			)
		}

		logger.Info("Seeded user successfully", slog.String("email", u.Email), slog.String("role", string(u.Role)))
	}

	logger.Info("Database seeding completed successfully!")
}
