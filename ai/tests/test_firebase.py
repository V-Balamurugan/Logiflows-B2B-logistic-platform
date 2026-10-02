import pytest
from fastapi.testclient import TestClient
try:
    from app.main import app
    from app.services.firebase_service import FirebaseService
    from app.core.firebase_auth import verify_firebase_id_token
except ImportError:
    from ai.app.main import app
    from ai.app.services.firebase_service import FirebaseService
    from ai.app.core.firebase_auth import verify_firebase_id_token

client = TestClient(app)

def test_firebase_status_endpoint():
    response = client.get("/api/v1/firebase/status")
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "OK"
    assert "collections" in data
    assert "delivery_locations" in data["collections"]
    assert "users" in data["collections"]

def test_firestore_crud_operations():
    test_id = "test_doc_101"
    payload = {"name": "Central Hub", "code": "HUB-01", "city": "Madurai"}

    # Create
    created = FirebaseService.create_document("branches", test_id, payload)
    assert created["name"] == "Central Hub"
    assert "created_at" in created

    # Get
    retrieved = FirebaseService.get_document("branches", test_id)
    assert retrieved is not None
    assert retrieved["code"] == "HUB-01"

    # Update
    updated = FirebaseService.update_document("branches", test_id, {"city": "Chennai"})
    assert updated["city"] == "Chennai"

    # Query
    results = FirebaseService.query_collection("branches", filters=[("code", "==", "HUB-01")])
    assert len(results) >= 1
    assert results[0]["name"] == "Central Hub"

    # Delete
    deleted = FirebaseService.delete_document("branches", test_id)
    assert deleted is True
    assert FirebaseService.get_document("branches", test_id) is None

def test_delivery_location_publishing():
    loc = FirebaseService.update_delivery_location(
        assignment_id="asg_999",
        employee_id="emp_42",
        latitude=9.9252,
        longitude=78.1198,
        speed=45.0,
        heading=180.0,
    )
    assert loc["assignment_id"] == "asg_999"
    assert loc["latitude"] == 9.9252
    assert "timestamp" in loc

    # Verify retrieval
    fetched = FirebaseService.get_document("delivery_locations", "asg_999")
    assert fetched is not None
    assert fetched["speed"] == 45.0

def test_storage_signed_url_generation():
    url = FirebaseService.generate_signed_download_url("proof-of-delivery/parcel_55/signature.png")
    assert url.startswith("https://")
    assert "proof-of-delivery" in url

def test_firebase_auth_mock_verification():
    claims = verify_firebase_id_token("mock_token_employee_123")
    assert claims["role"] == "EMPLOYEE"
    assert claims["email"] == "employee@logiflows.io"

def test_protected_location_update_endpoint():
    # Attempt without token -> 401
    res = client.post(
        "/api/v1/firebase/deliveries/location",
        json={
            "assignment_id": "asg_123",
            "employee_id": "emp_1",
            "latitude": 9.58,
            "longitude": 78.08,
        },
    )
    assert res.status_code == 401

    # Attempt with valid employee token -> 200
    res = client.post(
        "/api/v1/firebase/deliveries/location",
        headers={"Authorization": "Bearer mock_token_employee_test"},
        json={
            "assignment_id": "asg_123",
            "employee_id": "emp_1",
            "latitude": 9.58,
            "longitude": 78.08,
            "speed": 30.0,
            "heading": 90.0,
        },
    )
    assert res.status_code == 200
    assert res.json()["status"] == "UPDATED"
