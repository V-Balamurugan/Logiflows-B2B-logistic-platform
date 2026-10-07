# Regression Test Plan

## Regression Suite Execution
Regression test suites must be executed:
1. After every feature implementation loop prior to commit.
2. Prior to closing any phase gate report.
3. Automatically in CI on pull requests to `develop` and `main`.

## Suite Composition
1. **Core Go Backend**: `go test ./...` in `backend/` verifying configuration, database migrations, auth/RBAC middleware, fleet operations, tenancy serviceability, and HTTP API handlers.
2. **AI Microservice**: `pytest` in `ai/` verifying prediction endpoints and fallback schemas.
3. **Web Portal**: `npm test` and `npm run build` in `web/` guaranteeing bundle integrity and type safety.
4. **Mobile Client**: `flutter test` in `mobile/` ensuring clean UI rendering and API client reliability.

## Automated Execution Command
```powershell
powershell -ExecutionPolicy Bypass -File scripts/test-all.ps1
```
