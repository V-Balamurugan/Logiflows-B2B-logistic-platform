# Decision Log

| ID | Date | Decision | Rationale | Alternatives Considered |
| :--- | :--- | :--- | :--- | :--- |
| **DEC-001** | 2026-10-06 | Core Go backend + Flutter mobile app | High performance, memory safety, cross-platform mobile consistency | Node.js backend, React Native |
| **DEC-002** | 2026-10-06 | PostgreSQL 16 + PostGIS for spatial logistics | Native point-in-polygon queries, GiST spatial indexing, ACID transaction guarantees | MongoDB with 2dsphere, Elasticsearch |
| **DEC-003** | 2026-10-06 | Go owns all authorization and business state; Python strictly advisory | Single source of truth for business rules, zero duplication of business logic | Direct DB writes from Python microservice |
| **DEC-004** | 2026-10-06 | Proceed with Phase 4 after verifying Phases 0–3 | Source code inspection and full test runs confirmed Phases 0–3 are fully green and merged | Re-implementing existing working code |
