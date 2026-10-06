import os
import sys

_current_dir = os.path.dirname(os.path.abspath(__file__))
_project_root = os.path.dirname(_current_dir)
_ai_app_path = os.path.join(_project_root, "ai", "app")
_ai_root_path = os.path.join(_project_root, "ai")

for p in [_ai_app_path, _ai_root_path, _project_root]:
    if p not in sys.path:
        sys.path.insert(0, p)

try:
    from app.services.firebase_service import FirebaseService
except ImportError:
    from ai.app.services.firebase_service import FirebaseService

# Reusable function exports as requested in Section 8
create_document = FirebaseService.create_document
get_document = FirebaseService.get_document
update_document = FirebaseService.update_document
delete_document = FirebaseService.delete_document
query_collection = FirebaseService.query_collection
set_document = FirebaseService.set_document
add_document = FirebaseService.add_document

update_parcel_realtime_state = FirebaseService.update_parcel_realtime_state
update_delivery_location = FirebaseService.update_delivery_location
create_notification = FirebaseService.create_notification
register_device_token = FirebaseService.register_device_token
create_system_event = FirebaseService.create_system_event

__all__ = [
    "FirebaseService",
    "create_document",
    "get_document",
    "update_document",
    "delete_document",
    "query_collection",
    "set_document",
    "add_document",
    "update_parcel_realtime_state",
    "update_delivery_location",
    "create_notification",
    "register_device_token",
    "create_system_event",
]
