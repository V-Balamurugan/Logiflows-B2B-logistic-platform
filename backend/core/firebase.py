import os
import sys

# Ensure ai directory is accessible in sys.path
_current_dir = os.path.dirname(os.path.abspath(__file__))
_project_root = os.path.dirname(_current_dir)
_ai_app_path = os.path.join(_project_root, "ai", "app")
_ai_root_path = os.path.join(_project_root, "ai")

for p in [_ai_app_path, _ai_root_path, _project_root]:
    if p not in sys.path:
        sys.path.insert(0, p)

try:
    from app.core.firebase import (
        get_firebase_app,
        get_firestore_db,
        is_firebase_live,
    )
    # Expose db client directly
    db = get_firestore_db()
except ImportError:
    try:
        from ai.app.core.firebase import (
            get_firebase_app,
            get_firestore_db,
            is_firebase_live,
        )
        db = get_firestore_db()
    except Exception as e:
        import firebase_admin
        from firebase_admin import credentials, firestore

        if not firebase_admin._apps:
            credential_path = os.getenv(
                "GOOGLE_APPLICATION_CREDENTIALS",
                "config/firebase-service-account.json"
            )
            if os.path.exists(credential_path):
                cred = credentials.Certificate(credential_path)
                firebase_admin.initialize_app(
                    cred,
                    {"projectId": os.getenv("FIREBASE_PROJECT_ID", "logiflows-platform")}
                )
        db = firestore.client() if firebase_admin._apps else None
