# LogiFlows — Backend Architecture Specification

## 1. Structure

The authoritative backend follows clean architecture with strict domain boundaries:

```text
backend/
├── cmd/
│   ├── api/        # Application HTTP daemon entrypoint
│   └── migrate/    # Standalone migration CLI runner
├── internal/
│   ├── config/     # Environment parsing & validation
│   ├── database/   # PostgreSQL & Redis connection pools
│   ├── httpapi/    # Routers, handlers, standard response envelopes
│   ├── logging/    # Context-aware structured slog logger
│   ├── middleware/ # Request ID, Recovery, Logging, CORS, Auth
│   └── migrations/ # Embedded SQL migrations
```

## 2. Standard Response Format (`/api/v1`)

### Success Envelope
```json
{
  "data": { ... },
  "meta": { "total": 100, "page": 1 },
  "request_id": "c138f5f3-5bf9-450f-a393-01d7ee4d8c83"
}
```

### Error Envelope
```json
{
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "The requested resource does not exist.",
    "request_id": "c138f5f3-5bf9-450f-a393-01d7ee4d8c83",
    "details": {}
  }
}
```
