# ADR-002: Go Core Backend & Python AI Service Boundary

## Status
Accepted

## Context
LogiFlows requires ultra-reliable, high-throughput backend services for API orchestration, JWT authentication, RBAC authorization, and state machines, as well as machine learning models for predictive delay-risk scoring.

## Decision
- **Go (Gin/Standard Library)** owns all business domain state, database persistence, and authorization enforcement.
- **Python (FastAPI)** acts strictly as an advisory AI microservice for delay-risk scoring and feature extraction.
- Python never writes directly to core PostgreSQL business tables.
- Go communicates with Python over internal HTTP (`/api/v1/predict/delay-risk`) with graceful fallback when AI services are unreachable or degraded.

## Consequences
- Business logic is unified and never duplicated across Go and Python.
- Failure of AI prediction services degrades gracefully without impacting core courier booking, scanning, or delivery workflows.
