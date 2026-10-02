package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Role string

const (
	RolePlatformAdmin Role = "PLATFORM_ADMIN"
	RoleTenant        Role = "TENANT"
	RoleTenantAdmin   Role = "TENANT_ADMIN"
	RoleEmployee      Role = "EMPLOYEE"
	RoleCustomer      Role = "CUSTOMER"
)

func (r Role) IsValid() bool {
	switch r {
	case RolePlatformAdmin, RoleTenant, RoleTenantAdmin, RoleEmployee, RoleCustomer:
		return true
	default:
		return false
	}
}

type Permission string

const (
	PermTenantRead         Permission = "tenant.read"
	PermTenantCreate       Permission = "tenant.create"
	PermTenantUpdate       Permission = "tenant.update"
	PermTenantDisable      Permission = "tenant.disable"
	PermPlatformAdminCreate Permission = "platform_admin.create"
	PermTenantAdminCreate  Permission = "tenant_admin.create"
	PermEmployeeCreate     Permission = "employee.create"
	PermEmployeeRead       Permission = "employee.read"
	PermEmployeeUpdate     Permission = "employee.update"
	PermCustomerCreate     Permission = "customer.create"
	PermCustomerRead       Permission = "customer.read"
	PermCustomerUpdate     Permission = "customer.update"
	PermParcelCreate       Permission = "parcel.create"
	PermParcelRead         Permission = "parcel.read"
	PermParcelUpdate       Permission = "parcel.update"
	PermParcelCancel       Permission = "parcel.cancel"
	PermParcelTrack        Permission = "parcel.track"
	PermParcelQRGenerate   Permission = "parcel.qr.generate"
	PermParcelQRDownload   Permission = "parcel.qr.download"
	PermCustodyRead        Permission = "custody.read"
	PermCustodyCreate      Permission = "custody.create"
	PermCustodyCorrect     Permission = "custody.correct"
	PermDeliveryAssign     Permission = "delivery.assign"
	PermDeliveryUpdate     Permission = "delivery.update"
	PermBranchManage       Permission = "branch.manage"
	PermVehicleManage      Permission = "vehicle.manage"
	PermRouteReview        Permission = "route.review"
	PermReportRead         Permission = "report.read"
	PermAuditRead          Permission = "audit.read"
	PermSystemSettings     Permission = "system.settings"
)

// RolePermissions defines authoritative permissions mapped to each role
var RolePermissions = map[Role][]Permission{
	RolePlatformAdmin: {
		PermTenantRead, PermTenantCreate, PermTenantUpdate, PermTenantDisable,
		PermPlatformAdminCreate, PermTenantAdminCreate, PermEmployeeRead,
		PermBranchManage,
		PermParcelRead, PermParcelTrack, PermReportRead, PermAuditRead, PermSystemSettings,
	},
	RoleTenant: {
		PermTenantRead, PermTenantAdminCreate, PermEmployeeCreate, PermEmployeeRead, PermEmployeeUpdate,
		PermCustomerCreate, PermCustomerRead, PermCustomerUpdate,
		PermParcelCreate, PermParcelRead, PermParcelUpdate, PermParcelCancel, PermParcelTrack,
		PermParcelQRGenerate, PermParcelQRDownload, PermCustodyRead,
		PermDeliveryAssign, PermDeliveryUpdate, PermBranchManage, PermVehicleManage,
		PermRouteReview, PermReportRead, PermAuditRead,
	},
	RoleTenantAdmin: {
		PermTenantRead, PermEmployeeCreate, PermEmployeeRead, PermEmployeeUpdate,
		PermCustomerCreate, PermCustomerRead, PermCustomerUpdate,
		PermParcelCreate, PermParcelRead, PermParcelUpdate, PermParcelCancel, PermParcelTrack,
		PermParcelQRGenerate, PermParcelQRDownload, PermCustodyRead, PermCustodyCreate,
		PermDeliveryAssign, PermDeliveryUpdate, PermBranchManage, PermVehicleManage,
		PermRouteReview, PermReportRead,
	},
	RoleEmployee: {
		PermParcelRead, PermParcelTrack, PermParcelQRGenerate, PermParcelQRDownload,
		PermCustodyRead, PermCustodyCreate, PermDeliveryUpdate, PermRouteReview,
	},
	RoleCustomer: {
		PermCustomerRead, PermCustomerUpdate,
		PermParcelCreate, PermParcelRead, PermParcelCancel, PermParcelTrack,
	},
}

type User struct {
	ID           uuid.UUID  `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	FirstName    string     `json:"first_name"`
	LastName     string     `json:"last_name"`
	Phone        string     `json:"phone,omitempty"`
	Role         Role       `json:"role"`
	IsActive     bool       `json:"is_active"`
	TenantID     *uuid.UUID `json:"tenant_id,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type UserClaims struct {
	UserID   uuid.UUID  `json:"user_id"`
	Email    string     `json:"email"`
	Role     Role       `json:"role"`
	TenantID *uuid.UUID `json:"tenant_id,omitempty"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func HashRefreshToken(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return hex.EncodeToString(h.Sum(nil))
}

func HasPermission(role Role, perm Permission) bool {
	perms, exists := RolePermissions[role]
	if !exists {
		return false
	}
	for _, p := range perms {
		if p == perm {
			return true
		}
	}
	return false
}

func ValidateRole(roleStr string) (Role, error) {
	r := Role(roleStr)
	if !r.IsValid() {
		return "", errors.New("invalid user role")
	}
	return r, nil
}
