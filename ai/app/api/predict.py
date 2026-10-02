from typing import List, Optional
from pydantic import BaseModel, Field
from fastapi import APIRouter

router = APIRouter()

class DelayPredictionRequest(BaseModel):
    parcel_id: str = Field(..., description="Unique identifier of parcel")
    route_distance_km: float = Field(..., ge=0, description="Estimated total route distance in km")
    estimated_duration_min: float = Field(..., ge=0, description="Estimated standard duration in minutes")
    stops_count: int = Field(..., ge=0, description="Number of delivery stops on this run")
    time_of_day: int = Field(..., ge=0, le=23, description="Hour of departure (0-23)")
    day_of_week: int = Field(..., ge=0, le=6, description="Day of week (0=Sunday, 6=Saturday)")
    employee_active_deliveries: int = Field(0, ge=0, description="Current courier workload")

class DelayPredictionResponse(BaseModel):
    parcel_id: str
    risk_level: str = Field(..., description="Risk category: LOW, MEDIUM, HIGH")
    delay_probability: float = Field(..., ge=0.0, le=1.0, description="Predicted probability of delay")
    confidence_score: float = Field(..., ge=0.0, le=1.0, description="Model confidence score")
    predicted_delay_minutes: float = Field(..., ge=0.0, description="Estimated delay in minutes")
    contributing_factors: List[str]
    model_version: str = "0.1.0-baseline"

@router.post("/delay-risk", response_model=DelayPredictionResponse)
def predict_delay_risk(payload: DelayPredictionRequest):
    # Baseline heuristic rule engine before training phase
    base_prob = 0.05
    factors = []

    if payload.route_distance_km > 30.0:
        base_prob += 0.25
        factors.append("HIGH_DISTANCE")
    if payload.stops_count > 15:
        base_prob += 0.30
        factors.append("HIGH_STOP_DENSITY")
    if payload.time_of_day in [8, 9, 17, 18, 19]:
        base_prob += 0.20
        factors.append("PEAK_TRAFFIC_HOURS")
    if payload.employee_active_deliveries > 20:
        base_prob += 0.20
        factors.append("COURIER_OVERLOAD")

    risk_level = "LOW"
    if base_prob >= 0.60:
        risk_level = "HIGH"
    elif base_prob >= 0.30:
        risk_level = "MEDIUM"

    est_delay = base_prob * 45.0 # Max expected delay window for calculation

    return DelayPredictionResponse(
        parcel_id=payload.parcel_id,
        risk_level=risk_level,
        delay_probability=min(base_prob, 0.99),
        confidence_score=0.88,
        predicted_delay_minutes=round(est_delay, 1),
        contributing_factors=factors if factors else ["NORMAL_OPERATING_CONDITIONS"],
        model_version="0.1.0-baseline"
    )
