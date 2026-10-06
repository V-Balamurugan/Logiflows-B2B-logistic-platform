# Test Case Repository

## Phase 0–3 Baseline Test Cases

### TC-AUTH-001: JWT Validation & Signature Enforcement
- **Module:** Auth / Middleware
- **Requirement ID:** US-1-03
- **Priority:** P1 | **Severity:** Critical
- **Type:** Unit / Security
- **Precondition:** Secret configured
- **Steps:** Sign JWT with secret; verify claims; mutate payload; verify rejection
- **Automated test location:** `backend/internal/middleware/middleware_test.go:TestAuthenticate_ValidToken` & `TestAuthenticate_ExpiredToken`
- **Status:** PASS

### TC-AUTH-002: RBAC Deny-by-Default Role Guard
- **Module:** Auth / Middleware
- **Requirement ID:** US-1-08
- **Priority:** P1 | **Severity:** Critical
- **Type:** API / Security
- **Precondition:** Authenticated user with role CUSTOMER
- **Steps:** Attempt request to admin route; expect HTTP 403 Forbidden
- **Automated test location:** `backend/internal/middleware/middleware_test.go:TestRequireRole_Forbidden`
- **Status:** PASS

### TC-TENANT-001: PostGIS Serviceability Point-in-Polygon
- **Module:** Tenancy / Geospatial
- **Requirement ID:** US-2-02
- **Priority:** P1 | **Severity:** High
- **Type:** Integration
- **Precondition:** Branch created with spatial polygon boundary
- **Steps:** Query point inside polygon -> serviceable=true; query point outside -> serviceable=false
- **Automated test location:** `backend/internal/tenancy/tenancy_test.go`
- **Status:** PASS

### TC-FLEET-001: Telematics Ingestion & Breadcrumbs
- **Module:** Fleet / Telematics
- **Requirement ID:** US-3-02
- **Priority:** P1 | **Severity:** High
- **Type:** Unit / Integration
- **Precondition:** Vehicle registered with VIN and license plate
- **Steps:** Send GPS coordinates; retrieve latest position; retrieve historical trajectory breadcrumbs
- **Automated test location:** `backend/internal/fleet/fleet_test.go`
- **Status:** PASS
