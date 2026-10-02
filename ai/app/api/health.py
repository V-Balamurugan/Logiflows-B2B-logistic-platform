import time
from datetime import datetime, timezone
from fastapi import APIRouter

router = APIRouter()
START_TIME = time.time()

@router.get("/healthz")
def healthz():
    return {
        "status": "OK",
        "service": "logiflows-ai",
        "uptime_seconds": int(time.time() - START_TIME),
        "timestamp": datetime.now(timezone.utc).isoformat()
    }

@router.get("/readyz")
def readyz():
    return {
        "status": "READY",
        "dependencies": {
            "model_engine": "INITIALIZED",
            "feature_store": "READY"
        },
        "timestamp": datetime.now(timezone.utc).isoformat()
    }
