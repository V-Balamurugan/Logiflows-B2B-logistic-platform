package migrations

import (
	"strings"
	"testing"
)

func TestEmbeddedMigrations(t *testing.T) {
	entries, err := migrationFS.ReadDir("sql")
	if err != nil {
		t.Fatalf("failed to read embedded sql dir: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("expected embedded migration files, got 0")
	}

	foundFleet := false
	for _, entry := range entries {
		if entry.Name() == "0004_fleet_telematics.sql" {
			foundFleet = true
			content, err := migrationFS.ReadFile("sql/" + entry.Name())
			if err != nil {
				t.Fatalf("failed to read 0004_fleet_telematics.sql: %v", err)
			}
			sqlStr := string(content)
			if !strings.Contains(sqlStr, "CREATE TABLE IF NOT EXISTS vehicles") {
				t.Errorf("expected vehicles table definition in 0004_fleet_telematics.sql")
			}
			if !strings.Contains(sqlStr, "CREATE TABLE IF NOT EXISTS vehicle_telematics") {
				t.Errorf("expected vehicle_telematics table definition in 0004_fleet_telematics.sql")
			}
			if !strings.Contains(sqlStr, "CREATE TABLE IF NOT EXISTS maintenance_records") {
				t.Errorf("expected maintenance_records table definition in 0004_fleet_telematics.sql")
			}
		}
	}

	if !foundFleet {
		t.Errorf("0004_fleet_telematics.sql was not found in embedded migrations")
	}
}
