# LogiFlows — Intelligent Logistics Platform

> **AI-Powered Smart Postal & Courier Delivery Management System with Predictive Delay Intelligence, Dynamic Route Optimization, Intelligent Delivery Management and Real-Time Tracking**

---

## 1. Architectural Highlights

- **Authoritative Backend:** Go 1.22+ (`backend/`)
  - Authoritative owner of business logic, state machines, RBAC, tenant boundaries, QR custody, and dispatch orchestration.
  - Interactive **Swagger UI** available at `/swagger` and `/docs`; raw OpenAPI 3.0 specification at `/api/v1/openapi.yaml`.
- **Transactional & Spatial Database:** PostgreSQL 16 + PostGIS (`backend/internal/migrations/`)
  - Stores all relational entities and geospatial geometries for hub serviceability and GPS telemetry trails.
- **Transient Real-Time & Caching:** Redis 7 (`backend/internal/database/redis.go`)
  - Pub/Sub message broker and ephemeral GPS telemetry cache with graceful fallback.
- **Advisory AI Intelligence:** Python + FastAPI (`ai/`)
  - Machine learning delay risk forecasting (`LOW`, `MEDIUM`, `HIGH`) and route performance evaluations.
- **Web Operations Portal:** React 18 + TypeScript + Vite + Tailwind CSS (`web/`)
  - Unified single-portal architecture projecting role-specific interfaces with live infrastructure health monitoring.
- **Mobile Courier Operations:** Flutter + Dart (`mobile/`)
  - Employee operations application with clean architecture, offline caching, and telemetry heartbeat.

---

## 2. Master Implementation Roadmap

| Phase | Description | Status | Branch |
| :--- | :--- | :--- | :--- |
| **0** | **Foundation (Core Scaffolds, PostGIS, Health, CI)** | **COMPLETE** | `phase/00-foundation` |
| **1** | **Authentication, Single Portal & RBAC** | **COMPLETE** | `phase/01-auth-rbac` |
| **2** | Tenants, Branches & Geospatial Serviceability | Planned | `phase/02-tenants-branches` |
| **3** | Employees, Roles & Vehicles | Planned | `phase/03-employees-vehicles` |
| **4** | Customers & Parcel Booking | Planned | `phase/04-parcel-booking` |
| **5** | QR Identity, QR PDF & Custody State Machine | Planned | `phase/05-qr-custody` |
| **6** | Assignment & Dispatch Delivery Workflows | Planned | `phase/06-assignment-delivery` |
| **7** | Flutter Mobile Courier Application | Planned | `phase/07-mobile` |
| **8** | Real-Time GPS Tracking & Pub/Sub Pipeline | Planned | `phase/08-realtime-tracking` |
| **9** | React Live Tracking & Dispatcher Dashboards | Planned | `phase/09-live-dashboard` |
| **10**| Route Planning & Baseline ETA Calculation | Planned | `phase/10-routing-eta` |
| **11**| Dynamic ETA & Exception Handling | Planned | `phase/11-dynamic-eta` |
| **12**| Python AI Delay Risk Prediction Integration | Planned | `phase/12-ai-delay-prediction` |
| **13**| Firebase Notifications & Proof of Delivery | Planned | `phase/13-firebase-notifications` |
| **14**| Analytics, Reporting & SLA Performance | Planned | `phase/14-analytics` |
| **15**| Security Hardening, Scalability & Resilience | Planned | `phase/15-security-reliability` |
| **16**| End-to-End System Acceptance Testing | Planned | `phase/16-testing` |
| **17**| Production Deployment & Cloud Orchestration | Planned | `phase/17-deployment` |
| **18**| Optimization, User Manuals & Final Release | Planned | `phase/18-finalization` |

---

## 3. Quickstart & Local Development

### Prerequisites
- Go 1.22+
- Node.js 20+ & npm
- Python 3.11+
- Flutter 3.22+
- Docker & Docker Compose

### Starting the Platform with Docker Compose
```bash
# Start all infrastructure and services
docker compose up --build
```

### Running Services Independently
```bash
# 1. Authoritative Go Backend (Port 8080)
cd backend
go run ./cmd/api

# 2. React Web Portal (Port 5173)
cd web
npm install
npm run dev

# 3. Python AI Advisory Service (Port 8000)
cd ai
python -m venv .venv
.\.venv\Scripts\Activate.ps1
pip install -r requirements.txt
uvicorn app.main:app --port 8000 --reload

# 4. Flutter Employee Mobile App
cd mobile
flutter pub get
flutter run
```

---

## 4. Verification & Testing

```bash
# Backend Go Tests
cd backend && go test -v ./...

# Web Frontend Build & Typecheck
cd web && npm run build

# Python AI Tests
cd ai && pytest -v tests/

# Flutter Mobile Tests
cd mobile && flutter test
```
