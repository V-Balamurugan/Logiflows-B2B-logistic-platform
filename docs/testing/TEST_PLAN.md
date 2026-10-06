# Master Test Plan

## 1. Overview & Strategy
LogiFlows employs a strict testing strategy ensuring every feature is accompanied by automated unit, integration, API, and regression tests.
No story is marked DONE without executed passing test evidence.

## 2. Testing Levels & Tooling
- **Unit Testing:**
  - Go: `testing` + `testify`
  - Python: `pytest`
  - Web: Node test runner (`node --test`) + Vitest / React Testing Library
  - Mobile: `flutter_test`
- **Integration Testing:**
  - Go database repositories against real PostgreSQL 16 PostGIS container.
  - End-to-end API HTTP calls testing JWT tampering, tenant isolation, and RBAC authorization boundaries.
- **Regression Testing:**
  - `scripts/test-all.ps1` executes the full suite across all four subsystems before every phase gate merge.

## 3. Defect Classification
- **Critical:** Platform unusable, data loss, cross-tenant isolation breach, or auth bypass. Blocks phase gate immediately.
- **High:** Major workflow broken with no workaround. Blocks phase gate.
- **Medium:** Minor functional impairment with acceptable workaround.
- **Low:** Cosmetic or UI polish deficiency.
