# ADR-003: Single Portal, Central RBAC & Multi-Tenant Isolation Model

## Status
Accepted

## Context
LogiFlows serves multiple user archetypes: `PLATFORM_ADMIN`, `TENANT`, `TENANT_ADMIN`, `EMPLOYEE`, and `CUSTOMER`. We need a unified web experience while enforcing strict backend access boundaries and multi-tenant isolation.

## Decision
- Single entry-point portal (`/login`, `/register`).
- The client frontend never selects or dictates roles. The backend authenticates credentials, signs a JWT containing the authoritative role and tenant ID, and the client redirects accordingly:
  - `PLATFORM_ADMIN` -> `/admin/dashboard`
  - `TENANT` -> `/tenant/dashboard`
  - `TENANT_ADMIN` -> `/tenant-admin/dashboard`
  - `EMPLOYEE` -> `/employee/dashboard`
  - `CUSTOMER` -> `/customer/dashboard`
- Deny-by-default RBAC enforced at Go middleware layer (`RequirePermission`, `RequireRole`).
- Repositories strictly scope all operational queries by `tenant_id` to guarantee tenant isolation.

## Consequences
- Impossible for users to escalate privileges via client-side UI tampering.
- Clean separation of concerns with unified front-end authentication routing.
