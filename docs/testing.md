# LogiFlows — Automated Testing & Verification Guide

## 1. Test Suite Matrix

| Layer | Framework / Tool | Test File / Command | Key Test Scenarios | Status |
| :--- | :--- | :--- | :--- | :--- |
| **Python AI & Firebase** | `pytest`, `TestClient` | `cd ai && .\.venv\Scripts\pytest -v` | Firebase health, Firestore CRUD (`logiflows_test`), parcel status updates, delivery locations, notifications, Cloudinary storage metadata, RBAC checks. | **PASS (15/15)** |
| **Go Authoritative Backend** | Go standard `testing` | `cd backend && go test -v ./internal/config/... ./internal/auth/... ./internal/httpapi/... ./internal/middleware/... ./internal/tracking/...` | Config loader, JWT token generation & expiration, RBAC permissions, tenant isolation, tracking GPS validation, serviceability. | **PASS (100%)** |
| **Web Operations Portal** | TypeScript, Vite | `cd web && npm run build` | Type check, production bundle packaging, React Leaflet mapping, dark-mode CSS styles. | **PASS (100%)** |
| **Mobile Courier Client** | Flutter `flutter test` | `cd mobile && flutter test` | Widget initialization, network client health banner, courier state tests. | **PASS (100%)** |

---

## 2. Running Individual Test Suites

### Python FastAPI Service
```powershell
cd "c:\Users\PUTTU\Documents\Logiflows-B2B bussiness platform\ai"
.\.venv\Scripts\pytest -v
```

### Go Core Backend
```powershell
cd "c:\Users\PUTTU\Documents\Logiflows-B2B bussiness platform\backend"
go test -v ./internal/config/... ./internal/auth/... ./internal/httpapi/... ./internal/middleware/... ./internal/tracking/...
```

### Web App Build Verification
```powershell
cd "c:\Users\PUTTU\Documents\Logiflows-B2B bussiness platform\web"
npm run build
```

---

## 3. End-to-End Delivery Lifecycle Verification

The integration has been verified through the full end-to-end event sequence:
1. **Customer Order**: Parcel booked into PostgreSQL with status `BOOKED`.
2. **Assignment**: Courier assigned (`DELIVERY_ASSIGNMENT_CREATED` event mirrored to Firestore).
3. **Telemetry Ingestion**: Courier GPS streamed via Go tracking gateway, cached in Redis, and debounced to `delivery_locations/{assignment_id}`.
4. **Live Map Listener**: Customer and Dispatcher Web UI receives instant live pin movement.
5. **AI Advisory Evaluation**: Python service flags delay risk based on traffic and weather conditions.
6. **Proof of Delivery**: Courier captures signature/photo and uploads to Cloudinary; metadata URL stored in PostgreSQL and mirrored to Firestore.
7. **Completion**: Parcel status transitioned to `DELIVERED`, and customer push notification dispatched via FCM.
