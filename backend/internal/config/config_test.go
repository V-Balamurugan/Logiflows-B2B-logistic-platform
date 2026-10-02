package config

import (
	"os"
	"testing"
)

func TestLoadConfig_Defaults(t *testing.T) {
	os.Clearenv()
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error loading defaults, got %v", err)
	}

	if cfg.AppPort != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.AppPort)
	}
	if cfg.AppEnv != "development" {
		t.Errorf("expected env development, got %s", cfg.AppEnv)
	}
	if len(cfg.AllowedOrigins) == 0 {
		t.Errorf("expected default allowed origins to not be empty")
	}
}

func TestLoadConfig_CustomEnv(t *testing.T) {
	os.Setenv("APP_PORT", "9090")
	os.Setenv("APP_ENV", "production")
	os.Setenv("DEBUG", "true")
	defer os.Clearenv()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.AppPort != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.AppPort)
	}
	if cfg.AppEnv != "production" {
		t.Errorf("expected env production, got %s", cfg.AppEnv)
	}
	if !cfg.Debug {
		t.Errorf("expected debug to be true")
	}
}
