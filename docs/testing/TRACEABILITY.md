# Requirements Traceability Matrix (RTM)

| Requirement ID | Phase | Feature Title | Test Case ID | Test Level | Automated Test Location | Execution Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **US-0-01** | P0 | Toolchain & Scripts | TC-ENV-001 | System | `scripts/check-tools.ps1` | PASS |
| **US-1-01** | P1 | Users & RBAC Schema | TC-AUTH-001 | Integration | `internal/migrations/migrate_test.go` | PASS |
| **US-1-03** | P1 | Login with JWT & Refresh | TC-AUTH-002 | Unit/Security | `internal/middleware/middleware_test.go` | PASS |
| **US-1-08** | P1 | RBAC Deny-by-Default Guard | TC-AUTH-003 | Unit/Security | `internal/middleware/middleware_test.go` | PASS |
| **US-2-01** | P2 | Enterprise Tenants & Branches | TC-TENANT-001 | Unit/Integration | `internal/tenancy/tenancy_test.go` | PASS |
| **US-2-02** | P2 | PostGIS Serviceability Engine | TC-TENANT-002 | Integration | `internal/tenancy/integration_test.go` | PASS |
| **US-3-01** | P3 | Fleet Vehicle Inventory | TC-FLEET-001 | Unit | `internal/fleet/fleet_test.go` | PASS |
| **US-3-02** | P3 | Sub-second Telematics Ingestion | TC-FLEET-002 | Unit/Integration | `internal/fleet/fleet_test.go` | PASS |
| **US-3-03** | P3 | Resilient Routing Engine | TC-ROUTING-001 | Unit/Mock | `internal/routing/routing_test.go` | PASS |
| **US-4-01** | P4 | Customers & Addresses Schema | TC-CUST-001 | Integration | `internal/migrations/migrate_test.go` | PLANNED |
| **US-4-02** | P4 | Customer Address Book Management| TC-CUST-002 | API | `internal/httpapi/customer_test.go` | PLANNED |
| **US-4-03** | P4 | Parcel Booking & State Machine | TC-PARCEL-001 | Unit/Integration | `internal/parcel/parcel_test.go` | PLANNED |
| **US-4-04** | P4 | Idempotent Booking & Tracking ID | TC-PARCEL-002 | API | `internal/httpapi/parcel_test.go` | PLANNED |
| **US-4-05** | P4 | Customer Parcel Portal & UI | TC-UI-001 | Component | `web/tests/booking.test.js` | PLANNED |
