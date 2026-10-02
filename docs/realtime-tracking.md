# Realtime Delivery Tracking & Telemetry Pipeline

## Ingestion Architecture

```
Mobile Driver App
      │
      │ HTTP POST /api/v1/tracking/location (every 1-3s)
      ▼
Go Tracking Gateway (:8080)
      │
      ├─► Redis In-Memory Cache (TTL: 24h, < 1ms access)
      │     └─ Key: `delivery_location:{assignment_id}`
      │
      ├─► Redis Pub/Sub Fanout
      │     └─ Channel: `tracking:{assignment_id}`
      │           │
      │           └─► Go Gorilla WebSocket Stream (:8080/api/v1/tracking/ws/{assignment_id})
      │                 └─► React Live Dispatch & Customer Map Dashboard
      │
      └─► Debounced Firestore Syncer (every 5 seconds)
            └─ Collection: `delivery_locations/{assignment_id}`
```

## Endpoints

1. **Submit Location Ping (Mobile App):**
   - `POST /api/v1/tracking/location`
   - Payload:
     ```json
     {
       "assignment_id": "asg_demo_101",
       "employee_id": "emp_courier_01",
       "latitude": 9.9252,
       "longitude": 78.1198,
       "speed": 35.5,
       "heading": 120.0
     }
     ```
2. **Fetch Latest Location (REST):**
   - `GET /api/v1/tracking/location/{assignment_id}`
3. **Live WebSocket Stream (Web / Customer):**
   - `GET /api/v1/tracking/ws/{assignment_id}`
