# CommunityShield Mobile API

## Base URL
https://communityshield-backend.onrender.com/api/v1

## Auth
Add this header to all requests:
Authorization: Bearer YOUR_JWT_TOKEN

## Endpoints

### GET /mobile/config
Returns app configuration
Response: { "config": { "appName": "CommunityShield", "version": "1.0.0" } }

### GET /mobile/dashboard
Returns user dashboard data
Response: { "totalCases": 0, "pendingCases": 0, "user": {...} }

### GET /mobile/notifications
Returns user notifications
Response: { "notifications": [], "unreadCount": 0 }

### GET /mobile/sync?lastSync=2024-01-01T00:00:00Z
Syncs data for offline use
Response: { "cases": [], "units": [], "syncTime": "..." }

### POST /mobile/crash-report
Reports app crash
Body: { "message": "error", "device": "iPhone", "os": "iOS" }

### POST /auth/login
Login user
Body: { "email": "user@example.com", "password": "password" }

### POST /auth/register
Register user
Body: { "email": "user@example.com", "firstName": "John", "lastName": "Doe", "password": "password", "role": "citizen" }

### POST /cases
Report case
Body: { "title": "Suspicious", "description": "...", "latitude": 6.5244, "longitude": 3.3792 }

### POST /sos/send
Send SOS
Body: { "latitude": 6.5244, "longitude": 3.3792, "description": "Emergency!" }