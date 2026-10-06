# LogiFlows — Changelog

All notable changes to the LogiFlows platform will be documented in this file.

## [Unreleased] - 2026-10-02

### Added
- **Firebase Dual-Ecosystem Integration**:
  - Python Admin SDK (`firebase-admin`) in AI service with Firestore CRUD, collection querying, and Storage signed URLs.
  - Go Admin SDK (`firebase.google.com/go/v4`) in authoritative core backend with FCM notification triggers and proof-of-delivery storage utilities.
- **Realtime Delivery Telemetry & WebSocket Pipeline**:
  - `POST /api/v1/tracking/location`: High-frequency mobile GPS ping ingestion into Redis.
  - `GET /api/v1/tracking/location/{assignment_id}`: Sub-millisecond cached location retrieval.
  - `GET /api/v1/tracking/ws/{assignment_id}`: Gorilla WebSocket stream for real-time customer and dispatcher map tracking.
  - Debounced background sync to Cloud Firestore collection `delivery_locations`.
- **Security & Authorization Rules**:
  - `firestore.rules`: Role-based rules enforcing default deny, separate boundaries for `ADMIN`, `EMPLOYEE`, and `CUSTOMER`.
  - `firebase.storage.rules`: Multi-path access limits for Proof of Delivery (`proof-of-delivery/`), assignment logs, parcel documents, and employee credentials.
- **Client-Side Configuration**:
  - `web/src/firebase/config.ts`: Safe public configuration ensuring no private keys or service accounts leak to the browser.
- **Comprehensive Test Suite**:
  - 10 passing unit and integration tests across health, delay prediction, Firestore CRUD, location streaming, and role verification.
