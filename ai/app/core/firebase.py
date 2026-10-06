import os
import json
import logging
from typing import Optional, Any
import firebase_admin
from firebase_admin import credentials, firestore, storage, messaging, auth

logger = logging.getLogger("logiflows.ai.firebase")

_app: Optional[firebase_admin.App] = None
_db: Optional[Any] = None
_bucket: Optional[Any] = None
_is_initialized: bool = False
_is_mock: bool = True

def get_firebase_app() -> Optional[firebase_admin.App]:
    global _app, _db, _bucket, _is_initialized, _is_mock

    if _is_initialized:
        return _app

    project_id = os.getenv("FIREBASE_PROJECT_ID", "logiflows-platform")
    storage_bucket = os.getenv("FIREBASE_STORAGE_BUCKET", f"{project_id}.appspot.com")

    # 1. Check GOOGLE_APPLICATION_CREDENTIALS or local service account file
    cred_path = os.getenv("GOOGLE_APPLICATION_CREDENTIALS", "")
    possible_paths = [
        cred_path,
        "config/firebase-service-account.json",
        "serviceAccountKey.json",
        "../serviceAccountKey.json",
        "../backend/serviceAccountKey.json",
    ]

    selected_cred = None
    for p in possible_paths:
        if p and os.path.exists(p):
            try:
                selected_cred = credentials.Certificate(p)
                logger.info(f"Loaded Firebase service account credentials from {p}")
                break
            except Exception as e:
                logger.warning(f"Failed to load credentials from {p}: {e}")

    # 2. Check direct JSON string
    if not selected_cred and os.getenv("FIREBASE_CREDENTIALS_JSON"):
        try:
            cred_dict = json.loads(os.getenv("FIREBASE_CREDENTIALS_JSON", "{}"))
            selected_cred = credentials.Certificate(cred_dict)
            logger.info("Loaded Firebase credentials from FIREBASE_CREDENTIALS_JSON")
        except Exception as e:
            logger.warning(f"Failed to parse FIREBASE_CREDENTIALS_JSON: {e}")

    # 3. Check individual parameters
    client_email = os.getenv("FIREBASE_CLIENT_EMAIL")
    private_key = os.getenv("FIREBASE_PRIVATE_KEY")
    if not selected_cred and client_email and private_key and "-----BEGIN PRIVATE KEY-----" in private_key:
        try:
            formatted_key = private_key.replace("\\n", "\n")
            cred_dict = {
                "type": "service_account",
                "project_id": project_id,
                "client_email": client_email,
                "private_key": formatted_key,
                "token_uri": "https://oauth2.googleapis.com/token",
            }
            selected_cred = credentials.Certificate(cred_dict)
            logger.info(f"Constructed Firebase credentials for project {project_id}")
        except Exception as e:
            logger.warning(f"Failed to parse individual Firebase environment variables: {e}")

    if selected_cred:
        try:
            if not firebase_admin._apps:
                _app = firebase_admin.initialize_app(
                    selected_cred,
                    {"storageBucket": storage_bucket, "projectId": project_id},
                )
            else:
                _app = firebase_admin.get_app()

            _db = firestore.client(app=_app)
            _bucket = storage.bucket(app=_app)
            _is_mock = False
            logger.info(f"Firebase Admin SDK initialized successfully for {project_id}")
        except Exception as e:
            logger.error(f"Failed to initialize live Firebase Admin SDK: {e}")
            _is_mock = True
    else:
        logger.warning(
            "Firebase credentials not found. Initialized in safe simulated mock mode."
        )
        _is_mock = True

    _is_initialized = True
    return _app

def get_firestore_db() -> Optional[Any]:
    get_firebase_app()
    return _db

def get_storage_bucket() -> Optional[Any]:
    get_firebase_app()
    return _bucket

def is_firebase_live() -> bool:
    get_firebase_app()
    return not _is_mock
