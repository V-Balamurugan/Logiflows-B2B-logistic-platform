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
    from app.services.cloudinary_service import CloudinaryService
except ImportError:
    from ai.app.services.cloudinary_service import CloudinaryService

__all__ = ["CloudinaryService"]
