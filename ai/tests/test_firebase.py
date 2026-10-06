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

def test_firebase_health_endpoint():
    # Test both /api/firebase/health and /api/v1/firebase/health
    for path in ["/api/firebase/health", "/api/v1/firebase/health"]:
        res = client.get(path)
        assert res.status_code == 200
        data = res.json()
        assert data["firebase"] == "connected"
        assert data["firestore"] == "connected"
        assert "project_id" in data

def test_firestore_test_crud_endpoints():
    # 1. CREATE via POST /api/firebase/test
    create_payload = {"id": "test_endpoint_doc_1", "message": "Testing firestore endpoint", "env": "test"}
    res = client.post("/api/firebase/test", json=create_payload)
    assert res.status_code == 200
    assert res.json()["status"] == "SUCCESS"
    assert res.json()["document_id"] == "test_endpoint_doc_1"

    # 2. READ via GET /api/firebase/test/{document_id}
    res = client.get("/api/firebase/test/test_endpoint_doc_1")
    assert res.status_code == 200
    assert res.json()["status"] == "SUCCESS"
    assert res.json()["data"]["message"] == "Testing firestore endpoint"

    # 3. UPDATE via PUT /api/firebase/test/{document_id}
    res = client.put("/api/firebase/test/test_endpoint_doc_1", json={"message": "Updated firestore endpoint message"})
    assert res.status_code == 200
    assert res.json()["status"] == "SUCCESS"
    assert res.json()["data"]["message"] == "Updated firestore endpoint message"

    # 4. DELETE via DELETE /api/firebase/test/{document_id}
    res = client.delete("/api/firebase/test/test_endpoint_doc_1")
    assert res.status_code == 200
    assert "deleted" in res.json()["message"]

    # Verify 404 after deletion
    res = client.get("/api/firebase/test/test_endpoint_doc_1")
    assert res.status_code == 404

def test_specialized_firebase_services():
    # Parcel realtime state
    parcel = FirebaseService.update_parcel_realtime_state(
        parcel_id="TRK-9901",
        status="OUT_FOR_DELIVERY",
        latitude=12.9716,
        longitude=77.5946,
        employee_id="emp_99",
        assignment_id="asg_88",
        eta="25 mins",
        checkpoint_description="Out for delivery in Indiranagar",
    )
    assert parcel["parcel_id"] == "TRK-9901"
    assert parcel["status"] == "OUT_FOR_DELIVERY"
    assert parcel["latitude"] == 12.9716

    # Notification
    notif = FirebaseService.create_notification(
        user_id="user_test_1",
        title="Parcel Update",
        message="Your parcel is out for delivery",
        notification_type="PARCEL_STATUS",
    )
    assert notif["user_id"] == "user_test_1"
    assert notif["type"] == "PARCEL_STATUS"
    assert notif["read"] is False

    # Device token
    dev = FirebaseService.register_device_token(
        user_id="user_test_1",
        token="mock_fcm_token_xyz_123",
        platform="android",
    )
    assert dev["token"] == "mock_fcm_token_xyz_123"
    assert dev["active"] is True

    # System event
    evt = FirebaseService.create_system_event(
        event_type="DELIVERY_ASSIGNMENT_CREATED",
        parcel_id="TRK-9901",
        assignment_id="asg_88",
        employee_id="emp_99",
    )
    assert evt["event_type"] == "DELIVERY_ASSIGNMENT_CREATED"
    assert evt["parcel_id"] == "TRK-9901"

def test_cloudinary_service_integration():
    try:
        from app.services.cloudinary_service import CloudinaryService
    except ImportError:
        from ai.app.services.cloudinary_service import CloudinaryService

    record = CloudinaryService.record_media_upload(
        public_id="pod_photo_trk9901_sig",
        cloudinary_url="https://res.cloudinary.com/logiflows-cloud/image/upload/v1/pod/sig.png",
        file_type="RECIPIENT_SIGNATURE",
        uploaded_by="courier_charlie",
        parcel_id="TRK-9901",
    )
    assert record["public_id"] == "pod_photo_trk9901_sig"
    assert "cloudinary.com" in record["cloudinary_url"]
    assert record["file_type"] == "RECIPIENT_SIGNATURE"

    params = CloudinaryService.generate_signed_upload_params(folder="logiflows/signatures")
    assert params["folder"] == "logiflows/signatures"
    assert "timestamp" in params

def test_customer_forbidden_from_courier_endpoints():
    # Customer token attempting to update courier delivery location -> 403 Forbidden
    res = client.post(
        "/api/v1/firebase/deliveries/location",
        headers={"Authorization": "Bearer mock_token_customer_test"},
        json={
            "assignment_id": "asg_123",
            "employee_id": "emp_1",
            "latitude": 9.58,
            "longitude": 78.08,
        },
    )
    assert res.status_code == 403

