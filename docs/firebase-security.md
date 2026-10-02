# Firebase Security & Secret Sanitization

## 1. Zero-Leak Credential Policy

- Private keys, client email addresses, and service account JSON files are **NEVER** exposed to frontend web code or mobile client bundles.
- The `.gitignore` file enforces exclusion for:
  ```gitignore
  .env
  *.env
  *.pem
  *.key
  service-account*.json
  serviceAccount*.json
  *serviceAccountKey*.json
  firebase-service-account.json
  *-firebase-adminsdk-*.json
  firebase-adminsdk*.json
  ```
- Git history audits are conducted before every release to guarantee zero secret leakage.

---

## 2. Firestore Rule Boundaries

- Default rule: **Deny All** (`allow read, write: if false;`).
- Roles are validated through custom claims or authoritative database mappings (`ADMIN`, `EMPLOYEE`, `CUSTOMER`).
- Customers can only read parcel tracking records related to their deliveries.
- Employees can only write GPS coordinates to their own assigned deliveries.
- Admins maintain role-scoped operational observability.

---

## 3. Storage Rule Boundaries

- Default rule: **Deny All**.
- Proof of Delivery (`proof-of-delivery/{parcelId}/{fileName}`): Restricted to authenticated couriers and admins, maximum file size 10MB, strictly validated MIME types (`image/*`, `application/pdf`).
- Unauthenticated or public write access is strictly prohibited across all bucket paths.
