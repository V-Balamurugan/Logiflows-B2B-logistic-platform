# Cloud Firestore Realtime Data Model

Firestore serves as a real-time event and status mirror for user-facing applications. The canonical transactional truth remains in PostgreSQL.

## Logical Collections

### 1. `users/{user_id}`
Client profile mirror and active role indicator:
```json
{
  "uid": "uuid-or-firebase-uid",
  "email": "courier@speedycourier.com",
  "role": "EMPLOYEE",
  "name": "Charlie Rider",
  "phone": "+919876543210",
  "tenant_id": "00000000-0000-0000-0000-000000000001",
  "is_active": true,
  "updated_at": "2026-10-02T14:30:00Z"
}
```

### 2. `delivery_locations/{assignment_id}`
Realtime active courier coordinates updated via debounced Go stream:
```json
{
  "assignment_id": "asg_demo_101",
  "employee_id": "6cce643f-0323-413e-939d-344c7ba27649",
  "latitude": 9.9252,
  "longitude": 78.1198,
  "speed": 35.5,
  "heading": 120.0,
  "timestamp": "2026-10-02T15:20:00Z"
}
```

### 3. `parcel_tracking/{parcel_id}`
Public/Customer-facing tracking checkpoint trail:
```json
{
  "parcel_id": "TRK-2026-0001",
  "tracking_number": "LF-98214301-IN",
  "current_status": "OUT_FOR_DELIVERY",
  "origin_hub": "Madurai Central Hub",
  "destination_hub": "Chennai South Hub",
  "last_checkpoint": "Out for delivery with courier Charlie Rider",
  "updated_at": "2026-10-02T15:22:00Z"
}
```

### 4. `device_tokens/{uid_platform}`
FCM push notification tokens mapped to users:
```json
{
  "user_id": "6cce643f-0323-413e-939d-344c7ba27649",
  "device_token": "fcm_token_string_here",
  "platform": "android",
  "active": true,
  "updated_at": "2026-10-02T15:00:00Z"
}
```

### 5. `system_events/{event_id}`
Audit events published upon assignment and custody shifts:
```json
{
  "event_type": "DELIVERY_ASSIGNMENT_CREATED",
  "assignment_id": "asg_demo_101",
  "parcel_id": "TRK-2026-0001",
  "employee_id": "6cce643f-0323-413e-939d-344c7ba27649",
  "created_at": "2026-10-02T15:10:00Z"
}
```
