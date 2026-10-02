import logging
from datetime import datetime, timezone
from typing import Dict, Any, List, Optional
try:
    from app.core.firebase import get_firestore_db, get_storage_bucket, is_firebase_live
except ImportError:
    from ai.app.core.firebase import get_firestore_db, get_storage_bucket, is_firebase_live

logger = logging.getLogger("logiflows.ai.firebase_service")

# Simulated in-memory store for development/testing when live Firebase credentials are not attached
_mock_firestore_store: Dict[str, Dict[str, Any]] = {
    "users": {},
    "branches": {},
    "parcel_tracking": {},
    "delivery_locations": {},
    "notifications": {},
    "device_tokens": {},
    "system_events": {},
}

class FirebaseService:
    @staticmethod
    def is_available() -> bool:
        return is_firebase_live()

    # -------------------------------------------------------------------------
    # Firestore Document Operations
    # -------------------------------------------------------------------------
    @staticmethod
    def create_document(collection: str, doc_id: str, data: Dict[str, Any]) -> Dict[str, Any]:
        data["created_at"] = data.get("created_at", datetime.now(timezone.utc).isoformat())
        data["updated_at"] = datetime.now(timezone.utc).isoformat()

        if is_firebase_live():
            try:
                db = get_firestore_db()
                db.collection(collection).document(doc_id).set(data)
                return data
            except Exception as e:
                logger.error(f"Firestore create error on {collection}/{doc_id}: {e}")
                raise

        # Simulated fallback
        if collection not in _mock_firestore_store:
            _mock_firestore_store[collection] = {}
        _mock_firestore_store[collection][doc_id] = data
        return data

    @staticmethod
    def get_document(collection: str, doc_id: str) -> Optional[Dict[str, Any]]:
        if is_firebase_live():
            try:
                db = get_firestore_db()
                doc = db.collection(collection).document(doc_id).get()
                return doc.to_dict() if doc.exists else None
            except Exception as e:
                logger.error(f"Firestore get error on {collection}/{doc_id}: {e}")
                raise

        # Simulated fallback
        return _mock_firestore_store.get(collection, {}).get(doc_id)

    @staticmethod
    def update_document(collection: str, doc_id: str, data: Dict[str, Any]) -> Dict[str, Any]:
        data["updated_at"] = datetime.now(timezone.utc).isoformat()

        if is_firebase_live():
            try:
                db = get_firestore_db()
                db.collection(collection).document(doc_id).update(data)
                return data
            except Exception as e:
                logger.error(f"Firestore update error on {collection}/{doc_id}: {e}")
                raise

        # Simulated fallback
        existing = _mock_firestore_store.get(collection, {}).get(doc_id, {})
        existing.update(data)
        if collection not in _mock_firestore_store:
            _mock_firestore_store[collection] = {}
        _mock_firestore_store[collection][doc_id] = existing
        return existing

    @staticmethod
    def delete_document(collection: str, doc_id: str) -> bool:
        if is_firebase_live():
            try:
                db = get_firestore_db()
                db.collection(collection).document(doc_id).delete()
                return True
            except Exception as e:
                logger.error(f"Firestore delete error on {collection}/{doc_id}: {e}")
                raise

        # Simulated fallback
        if collection in _mock_firestore_store and doc_id in _mock_firestore_store[collection]:
            del _mock_firestore_store[collection][doc_id]
            return True
        return False

    @staticmethod
    def query_collection(
        collection: str,
        filters: Optional[List[tuple]] = None,
        limit: int = 50,
    ) -> List[Dict[str, Any]]:
        """Query Firestore with optional list of (field, operator, value) triples."""
        if is_firebase_live():
            try:
                db = get_firestore_db()
                query = db.collection(collection)
                if filters:
                    for field, op, val in filters:
                        query = query.where(field, op, val)
                docs = query.limit(limit).stream()
                return [d.to_dict() for d in docs]
            except Exception as e:
                logger.error(f"Firestore query error on {collection}: {e}")
                raise

        # Simulated fallback
        items = list(_mock_firestore_store.get(collection, {}).values())
        if filters:
            for field, op, val in filters:
                if op == "==":
                    items = [it for it in items if it.get(field) == val]
        return items[:limit]

    # -------------------------------------------------------------------------
    # Real-Time Telemetry & Tracking Mirror
    # -------------------------------------------------------------------------
    @classmethod
    def update_delivery_location(
        cls,
        assignment_id: str,
        employee_id: str,
        latitude: float,
        longitude: float,
        speed: float = 0.0,
        heading: float = 0.0,
    ) -> Dict[str, Any]:
        payload = {
            "assignment_id": assignment_id,
            "employee_id": employee_id,
            "latitude": latitude,
            "longitude": longitude,
            "speed": speed,
            "heading": heading,
            "timestamp": datetime.now(timezone.utc).isoformat(),
        }
        return cls.create_document("delivery_locations", str(assignment_id), payload)

    @classmethod
    def publish_system_event(
        cls,
        event_type: str,
        assignment_id: Optional[str] = None,
        parcel_id: Optional[str] = None,
        employee_id: Optional[str] = None,
        metadata: Optional[Dict[str, Any]] = None,
    ) -> Dict[str, Any]:
        event_id = f"evt_{datetime.now(timezone.utc).strftime('%Y%m%d%H%M%S')}_{parcel_id or 'sys'}"
        payload = {
            "event_type": event_type,
            "assignment_id": assignment_id,
            "parcel_id": parcel_id,
            "employee_id": employee_id,
            "metadata": metadata or {},
            "created_at": datetime.now(timezone.utc).isoformat(),
        }
        return cls.create_document("system_events", event_id, payload)

    # -------------------------------------------------------------------------
    # FCM Cloud Messaging
    # -------------------------------------------------------------------------
    @staticmethod
    def send_push_notification(
        token: str,
        title: str,
        body: str,
        data: Optional[Dict[str, str]] = None,
    ) -> str:
        if is_firebase_live():
            try:
                from firebase_admin import messaging
                msg = messaging.Message(
                    token=token,
                    notification=messaging.Notification(title=title, body=body),
                    data=data or {},
                )
                msg_id = messaging.send(msg)
                logger.info(f"FCM message delivered: {msg_id}")
                return msg_id
            except Exception as e:
                logger.error(f"FCM send error: {e}")
                raise

        # Simulated fallback
        sim_id = f"mock_fcm_{datetime.now(timezone.utc).timestamp()}"
        logger.info(f"Simulated FCM send to {token[:6]}...: {title}")
        return sim_id

    # -------------------------------------------------------------------------
    # Firebase Storage URL Generation
    # -------------------------------------------------------------------------
    @staticmethod
    def generate_signed_download_url(object_path: str, expiration_minutes: int = 60) -> str:
        if is_firebase_live():
            try:
                bucket = get_storage_bucket()
                blob = bucket.blob(object_path)
                from datetime import timedelta
                url = blob.generate_signed_url(
                    version="v4",
                    expiration=timedelta(minutes=expiration_minutes),
                    method="GET",
                )
                return url
            except Exception as e:
                logger.error(f"Storage signed URL error on {object_path}: {e}")
                raise

        return f"https://storage.googleapis.com/logiflows-platform.appspot.com/{object_path}?mock_token=valid"
