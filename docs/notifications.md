# LogiFlows — Notification System & FCM Architecture

## 1. Multi-Channel Notification Pipeline

LogiFlows implements a resilient multi-tier notification architecture combining Firebase Cloud Messaging (FCM), in-app Firestore notification inboxes, and external webhooks (n8n / Brevo):

```mermaid
flowchart TD
    Trigger[Business Event: Assignment Created / Status Changed / Delay Predicted] --> Backend[FastAPI / Go Core]
    Backend -->|1. In-App Inbox| Firestore[Firestore notifications/{id}]
    Backend -->|2. Mobile Push| FCM[Firebase Cloud Messaging Admin SDK]
    Backend -->|3. Webhook / Email| n8nBrevo[n8n Workflow & Brevo Email Service]

    Firestore -->|Live Snapshot| WebPortal[React Web Dashboard]
    FCM -->|Push Notification| CourierDevice[Courier Flutter App]
    FCM -->|Push Notification| CustomerDevice[Customer Mobile / Web]
```

---

## 2. Supported Notification Types

1. **`DELIVERY_ASSIGNMENT`**: Dispatched to couriers when a parcel assignment is created.
2. **`PARCEL_STATUS`**: Updates to customers (`BOOKED`, `PICKED_UP`, `IN_TRANSIT`, `OUT_FOR_DELIVERY`, `DELIVERED`).
3. **`DELAY_ALERT`**: Dispatched when the Python AI service predicts medium or high risk of transit delay.
4. **`DELIVERY_STARTED`**: Dispatched when courier initiates delivery route.
5. **`DELIVERY_COMPLETED`**: Confirmation containing POD verification.
6. **`SYSTEM_ALERT`**: Maintenance notices or emergency branch rerouting.

---

## 3. Device Token Management

- Courier and customer apps register their push tokens on launch via `POST /api/v1/firebase/device-tokens`.
- Stored in Firestore under `device_tokens/{user_id}_{platform}`:
```json
{
  "user_id": "6cce643f-0323-413e-939d-344c7ba27649",
  "device_token": "fcm_token_string_here",
  "platform": "android",
  "active": true,
  "updated_at": "2026-10-02T15:00:00Z"
}
```

---

## 4. Multi-Tenant Notification Scoping

- In accordance with LogiFlows security rules, customers only receive notifications for parcels they own.
- Couriers only receive assignments and alerts within their assigned tenant organization.
- Device tokens are strictly protected by Firestore security rules so that tokens cannot be read or hijacked across user boundaries.
