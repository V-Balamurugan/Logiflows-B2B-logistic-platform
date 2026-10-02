from fastapi.testclient import TestClient
from app.main import app

client = TestClient(app)

def test_healthz():
    response = client.get("/healthz")
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "OK"
    assert data["service"] == "logiflows-ai"

def test_readyz():
    response = client.get("/readyz")
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "READY"

def test_predict_delay_risk_low():
    payload = {
        "parcel_id": "test-parcel-001",
        "route_distance_km": 5.0,
        "estimated_duration_min": 15.0,
        "stops_count": 3,
        "time_of_day": 14,
        "day_of_week": 2,
        "employee_active_deliveries": 4
    }
    response = client.post("/api/v1/predict/delay-risk", json=payload)
    assert response.status_code == 200
    data = response.json()
    assert data["risk_level"] == "LOW"
    assert data["parcel_id"] == "test-parcel-001"

def test_predict_delay_risk_high():
    payload = {
        "parcel_id": "test-parcel-002",
        "route_distance_km": 45.0,
        "estimated_duration_min": 80.0,
        "stops_count": 22,
        "time_of_day": 18,
        "day_of_week": 5,
        "employee_active_deliveries": 25
    }
    response = client.post("/api/v1/predict/delay-risk", json=payload)
    assert response.status_code == 200
    data = response.json()
    assert data["risk_level"] == "HIGH"
    assert "PEAK_TRAFFIC_HOURS" in data["contributing_factors"]
