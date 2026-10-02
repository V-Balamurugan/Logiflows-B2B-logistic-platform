import logging
from typing import Dict, Any, Optional
from fastapi import HTTPException, Security, Depends
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
try:
    from app.core.firebase import is_firebase_live, get_firebase_app
except ImportError:
    from ai.app.core.firebase import is_firebase_live, get_firebase_app

logger = logging.getLogger("logiflows.ai.firebase_auth")
security = HTTPBearer(auto_error=False)

VALID_ROLES = {"ADMIN", "EMPLOYEE", "CUSTOMER", "PLATFORM_ADMIN", "TENANT_ADMIN"}

def verify_firebase_id_token(id_token: str) -> Dict[str, Any]:
    """Verify incoming Firebase JWT ID token and return decoded claims."""
    if is_firebase_live():
        try:
            from firebase_admin import auth
            decoded_token = auth.verify_id_token(id_token)
            return decoded_token
        except Exception as e:
            logger.warning(f"Firebase token verification failed: {e}")
            raise HTTPException(
                status_code=401,
                detail=f"Invalid or expired Firebase authentication token: {str(e)}",
            )

    # Simulated fallback for development testing
    if id_token.startswith("mock_token_"):
        parts = id_token.split("_")
        role = parts[2].upper() if len(parts) > 2 else "CUSTOMER"
        return {
            "uid": f"mock_uid_{role.lower()}",
            "email": f"{role.lower()}@logiflows.io",
            "role": role if role in VALID_ROLES else "CUSTOMER",
            "auth_time": 1700000000,
        }

    raise HTTPException(status_code=401, detail="Authentication token required")

def get_current_firebase_user(
    credentials: Optional[HTTPAuthorizationCredentials] = Security(security),
) -> Dict[str, Any]:
    if not credentials or not credentials.credentials:
        raise HTTPException(
            status_code=401,
            detail="Missing Authorization Bearer token",
            headers={"WWW-Authenticate": "Bearer"},
        )
    return verify_firebase_id_token(credentials.credentials)

def require_role(required_role: str):
    def role_checker(user: Dict[str, Any] = Depends(get_current_firebase_user)) -> Dict[str, Any]:
        role = user.get("role", "CUSTOMER")
        if role != required_role and role != "ADMIN" and role != "PLATFORM_ADMIN":
            raise HTTPException(
                status_code=403,
                detail=f"Forbidden: role '{required_role}' required",
            )
        return user
    return role_checker
