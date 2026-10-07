# Phase 4 — Customers, Address Book, Parcel Booking & Authoritative State Engine

## 1. Executive Summary
Phase 4 implements the end-to-end customer consignment booking, address book management, cryptographically signed QR token identity generation, public and authenticated shipment tracking, and table-driven parcel lifecycle state machine for **LogiFlows**. It connects customers, consignments, and spatial locations with multi-tenant isolation, idempotency deduplication, and immutable custody logging.

---

## 2. Key Deliverables

### 2.1 Database & PostGIS 16 Consignment Schema (`0005_customers_parcels_booking.sql`)
- **Customer Profiles Table (`customer_profiles`)**:
  - `id`: UUID primary key referencing `users.id`.
  - `tenant_id`: Multi-tenant isolation scope.
  - `company_name`, `tax_id`, `billing_address`.
  - `tier`: `RETAIL`, `BUSINESS`, `ENTERPRISE`.
  - `credit_limit`, `balance`.
- **Customer Saved Addresses Table (`addresses`)**:
  - `id`: UUID primary key.
  - `user_id`: Customer identity reference.
  - `tenant_id`: Multi-tenant scope.
  - `label`: Friendly name (e.g., "Main Warehouse", "Bengaluru HQ").
  - `contact_name`, `contact_phone`, `contact_email`.
  - `street_line1`, `street_line2`, `city`, `state`, `postal_code`, `country`.
  - `location`: `GEOMETRY(Point, 4326)` for spatial distance calculations and dispatching.
  - `is_default_pickup`, `is_default_delivery`.
- **Parcels Table (`parcels`)**:
  - `id`: UUID primary key.
  - `tenant_id`: Multi-tenant isolation scope.
  - `tracking_number`: Unique authoritative identifier (`LF-YYYYMMDD-XXXXXX`).
  - `qr_token`: Opaque HMAC-SHA256 token (no raw PII embedded).
  - `sender_customer_id`: Authenticated booking customer reference.
  - `service_type`: `STANDARD`, `EXPRESS`, `SAME_DAY`, `COLD_CHAIN`, `FREIGHT`.
  - `status`: Strict lifecycle state machine (`CREATED`, `CONFIRMED`, `RECEIVED_AT_BRANCH`, `ASSIGNED`, `PICKED_UP`, `IN_TRANSIT`, `OUT_FOR_DELIVERY`, `DELIVERED`, `CANCELLED`, `FAILED`, `RETURNED`, `ON_HOLD`, `EXCEPTION`).
  - `weight_kg`, `length_cm`, `width_cm`, `height_cm`, `declared_value`.
  - `shipping_cost`, `currency`.
  - `idempotency_key`: Client-provided deduplication key preventing duplicate charges or bookings.
- **Custody Events Table (`custody_events`)**:
  - Append-only immutable custody chain audit trail.
  - `parcel_id`, `actor_id`, `tenant_id`, `branch_id`.
  - `event_type`: e.g., `PARCEL_CREATED`, `STATUS_UPDATED`, `BRANCH_INTAKE`.
  - `location`: Spatial coordinates of the handover.

---

### 2.2 Parcel Domain Engine & Table-Driven State Machine (`internal/parcel`)
- **Strict State Transition Matrix (`CanTransition`)**:
  - Enforces valid lifecycle transitions and rejects invalid state jumps with 409 Conflict.
- **Authoritative Tracking Number Generator**:
  - Format: `LF-YYYYMMDD-XXXXXX` with cryptographic randomness.
- **Secure Opaque QR Token Generation**:
  - HMAC-SHA256 signature generated without embedding recipient/sender PII into the barcode.
- **Dynamic Tier-Based Shipping Calculator**:
  - Pricing formula combining base tier costs (Standard ₹80, Express ₹150, Same Day ₹250, Cold Chain ₹300, Freight ₹500) with weight increments.
- **Strict Multi-Tenant Isolation & Customer Scoping**:
  - Non-privileged customers only view and manage their own consignments; cross-tenant access is rejected.

---

### 2.3 HTTP API Layer (`internal/httpapi`)
- `POST /api/v1/parcels/book`: Book consignment with idempotency deduplication.
- `GET /api/v1/parcels`: List customer or tenant consignments with pagination.
- `GET /api/v1/parcels/{id}`: Detailed parcel metadata with role scoping.
- `PATCH /api/v1/parcels/{id}/status`: Authoritative status change with custody logging.
- `GET /api/v1/parcels/track/{tracking_number}`: Public tracking endpoint returning custody timeline without leaking PII.
- `POST /api/v1/parcels/scan`: Branch intake scan by QR token.
- `GET /api/v1/parcels/{id}/custody`: Detailed immutable custody history.
- `POST /api/v1/customers/addresses`: Add address to customer address book.
- `GET /api/v1/customers/addresses`: List saved pickup/delivery addresses.

---

### 2.4 Web Portal Consignment Interfaces (`web/src`)
- **Consignment Booking Wizard (`BookParcelModal.tsx`)**:
  - Interactive multi-step form with service tier cards, dynamic price estimates, client UUID idempotency keys, and instant tracking confirmation.
- **Shipment Tracker Modal (`ParcelTrackerModal.tsx`)**:
  - Public & authenticated tracker rendering live custody timeline nodes, transit route details, and status badges.
- **Address Book Manager (`AddressBookView.tsx`)**:
  - Saved addresses for fast pickup/delivery selection.
- **Integrated Customer Portal (`CustomerDashboard.tsx`)**:
  - Live consignment roster querying `/api/v1/parcels`, clickable tracking numbers, and instant modal dispatch.

---

## 3. Phase Gate Verification Table

| Gate Criterion | Verification Method | Status | Evidence |
| :--- | :--- | :--- | :--- |
| **Consignment Booking & Idempotency** | Automated unit & HTTP integration tests | **PASS** | `TestIdempotency` & `TestBookParcelHandler_Success` |
| **Strict State Machine Enforcement** | Unit test matrix (`CanTransition`) | **PASS** | `TestCanTransition` verified across all valid and invalid pairs |
| **Customer Scoping & Multi-Tenant Isolation** | Repository and API isolation tests | **PASS** | `TestCustomerIsolation` rejects unauthorized access |
| **Secure Opaque QR Tokens** | HMAC SHA-256 validation | **PASS** | `TestGenerateQRToken` verifies entropy & absence of PII |
| **Web Frontend Build & Typings** | `tsc && vite build` | **PASS** | 1,506 modules transformed, exit code 0 |
| **Full Pre-Commit & Security Checks** | `scripts/pre-commit-check.ps1` | **PASS** | 0 secrets, go vet clean, all tests passed |

---

## 4. Test Execution Evidence
```text
=== LogiFlows Pre-Commit Quality & Security Verification ===
[1/3] Scanning for accidentally staged secrets / raw keys...
  -> Secret scan clean.
[2/3] Verifying code hygiene...
  -> Go vet passed.
[3/3] Running module tests...
ok  	logiflows/backend/internal/auth	(cached)
ok  	logiflows/backend/internal/config	(cached)
ok  	logiflows/backend/internal/fleet	(cached)
ok  	logiflows/backend/internal/httpapi	(cached)
ok  	logiflows/backend/internal/middleware	(cached)
ok  	logiflows/backend/internal/migrations	(cached)
ok  	logiflows/backend/internal/parcel	(cached)
ok  	logiflows/backend/internal/routing	(cached)
ok  	logiflows/backend/internal/swagger	(cached)
ok  	logiflows/backend/internal/tenancy	(cached)
ok  	logiflows/backend/internal/tracking	(cached)
All pre-commit checks PASSED.
```
