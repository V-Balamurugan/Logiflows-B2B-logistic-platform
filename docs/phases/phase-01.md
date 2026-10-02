# Phase 1 — Authentication, Single Portal & RBAC: Implementation Report

**Phase:** Phase 1 — Authentication, Single Portal & RBAC  
**Branch:** `phase/01-auth-rbac`  
**Status:** COMPLETE  

---

## 1. Summary of Deliverables

Phase 1 completes the authoritative authentication, single portal, role-based access control (RBAC), and multi-tenant isolation foundation:

1. **Database Schema & Migrations (`0002_users_auth.sql`):**
   - Tables created: `tenants`, `users`, `memberships`, `refresh_tokens`, `audit_logs`.
   - Foreign key constraints, composite unique constraints (`uq_membership_tenant_user`), and B-tree indexes for fast email, role, and hash lookups.
2. **Authoritative Core Backend:**
   - Password security: `bcrypt` password hashing with default cost.
   - JWT tokens: Cryptographically signed access tokens (15m expiration) containing claims (`user_id`, `role`, `tenant_id`).
   - Token rotation & revocation: Cryptographically strong UUID refresh tokens tracked by SHA-256 hash in `refresh_tokens`.
   - Centralized RBAC: Granular permissions mapped to roles (`PLATFORM_ADMIN`, `TENANT`, `TENANT_ADMIN`, `EMPLOYEE`, `CUSTOMER`).
   - Composable Go middlewares:
     - `middleware.Authenticate(cfg)`: Validates JWT signature and injects claims into context.
     - `middleware.RequireRole(roles...)`: Restricts endpoints to specific roles.
     - `middleware.RequirePermission(perm)`: Enforces specific operational permissions.
     - `middleware.EnforceTenantIsolation(paramKey)`: Blocks cross-tenant data leaks and prevents IDOR.
   - Endpoints: `POST /api/v1/auth/register`, `POST /api/v1/auth/login`, `POST /api/v1/auth/refresh`, `POST /api/v1/auth/logout`, `GET /api/v1/auth/me`.
   - Seed CLI tool: `backend/cmd/seed-admin/main.go` seeding test accounts for all 5 roles.
3. **Single Portal Web Application (`web/`):**
   - Unified authentication portal at `/login` and `/register`.
   - Client never decides role via dropdown; role redirection is driven strictly by backend claims.
   - 5 Role Dashboards:
     - Platform Admin (`/admin/dashboard`)
     - Tenant Owner (`/tenant/dashboard`)
     - Tenant Admin / Dispatcher (`/tenant-admin/dashboard`)
     - Employee / Courier (`/employee/dashboard`)
     - Customer (`/customer/dashboard`)
   - `AuthContext` with automatic token storage and permission helpers.
4. **Testing & Security:**
   - Unit tests for password hashing, RBAC permissions, token hashing.
   - Middleware security tests:
     - Valid token authorization test.
     - Expired token rejection test.
     - Role requirement rejection test (Customer trying to access Admin endpoint -> 403 Forbidden).
     - Permission requirement test (Customer with `parcel.create` allowed; Customer trying `tenant.create` -> 403 Forbidden).
     - Tenant isolation IDOR test (Tenant A attempting to access Tenant B resource -> 403 Forbidden).
