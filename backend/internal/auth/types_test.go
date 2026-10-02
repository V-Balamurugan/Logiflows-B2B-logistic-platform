package auth

import (
	"testing"
)

func TestPasswordHashing(t *testing.T) {
	password := "SecurePassword123!"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("expected no error hashing password, got %v", err)
	}

	if !CheckPassword(password, hash) {
		t.Fatal("expected password check to succeed")
	}

	if CheckPassword("WrongPassword", hash) {
		t.Fatal("expected password check with wrong password to fail")
	}
}

func TestRBACPermissions(t *testing.T) {
	// Platform admin permissions
	if !HasPermission(RolePlatformAdmin, PermTenantCreate) {
		t.Errorf("platform admin must have tenant.create permission")
	}
	if HasPermission(RoleCustomer, PermTenantCreate) {
		t.Errorf("customer must NOT have tenant.create permission")
	}

	// Customer permissions
	if !HasPermission(RoleCustomer, PermParcelCreate) {
		t.Errorf("customer must have parcel.create permission")
	}
	if !HasPermission(RoleCustomer, PermParcelTrack) {
		t.Errorf("customer must have parcel.track permission")
	}

	// Employee permissions
	if !HasPermission(RoleEmployee, PermCustodyCreate) {
		t.Errorf("employee must have custody.create permission")
	}
	if HasPermission(RoleEmployee, PermTenantDisable) {
		t.Errorf("employee must NOT have tenant.disable permission")
	}
}

func TestHashRefreshToken(t *testing.T) {
	token := "sample-refresh-token-xyz"
	hash1 := HashRefreshToken(token)
	hash2 := HashRefreshToken(token)

	if hash1 != hash2 {
		t.Fatalf("hash of identical token must match")
	}
	if len(hash1) == 0 {
		t.Fatalf("hash must not be empty")
	}
}
