# ADR-001: PostgreSQL + PostGIS as Transactional & Geospatial Source of Truth

## Status
Accepted

## Context
LogiFlows requires high-integrity multi-tenant transactional logistics operations (parcels, custody handovers, assignments, fleet assets) alongside advanced geospatial query capabilities (serviceability point-in-polygon checks, GiST-indexed proximity searches, geofences, and coordinate tracking).

## Decision
Adopt PostgreSQL 16 with the PostGIS 3.4+ extension as the single persistent source of truth. Redis is utilized exclusively for ephemeral pub/sub, real-time transient coordinate caching, and rate limiting—never as primary persistence.

## Consequences
- ACID compliance and atomic transactions across custody chains and multi-step dispatch operations.
- Native spatial operators (`ST_Contains`, `ST_Distance`, `ST_MakePoint`, GiST indexing) deliver sub-millisecond serviceability resolution without third-party spatial databases.
- Relational foreign keys and composite constraints guarantee strict tenant isolation.
