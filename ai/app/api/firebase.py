from typing import Dict, Any, List, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from pydantic import BaseModel, Field
try:
    from app.services.firebase_service import FirebaseService
    from app.core.firebase_auth import get_current_firebase_user, require_role
except ImportError:
    from ai.app.services.firebase_service import FirebaseService
    from ai.app.core.firebase_auth import get_current_firebase_user, require_role

router = APIRouter(prefix="/firebase", tags=["Firebase Services"])

class DeviceTokenPayload(BaseModel):
    device_token: str = Field(..., description="FCM device registration token")
    platform: str = Field(default="android", description="Platform: android, ios, web")

class DeliveryLocationPayload(BaseModel):
    assignment_id: str
    employee_id: str
    latitude: float
    longitude: float
    speed: float = 0.0
    heading: float = 0.0

@router.get("/status")
def get_firebase_status():
    return {
        "status": "OK",
        "live": FirebaseService.is_available(),
        "mode": "LIVE_CLOUD" if FirebaseService.is_available() else "SIMULATED_MOCK",
        "collections": [
            "users",
            "branches",
            "parcel_tracking",
            "delivery_locations",
            "notifications",
            "device_tokens",
            "system_events",
        ],
    }

@router.post("/device-tokens")
def register_device_token(
    payload: DeviceTokenPayload,
    user: Dict[str, Any] = Depends(get_current_firebase_user),
):
    uid = user.get("uid", "anonymous")
    data = {
        "user_id": uid,
        "email": user.get("email"),
        "role": user.get("role"),
        "device_token": payload.device_token,
        "platform": payload.platform,
        "active": True,
    }
    FirebaseService.create_document("device_tokens", f"{uid}_{payload.platform}", data)
    return {"status": "SUCCESS", "message": "Device token registered"}

@router.get("/notifications")
def get_user_notifications(
    user: Dict[str, Any] = Depends(get_current_firebase_user),
    limit: int = Query(default=20, le=100),
):
    uid = user.get("uid")
    notifications = FirebaseService.query_collection(
        "notifications",
        filters=[("user_id", "==", uid)],
        limit=limit,
    )
    return {"data": notifications, "count": len(notifications)}

@router.post("/deliveries/location")
def update_location(
    payload: DeliveryLocationPayload,
    user: Dict[str, Any] = Depends(require_role("EMPLOYEE")),
):
    record = FirebaseService.update_delivery_location(
        assignment_id=payload.assignment_id,
        employee_id=payload.employee_id,
        latitude=payload.latitude,
        longitude=payload.longitude,
        speed=payload.speed,
        heading=payload.heading,
    )
    return {"status": "UPDATED", "data": record}

@router.get("/deliveries/{assignment_id}/location")
def get_delivery_location(
    assignment_id: str,
    user: Dict[str, Any] = Depends(get_current_firebase_user),
):
    loc = FirebaseService.get_document("delivery_locations", assignment_id)
    if not loc:
        raise HTTPException(status_code=404, detail="Live location not found for assignment")
    return {"status": "OK", "data": loc}

@router.get("/parcels/{parcel_id}/tracking")
def get_parcel_tracking(
    parcel_id: str,
    user: Dict[str, Any] = Depends(get_current_firebase_user),
):
    tracking = FirebaseService.get_document("parcel_tracking", parcel_id)
    if not tracking:
        # Return initialized mirror structure
        return {
            "status": "INITIALIZED",
            "data": {
                "parcel_id": parcel_id,
                "current_status": "BOOKED",
                "events": [],
            },
        }
    return {"status": "OK", "data": tracking}
