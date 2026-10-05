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

## 3. Media Storage Boundaries (Cloudinary Architecture)

- **Cloudinary Mandate**: In LogiFlows, all customer Proof of Delivery (POD) photos, recipient signatures, and delivery documents are stored in Cloudinary.
- **No Binary Media in Firestore**: Cloud Firestore stores only structured metadata records containing the signed Cloudinary URL and public ID.
- **Firebase Storage Isolation**: If Firebase Storage is present in the Firebase project, its rules default to strict deny-all to ensure all file traffic is directed through Cloudinary.
