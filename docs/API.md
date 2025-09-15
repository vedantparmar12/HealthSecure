# HealthSecure API Documentation

## Overview

The HealthSecure API provides secure, HIPAA-compliant access to medical data with role-based permissions and comprehensive audit logging.

## Base URL

```
http://localhost:8080/api
```

## Authentication

All API endpoints (except login and OAuth) require JWT authentication via the Authorization header:

```
Authorization: Bearer <jwt_token>
```

## User Roles

- **Admin**: User management, system configuration, audit logs
- **Doctor**: Full patient data access, create/update medical records
- **Nurse**: Limited patient data access, update care information
- **System**: Internal system operations

## Endpoints

### Authentication

#### POST /api/auth/login
Login with email and password.

**Request:**
```json
{
  "email": "doctor@hospital.com",
  "password": "securepassword"
}
```

**Response:**
```json
{
  "message": "Login successful",
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_at": "2024-01-01T12:00:00Z",
  "user": {
    "id": 1,
    "email": "doctor@hospital.com",
    "name": "Dr. Smith",
    "role": "doctor",
    "active": true
  }
}
```

#### POST /api/auth/refresh
Refresh access token using refresh token.

#### POST /api/auth/logout
Logout and invalidate tokens.

#### GET /api/auth/me
Get current user information.

### Patients

#### GET /api/patients
List patients with filtering and pagination.

**Query Parameters:**
- `page`: Page number (default: 1)
- `limit`: Items per page (default: 20, max: 50)
- `first_name`: Filter by first name
- `last_name`: Filter by last name
- `ssn`: Filter by SSN (doctors only)
- `phone`: Filter by phone number

**Response:**
```json
{
  "patients": [
    {
      "id": 1,
      "first_name": "John",
      "last_name": "Doe",
      "date_of_birth": "1985-03-15T00:00:00Z",
      "phone": "+1-555-0123",
      "address": "123 Main St, Anytown, ST 12345",
      "emergency_contact": "Jane Doe (Wife) - +1-555-0124",
      "created_at": "2024-01-01T12:00:00Z"
    }
  ],
  "pagination": {
    "current_page": 1,
    "limit": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

#### POST /api/patients
Create a new patient (doctors and admins only).

**Request:**
```json
{
  "first_name": "Jane",
  "last_name": "Smith",
  "date_of_birth": "1990-05-20T00:00:00Z",
  "ssn": "123-45-6789",
  "phone": "+1-555-0123",
  "address": "456 Oak Ave, Somewhere, ST 23456",
  "emergency_contact": "John Smith (Husband) - +1-555-0124"
}
```

#### GET /api/patients/:id
Get patient by ID.

**Headers:**
- `X-Emergency-Access-Token`: Optional emergency access token

**Response:**
```json
{
  "patient": {
    "id": 1,
    "first_name": "John",
    "last_name": "Doe",
    "date_of_birth": "1985-03-15T00:00:00Z",
    "ssn": "123-45-6789",
    "phone": "+1-555-0123",
    "address": "123 Main St, Anytown, ST 12345",
    "emergency_contact": "Jane Doe (Wife) - +1-555-0124"
  }
}
```

#### PUT /api/patients/:id
Update patient information.

#### DELETE /api/patients/:id
Delete patient (admin only).

### Medical Records

#### GET /api/patients/:id/records
Get medical records for a patient.

#### POST /api/patients/:id/records
Create medical record (doctors only).

**Request:**
```json
{
  "diagnosis": "Hypertension",
  "treatment": "ACE inhibitor therapy",
  "notes": "Patient responding well to treatment",
  "medications": "Lisinopril 10mg daily",
  "severity": "medium"
}
```

#### GET /api/records/:id
Get specific medical record.

#### PUT /api/records/:id
Update medical record (doctors only).

### Emergency Access

#### POST /api/emergency/request
Request emergency access to patient data.

**Request:**
```json
{
  "patient_id": 1,
  "reason": "Patient brought to ER unconscious after accident. Need immediate access to medical history."
}
```

**Response:**
```json
{
  "id": 1,
  "access_token": "emergency_abc123...",
  "expires_at": "2024-01-01T13:00:00Z",
  "status": "active"
}
```

#### POST /api/emergency/activate/:id
Activate emergency access.

#### POST /api/emergency/revoke/:id
Revoke emergency access.

#### GET /api/emergency/active
Get active emergency access sessions (admin only).

### Audit Logs

#### GET /api/audit/logs
Get audit logs with filtering.

**Query Parameters:**
- `user_id`: Filter by user ID
- `patient_id`: Filter by patient ID
- `action`: Filter by action type
- `success`: Filter by success status
- `emergency`: Filter emergency access events
- `start_time`: Start date filter
- `end_time`: End date filter

#### GET /api/audit/users/:id
Get audit history for specific user.

#### GET /api/audit/patients/:id
Get audit history for specific patient.

#### GET /api/audit/security-events
Get security events (admin only).

#### GET /api/audit/statistics
Get audit statistics (admin only).

### Admin

#### GET /api/admin/users
Get all users (admin only).

#### POST /api/admin/users
Create new user (admin only).

#### GET /api/admin/users/:id
Get user by ID (admin only).

#### PUT /api/admin/users/:id
Update user (admin only).

#### POST /api/admin/users/:id/deactivate
Deactivate user (admin only).

## Error Responses

All endpoints return consistent error responses:

```json
{
  "error": "Error description"
}
```

### HTTP Status Codes

- `200 OK`: Request successful
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid request data
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

## Rate Limiting

API requests are rate limited to 100 requests per hour per IP address. Rate limit headers are included in responses:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1640995200
```

## Security Headers

All responses include security headers:

```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000; includeSubDomains
Content-Security-Policy: default-src 'self'
```

## Audit Logging

All API operations are automatically logged with:
- User ID and role
- Action performed
- Resource accessed
- IP address and user agent
- Success/failure status
- Timestamp
- Emergency access flag (if applicable)

## Emergency Access

Emergency access allows medical staff to access patient data in critical situations:

1. Request emergency access with justification
2. System generates time-limited access token
3. Use token in `X-Emergency-Access-Token` header
4. All emergency access is logged and audited
5. Access automatically expires after configured duration

## HIPAA Compliance

The API implements HIPAA technical safeguards:

- **Access Control**: Role-based permissions with unique user identification
- **Audit Controls**: Comprehensive logging of all data access
- **Integrity**: Data protection and validation
- **Person or Entity Authentication**: JWT-based authentication
- **Transmission Security**: HTTPS encryption for all communications

## Examples

### Complete Patient Workflow

```bash
# Login
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"doctor@hospital.com","password":"password123"}'

# Get patients
curl -H "Authorization: Bearer <token>" \
  http://localhost:8080/api/patients?limit=10

# Get specific patient
curl -H "Authorization: Bearer <token>" \
  http://localhost:8080/api/patients/1

# Create medical record
curl -X POST http://localhost:8080/api/patients/1/records \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "diagnosis": "Common Cold",
    "treatment": "Rest and fluids",
    "medications": "None",
    "severity": "low"
  }'

# Request emergency access
curl -X POST http://localhost:8080/api/emergency/request \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "patient_id": 1,
    "reason": "Emergency situation requiring immediate access"
  }'
```