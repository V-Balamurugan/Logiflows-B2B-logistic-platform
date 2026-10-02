# Firebase Architecture in LogiFlows

## Architectural Role of Firebase

In LogiFlows, Firebase acts strictly as an **ephemeral and real-time supporting cloud service**. It operates alongside PostgreSQL, Redis, and Go without compromising transactional integrity or business authority:

1. **Transactional & Spatial Authority: PostgreSQL 16 + PostGIS**
   - Retains exclusive ownership over business entities (Tenants, Branches, Users, Parcels, Custody records, Pricing, SLA timers).
2. **Authoritative State Enforcement: Go 1.22+**
   - Authoritative validator of state transitions, RBAC permissions, and QR custody scans.
3. **Transient High-Frequency Ingestion: Redis 7**
   - Buffers rapid GPS telemetry updates (1-2 Hz from mobile devices) with automatic TTL.
4. **Live Realtime Mirror & Push Alerts: Firebase**
   - **Firestore:** Broadcasts live driver locations, customer-facing parcel tracking events, and user notification feeds.
   - **FCM (Firebase Cloud Messaging):** Dispatches push notifications to couriers for new route assignments and to customers for delivery status milestones.
   - **Firebase Storage:** Archives high-throughput media including Proof of Delivery (POD) courier photos, recipient digital signatures, and bill-of-lading scans.
