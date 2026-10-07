# Firebase Setup & Deployment Guide

## 1. Prerequisites
- Firebase Project created on Google Cloud: `logiflows-platform`
- Downloaded service account JSON (`serviceAccountKey.json`)
- Enabled Firebase Services:
  - Cloud Firestore
  - Firebase Cloud Messaging (FCM)
  - Firebase Storage
  - Firebase Authentication

---

## 2. Local Development Credential Setup

You have two choices to connect your downloaded service account key:

### Method A: Direct File Placement (Recommended)
Place your downloaded key at:
```bash
backend/serviceAccountKey.json
```
or
```bash
ai/config/firebase-service-account.json
```
Both paths are strictly excluded in `.gitignore` to prevent leaks.

### Method B: Environment Variables
Populate [`.env`](file:///c:/Users/PUTTU/Documents/Logiflows-B2B%20bussiness%20platform/.env):
```env
FIREBASE_PROJECT_ID=logiflows-platform
FIREBASE_CLIENT_EMAIL=firebase-adminsdk@logiflows-platform.iam.gserviceaccount.com
FIREBASE_PRIVATE_KEY="-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n"
FIREBASE_STORAGE_BUCKET=logiflows-platform.appspot.com
```

---

## 3. Deploying Firestore & Storage Security Rules

Using Firebase CLI:
```bash
npm install -g firebase-tools
firebase login
firebase use logiflows-platform
firebase deploy --only firestore:rules,storage
```

---

## 4. Verification Endpoint

Verify operational status at any time:
```bash
# Go Authoritative Backend
curl http://localhost:8080/api/v1/firebase/status

# Python AI & Automation Service
curl http://localhost:8000/api/v1/firebase/status
```
When valid credentials are present, `mode` transitions automatically from `"SIMULATED_MOCK"` to `"LIVE_CLOUD"`.
