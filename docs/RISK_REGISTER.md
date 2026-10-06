# Risk Register

| Risk ID | Description | Severity | Likelihood | Mitigation Strategy | Status |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **RSK-001** | Scope too large | High | Medium | Enforce strict MoSCoW prioritization; defer Could items to later sprints. | Active |
| **RSK-002** | Cross-tenant data leakage | Critical | Low | Repository-level `tenant_id` enforcement; negative IDOR and cross-tenant unit/integration tests in every phase. | Active |
| **RSK-003** | Mobile offline sync duplicates | High | Medium | Client-generated UUID idempotency keys on all mutations; backend deduplication table. | Active |
| **RSK-004** | External provider outages (geocoding, routing, SMS) | Medium | Medium | Wrap all external services behind interfaces; implement local fallback (e.g. geodesic routing fallback). | Mitigated (Phase 3) |
| **RSK-005** | AI delay-risk underperformance | Medium | Low | Maintain deterministic rule-based heuristic baseline; AI model sits behind a feature flag with automatic fallback. | Active |
| **RSK-006** | Documentation drift | Medium | Medium | Make documentation updates a mandatory component of Definition of Done in every feature loop. | Active |
| **RSK-007** | Unverified completion claims | High | Low | Enforce strict status vocabulary (PLANNED, IMPLEMENTED, TESTED, VERIFIED) backed by actual CLI execution evidence. | Active |
| **RSK-008** | Tooling & environment variance | Medium | Low | Reproducible PowerShell scripts (`check-tools.ps1`, `test-all.ps1`, `pre-commit-check.ps1`). | Active |
