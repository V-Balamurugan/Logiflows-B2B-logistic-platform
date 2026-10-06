# LogiFlows — Firebase & Firestore Integration Architecture

## 1. Discovered Architecture Summary

LogiFlows is an AI-powered Smart Postal & Courier Delivery Management System built on a high-throughput, multi-tenant hybrid architecture:

| Tier | Component | Technology | Role & Authority |
| :--- | :--- | :--- | :--- |
| **Frontend Portal** | Web App | React 18, TypeScript, Tailwind CSS, Leaflet | Single multi-tenant portal for Admin, Tenant, Courier, and Customer. Listens to Firestore realtime mirrors. |
| **Mobile App** | Courier App | Flutter 3.22+, Dart | Courier field application with QR scanning, GPS tracking transmission, and offline queue. |
| **Authoritative Core** | Backend Gateway | Go 1.22+ (Chi, PostGIS, JWT) | Authoritative owner of business state machines, RBAC, PostGIS geospatial resolution, and high-frequency WebSocket GPS telemetry. |
| **AI Advisory Service**| AI Microservice | Python 3.11+ (FastAPI, Scikit-learn) | Delay risk forecasting, travel time prediction, route intelligence, and Firestore admin pipelines. |
| **Primary Database** | Relational DB | PostgreSQL 16 + PostGIS | **Primary transactional source of truth** for users, branches, parcels, delivery assignments, and audit logs. |
| **Transient Cache** | Pub/Sub & Cache | Redis 7 Alpine | High-frequency telemetry cache, WebSocket subscription broker, and rate limiting. |
| **Realtime Service** | Document Store | Cloud Firestore | **Supporting realtime/data service** for live delivery state, driver locations, parcel status mirrors, and notifications. |
| **Media Storage** | Cloud Storage | Cloudinary | **Dedicated media store** for Proof of Delivery (POD) photos, recipient signatures, parcel images, and documents. *(Firebase Storage is not used).* |
| **Push Alerts** | Notifications | FCM, n8n, Brevo | Multi-channel push alerts for delivery assignments, status changes, and delay predictions. |

---

## 2. Core Architectural Rule

**DO NOT replace PostgreSQL with Firestore.**
- **PostgreSQL**: Stores authoritative relational state, tenant configurations, branch polygons, parcel lifecycle transactions, courier assignments, and historical records.
- **Firestore**: Stores live transient states, live courier coordinates, push notification queues, device tokens, and audit event streams for client listeners.
- **Cloudinary**: Stores all binary media (POD photos, customer signatures, waybill PDFs). Metadata URLs and public IDs are stored in PostgreSQL and Firestore.
- **Go**: Handles WebSocket connections and high-frequency GPS telemetry, caching in Redis and debouncing writes to Firestore.
- **FastAPI**: Handles advisory AI, delay risk prediction, route optimization, and Firestore background operations.

---

## 3. Data Flow Diagram

```mermaid
flowchart TD
    Customer([Customer / Client])
    Courier([Courier Mobile App])
    Dispatcher([Dispatcher / Admin])

    subgraph Frontends
        ReactWeb[React Web Portal]
        FlutterApp[Flutter Mobile App]
    end

    subgraph CoreBackend [Authoritative Backend - Go 8080]
        GoAPI[Go Chi Router / RBAC]
        GoWS[Go WebSocket Ingestion]
        PostGIS[(PostgreSQL 16 + PostGIS)]
        RedisCache[(Redis 7 Telemetry)]
    end

    subgraph AIService [Python AI Advisory - FastAPI 8000]
        FastAPIApp[FastAPI Endpoints]
        MLModels[Delay Risk Classifier]
        PyFirebase[Firebase Admin SDK]
    end

    subgraph FirebaseCloud [Firebase & Supporting Cloud]
        FirestoreDB[(Cloud Firestore)]
        FCMService[Firebase Cloud Messaging]
        CloudinaryStore[Cloudinary Media Storage]
    end

    Courier -->|GPS via WebSocket| GoWS
    GoWS -->|Ephemeral Cache| RedisCache
    GoWS -->|Debounced Sync (5s)| FirestoreDB
    Courier -->|Upload POD Photo/Sig| CloudinaryStore

    Customer -->|Book Parcel / Auth| GoAPI
    GoAPI -->|ACID Transaction| PostGIS
    GoAPI -->|Mirror Status| FirestoreDB
    GoAPI -->|Trigger Delay Risk| FastAPIApp
    FastAPIApp -->|Predict Risk| MLModels
    FastAPIApp -->|Publish System Event| FirestoreDB
    FastAPIApp -->|Dispatch Push Alert| FCMService

    ReactWeb -->|Realtime Snapshot Listener| FirestoreDB
    ReactWeb -->|REST Management| GoAPI
```

---

## 4. Endpoints & Integrations

- **Health Probe**: `GET /api/firebase/health` and `GET /api/v1/firebase/health`
- **CRUD Verification**:
  - `POST /api/firebase/test`
  - `GET /api/firebase/test/{id}`
  - `PUT /api/firebase/test/{id}`
  - `DELETE /api/firebase/test/{id}`
- **Telemetry Update**: `POST /api/v1/tracking/location` (Go backend, latitude/longitude validated)
- **WebSocket Stream**: `GET /api/v1/tracking/ws/{assignment_id}`
