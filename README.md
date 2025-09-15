# HealthSecure - Comprehensive Documentation

## Table of Contents

1. [Project Overview](#project-overview)
2. [Architecture Overview](#architecture-overview)
3. [System Architecture](#system-architecture)
4. [Database Schema](#database-schema)
5. [Security Architecture](#security-architecture)
6. [Project Structure](#project-structure)
7. [Backend Components](#backend-components)
8. [Frontend Components](#frontend-components)
9. [API Documentation](#api-documentation)
10. [Security Features](#security-features)
11. [HIPAA Compliance](#hipaa-compliance)
12. [Deployment Guide](#deployment-guide)
13. [Testing](#testing)
14. [Development Setup](#development-setup)

---

## Project Overview

HealthSecure is a comprehensive healthcare management system designed with HIPAA compliance and security as top priorities. The system provides secure management of patient medical records, emergency access protocols, and comprehensive audit logging.

### Key Features

- **Secure Patient Management**: Comprehensive patient record management with role-based access control
- **Emergency Access System**: Break-glass access for emergency situations with full audit trails
- **HIPAA Compliance**: Built-in compliance features including audit logging and data encryption
- **Role-Based Access Control**: Granular permissions system for healthcare providers
- **OAuth Integration**: Secure authentication with OAuth providers
- **Audit Logging**: Comprehensive tracking of all system access and modifications

---

## Architecture Overview

HealthSecure follows a modern three-tier architecture with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                     Frontend (React)                        │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │    Dashboard    │  │  Patient Mgmt   │  │  Emergency      │ │
│  │                 │  │                 │  │  Access         │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │   Audit Logs    │  │    Profile      │  │     Login       │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                                │
                                │ HTTPS/REST API
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                    Backend (Go)                             │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │   Handlers      │  │   Middleware    │  │   Services      │ │
│  │   - Auth        │  │   - JWT Auth    │  │   - User        │ │
│  │   - Patients    │  │   - RBAC        │  │   - Patient     │ │
│  │   - Emergency   │  │   - Audit       │  │   - Emergency   │ │
│  │   - Admin       │  │   - CORS        │  │   - Audit       │ │
│  │   - Audit       │  └─────────────────┘  │   - Med Records │ │
│  └─────────────────┘                       └─────────────────┘ │
│                                │                               │
│  ┌─────────────────┐           │           ┌─────────────────┐ │
│  │     Models      │           │           │      Auth       │ │
│  │   - User        │           │           │   - JWT         │ │
│  │   - Patient     │           │           │   - OAuth       │ │
│  │   - MedRecord   │           │           │   - Middleware  │ │
│  │   - AuditLog    │           │           └─────────────────┘ │
│  │   - Emergency   │           │                               │
│  └─────────────────┘           │                               │
└─────────────────────────────────┼─────────────────────────────┘
                                │
                                │ Database Queries
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                   Database (PostgreSQL)                     │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │     users       │  │    patients     │  │ medical_records │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
│  ┌─────────────────┐  ┌─────────────────┐                     │
│  │   audit_logs    │  │emergency_access │                     │
│  └─────────────────┘  └─────────────────┘                     │
└─────────────────────────────────────────────────────────────┘
```

---

## System Architecture

### Component Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              CLIENT LAYER                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│  React Frontend (Port 3000)                                                │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐      │
│  │  Components  │ │   Contexts   │ │    Pages     │ │   Services   │      │
│  │   - Layout   │ │ - AuthContext│ │ - Dashboard  │ │   - API      │      │
│  │   - Header   │ │              │ │ - Patients   │ │   - Auth     │      │
│  │   - Sidebar  │ │              │ │ - Emergency  │ │              │      │
│  └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘      │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      │ HTTP/REST API
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           APPLICATION LAYER                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│  Go Backend Server (Port 8080)                                             │
│                                                                             │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐      │
│  │   Handlers   │ │  Middleware  │ │   Services   │ │    Models    │      │
│  │              │ │              │ │              │ │              │      │
│  │ ┌──────────┐ │ │ ┌──────────┐ │ │ ┌──────────┐ │ │ ┌──────────┐ │      │
│  │ │   Auth   │ │ │ │ JWT Auth │ │ │ │   User   │ │ │ │   User   │ │      │
│  │ └──────────┘ │ │ └──────────┘ │ │ └──────────┘ │ │ └──────────┘ │      │
│  │ ┌──────────┐ │ │ ┌──────────┐ │ │ ┌──────────┐ │ │ ┌──────────┐ │      │
│  │ │ Patients │ │ │ │   RBAC   │ │ │ │ Patient  │ │ │ │ Patient  │ │      │
│  │ └──────────┘ │ │ └──────────┘ │ │ └──────────┘ │ │ └──────────┘ │      │
│  │ ┌──────────┐ │ │ ┌──────────┐ │ │ ┌──────────┐ │ │ ┌──────────┐ │      │
│  │ │Emergency │ │ │ │  Audit   │ │ │ │Emergency │ │ │ │AuditLog  │ │      │
│  │ └──────────┘ │ │ └──────────┘ │ │ └──────────┘ │ │ └──────────┘ │      │
│  │ ┌──────────┐ │ │ ┌──────────┐ │ │ ┌──────────┐ │ │ ┌──────────┐ │      │
│  │ │  Admin   │ │ │ │   CORS   │ │ │ │  Audit   │ │ │ │MedRecord │ │      │
│  │ └──────────┘ │ │ └──────────┘ │ │ └──────────┘ │ │ └──────────┘ │      │
│  │ ┌──────────┐ │ │              │ │ ┌──────────┐ │ │ ┌──────────┐ │      │
│  │ │  Audit   │ │ │              │ │ │MedRecord │ │ │ │Emergency │ │      │
│  │ └──────────┘ │ │              │ │ └──────────┘ │ │ └──────────┘ │      │
│  └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘      │
│                                                                             │
│  ┌──────────────┐ ┌──────────────┐                                        │
│  │     Auth     │ │   Database   │                                        │
│  │              │ │  Connection  │                                        │
│  │ ┌──────────┐ │ │              │                                        │
│  │ │   JWT    │ │ │ ┌──────────┐ │                                        │
│  │ └──────────┘ │ │ │Connection│ │                                        │
│  │ ┌──────────┐ │ │ └──────────┘ │                                        │
│  │ │  OAuth   │ │ │              │                                        │
│  │ └──────────┘ │ │              │                                        │
│  └──────────────┘ └──────────────┘                                        │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      │ SQL Queries
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              DATA LAYER                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│  PostgreSQL Database                                                        │
│                                                                             │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐      │
│  │    users     │ │   patients   │ │medical_records│ │  audit_logs  │      │
│  │              │ │              │ │              │ │              │      │
│  │ - id         │ │ - id         │ │ - id         │ │ - id         │      │
│  │ - username   │ │ - mrn        │ │ - patient_id │ │ - user_id    │      │
│  │ - email      │ │ - first_name │ │ - record_type│ │ - action     │      │
│  │ - password   │ │ - last_name  │ │ - content    │ │ - resource   │      │
│  │ - role       │ │ - dob        │ │ - created_at │ │ - timestamp  │      │
│  │ - created_at │ │ - address    │ │ - updated_at │ │ - ip_address │      │
│  │ - updated_at │ │ - phone      │ │              │ │ - user_agent │      │
│  └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘      │
│                                                                             │
│  ┌──────────────┐                                                          │
│  │emergency_    │                                                          │
│  │   access     │                                                          │
│  │              │                                                          │
│  │ - id         │                                                          │
│  │ - user_id    │                                                          │
│  │ - patient_id │                                                          │
│  │ - reason     │                                                          │
│  │ - accessed_at│                                                          │
│  │ - expires_at │                                                          │
│  └──────────────┘                                                          │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Database Schema

### Entity Relationship Diagram

```
┌─────────────────┐         ┌─────────────────┐         ┌─────────────────┐
│      users      │         │    patients     │         │medical_records  │
├─────────────────┤    ┌────├─────────────────┤    ┌────├─────────────────┤
│ id (PK)         │    │    │ id (PK)         │    │    │ id (PK)         │
│ username        │    │    │ mrn (UNIQUE)    │    │    │ patient_id (FK) │
│ email           │    │    │ first_name      │    │    │ record_type     │
│ password_hash   │    │    │ last_name       │    │    │ content         │
│ role            │    │    │ date_of_birth   │    │    │ created_by (FK) │
│ is_active       │    │    │ gender          │    │    │ created_at      │
│ created_at      │    │    │ phone           │    │    │ updated_at      │
│ updated_at      │    │    │ email           │    │    │ version         │
│ last_login      │    │    │ address         │    │    └─────────────────┘
└─────────────────┘    │    │ emergency_contact│   │              │
          │            │    │ created_at      │   │              │
          │            │    │ updated_at      │   │              │
          │            │    │ is_active       │   │              │
          │            │    └─────────────────┘   │              │
          │            │              │           │              │
          │            └──────────────┼───────────┘              │
          │                           │                          │
          │                           ▼                          │
          │            ┌─────────────────┐                       │
          │            │emergency_access │                       │
          │            ├─────────────────┤                       │
          │            │ id (PK)         │                       │
          │            │ user_id (FK)    │◄──────────────────────┘
          │            │ patient_id (FK) │
          │            │ reason          │
          │            │ justification   │
          │            │ accessed_at     │
          │            │ expires_at      │
          │            │ is_approved     │
          │            │ approved_by (FK)│
          │            └─────────────────┘
          │                           │
          ▼                           │
┌─────────────────┐                   │
│   audit_logs    │                   │
├─────────────────┤                   │
│ id (PK)         │                   │
│ user_id (FK)    │◄──────────────────┘
│ patient_id (FK) │
│ action          │
│ resource        │
│ resource_id     │
│ old_values      │
│ new_values      │
│ ip_address      │
│ user_agent      │
│ session_id      │
│ timestamp       │
│ severity        │
└─────────────────┘
```

### Table Relationships

- **users** → **medical_records** (1:N) - Users can create multiple medical records
- **patients** → **medical_records** (1:N) - Patients can have multiple medical records
- **users** → **emergency_access** (1:N) - Users can have multiple emergency access requests
- **patients** → **emergency_access** (1:N) - Patients can be subject to multiple emergency accesses
- **users** → **audit_logs** (1:N) - Users generate multiple audit log entries
- **patients** → **audit_logs** (1:N) - Patient-related actions generate audit logs

---

## Security Architecture

### Security Layers Diagram

```
┌───────────────────────────────────────────────────────────────────────────┐
│                            SECURITY LAYERS                                │
└───────────────────────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────────────────────┐
│                        1. TRANSPORT SECURITY                              │
├───────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐       │
│  │  HTTPS/TLS 1.3  │    │   CORS Policy   │    │  Rate Limiting  │       │
│  │  Certificate    │    │   - Origins     │    │  - Per IP       │       │
│  │  Encryption     │    │   - Methods     │    │  - Per User     │       │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘       │
└───────────────────────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────────────────────┐
│                       2. AUTHENTICATION LAYER                             │
├───────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐       │
│  │   JWT Tokens    │    │ OAuth Providers │    │ Session Mgmt    │       │
│  │  - Access Token │    │  - Google       │    │ - Timeout       │       │
│  │  - Refresh Token│    │  - Microsoft    │    │ - Invalidation  │       │
│  │  - Expiration   │    │  - Custom       │    │ - Tracking      │       │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘       │
└───────────────────────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────────────────────┐
│                       3. AUTHORIZATION LAYER                              │
├───────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐       │
│  │      RBAC       │    │  Permissions    │    │ Resource Access │       │
│  │  - Roles        │    │  - Read         │    │ - Patient Data  │       │
│  │  - Hierarchies  │    │  - Write        │    │ - Medical Rec.  │       │
│  │  - Assignments  │    │  - Delete       │    │ - Audit Logs    │       │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘       │
└───────────────────────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────────────────────┐
│                        4. APPLICATION SECURITY                            │
├───────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐       │
│  │ Input Validation│    │ Output Encoding │    │ Error Handling  │       │
│  │ - SQL Injection │    │ - XSS Prevention│    │ - Sanitized     │       │
│  │ - NoSQL Inject. │    │ - CSRF Tokens   │    │ - No Info Leak  │       │
│  │ - Command Inject│    │ - Content Type  │    │ - Logging       │       │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘       │
└───────────────────────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────────────────────┐
│                          5. DATA SECURITY                                 │
├───────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐       │
│  │  Encryption     │    │ Data Masking    │    │ Access Logging  │       │
│  │ - At Rest       │    │ - PII Fields    │    │ - All Access    │       │
│  │ - In Transit    │    │ - Sensitive Data│    │ - Failed Attempts│       │
│  │ - Key Rotation  │    │ - Role-based    │    │ - Data Changes  │       │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘       │
└───────────────────────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────────────────────┐
│                       6. EMERGENCY ACCESS SECURITY                        │
├───────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐       │
│  │ Break-Glass     │    │ Approval Chain  │    │ Time-Limited    │       │
│  │ - Emergency ID  │    │ - Multi-level   │    │ - Auto Expiry   │       │
│  │ - Justification │    │ - Supervisor    │    │ - Max Duration  │       │
│  │ - Full Audit    │    │ - Admin Override│    │ - Extension Req │       │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘       │
└───────────────────────────────────────────────────────────────────────────┘
```

---

## Project Structure

### Backend Structure (Go)

```
backend/
├── cmd/
│   ├── migrate/         # Database migration utilities
│   │   └── main.go     # Migration runner
│   └── server/         # Main server application
│       └── main.go     # Server entry point
├── configs/
│   └── config.go       # Configuration management
├── internal/
│   ├── auth/           # Authentication & authorization
│   │   ├── jwt.go      # JWT token handling
│   │   ├── middleware.go # Auth middleware
│   │   └── oauth.go    # OAuth integration
│   ├── database/       # Database connection & utilities
│   │   └── connection.go
│   ├── handlers/       # HTTP request handlers
│   │   ├── admin.go    # Admin operations
│   │   ├── audit.go    # Audit log endpoints
│   │   ├── auth.go     # Authentication endpoints
│   │   ├── emergency.go # Emergency access
│   │   ├── medical_records.go # Medical records
│   │   └── patients.go # Patient management
│   ├── middleware/     # HTTP middleware
│   │   ├── audit.go    # Audit logging
│   │   ├── auth.go     # Authentication
│   │   └── rbac.go     # Role-based access control
│   ├── models/         # Data models
│   │   ├── audit_log.go
│   │   ├── emergency_access.go
│   │   ├── medical_record.go
│   │   ├── patient.go
│   │   └── user.go
│   └── services/       # Business logic
│       ├── audit_service.go
│       ├── emergency_service.go
│       ├── medical_record_service.go
│       ├── patient_service.go
│       └── user_service.go
```

### Frontend Structure (React)

```
frontend/src/
├── components/         # Reusable UI components
│   └── Layout/
│       ├── Header.jsx  # Application header
│       ├── Layout.jsx  # Main layout wrapper
│       └── Sidebar.jsx # Navigation sidebar
├── contexts/           # React contexts
│   └── AuthContext.jsx # Authentication context
├── pages/              # Main application pages
│   ├── AuditLogs.jsx   # Audit log viewer
│   ├── Dashboard.jsx   # Main dashboard
│   ├── EmergencyAccess.jsx # Emergency access
│   ├── Login.jsx       # Login page
│   ├── PatientDetail.jsx # Patient details
│   ├── Patients.jsx    # Patient list
│   └── Profile.jsx     # User profile
└── services/           # API service layer
    └── api.js          # API communication
```

---

## Backend Components

### Core Services

#### 1. User Service
- **Purpose**: Manages user accounts, authentication, and profiles
- **Key Functions**:
  - User registration and login
  - Password management
  - Role assignment
  - Profile updates

#### 2. Patient Service
- **Purpose**: Manages patient records and information
- **Key Functions**:
  - Patient registration
  - Medical record number (MRN) generation
  - Patient search and filtering
  - Contact information management

#### 3. Medical Record Service
- **Purpose**: Handles medical record creation, retrieval, and management
- **Key Functions**:
  - Create/update medical records
  - Version control for records
  - Access control validation
  - Record type categorization

#### 4. Emergency Service
- **Purpose**: Manages emergency access requests and approvals
- **Key Functions**:
  - Emergency access request creation
  - Approval workflow
  - Time-limited access grants
  - Emergency audit trail

#### 5. Audit Service
- **Purpose**: Comprehensive logging and auditing system
- **Key Functions**:
  - Action logging
  - Security event tracking
  - Compliance reporting
  - Access pattern analysis

### Middleware Components

#### 1. Authentication Middleware
- JWT token validation
- Session management
- Token refresh handling

#### 2. Authorization Middleware (RBAC)
- Role-based access control
- Permission validation
- Resource-level access control

#### 3. Audit Middleware
- Automatic action logging
- Request/response tracking
- Security event detection

---

## Frontend Components

### Page Components

#### 1. Dashboard
- **Purpose**: Main overview page showing key metrics and recent activity
- **Features**:
  - Patient statistics
  - Recent records
  - Emergency access alerts
  - System notifications

#### 2. Patient Management
- **Purpose**: Patient listing and detail views
- **Features**:
  - Patient search and filtering
  - Patient registration
  - Medical record access
  - Contact management

#### 3. Emergency Access
- **Purpose**: Emergency access request and management interface
- **Features**:
  - Emergency access requests
  - Approval status tracking
  - Reason documentation
  - Time-limited access display

#### 4. Audit Logs
- **Purpose**: Audit trail viewing and analysis
- **Features**:
  - Log filtering and search
  - Security event highlighting
  - Export functionality
  - Compliance reporting

### Context Providers

#### AuthContext
- User authentication state
- Login/logout functionality
- Token management
- Role-based UI rendering

---

## API Documentation

### Authentication Endpoints

```
POST /api/auth/login
POST /api/auth/logout
POST /api/auth/refresh
GET  /api/auth/profile
PUT  /api/auth/profile
```

### Patient Management Endpoints

```
GET    /api/patients          # List patients
POST   /api/patients          # Create patient
GET    /api/patients/:id      # Get patient details
PUT    /api/patients/:id      # Update patient
DELETE /api/patients/:id      # Delete patient
GET    /api/patients/search   # Search patients
```

### Medical Records Endpoints

```
GET    /api/patients/:id/records     # Get patient records
POST   /api/patients/:id/records     # Create new record
GET    /api/records/:id              # Get specific record
PUT    /api/records/:id              # Update record
DELETE /api/records/:id              # Delete record
GET    /api/records/:id/versions     # Get record versions
```

### Emergency Access Endpoints

```
POST   /api/emergency/request        # Request emergency access
GET    /api/emergency/requests       # List requests
PUT    /api/emergency/:id/approve    # Approve request
PUT    /api/emergency/:id/deny       # Deny request
GET    /api/emergency/active         # List active accesses
POST   /api/emergency/:id/revoke     # Revoke access
```

### Audit Endpoints

```
GET    /api/audit/logs               # Get audit logs
GET    /api/audit/logs/:id           # Get specific log
GET    /api/audit/reports            # Generate reports
GET    /api/audit/export             # Export audit data
```

### Admin Endpoints

```
GET    /api/admin/users              # List all users
POST   /api/admin/users              # Create user
PUT    /api/admin/users/:id/role     # Update user role
PUT    /api/admin/users/:id/status   # Enable/disable user
GET    /api/admin/statistics         # System statistics
GET    /api/admin/health             # System health check
```

---

## Security Features

### 1. Authentication & Authorization

#### Multi-Factor Authentication
- JWT-based token authentication
- OAuth integration (Google, Microsoft)
- Session management with timeout
- Refresh token rotation

#### Role-Based Access Control (RBAC)
- **Admin**: Full system access, user management
- **Doctor**: Patient records, emergency access
- **Nurse**: Limited patient records, basic operations
- **Receptionist**: Patient registration, basic info
- **Auditor**: Read-only access to audit logs

### 2. Data Protection

#### Encryption
- **In Transit**: TLS 1.3 for all communications
- **At Rest**: Database-level encryption
- **Application Level**: Sensitive field encryption

#### Data Masking
- PII fields masked based on user roles
- Partial data display for unauthorized users
- Full data access only for authorized roles

### 3. Emergency Access Controls

#### Break-Glass Access
- Emergency access with full justification
- Time-limited access (configurable duration)
- Supervisor approval required
- Complete audit trail

#### Access Monitoring
- Real-time access monitoring
- Unusual pattern detection
- Automatic alert generation
- Access revocation capabilities

### 4. Audit & Compliance

#### Comprehensive Logging
- All user actions logged
- System events tracked
- Security events highlighted
- Failed access attempts recorded

#### HIPAA Compliance
- Complete audit trails
- Access logs for all PHI
- Data breach detection
- Compliance reporting

---

## HIPAA Compliance

### Administrative Safeguards

#### 1. Security Officer
- Designated security officer role
- Security management responsibilities
- Workforce training requirements
- Information access management

#### 2. Access Management
- Unique user identification
- Automatic logoff
- Encryption and decryption procedures

### Physical Safeguards

#### 1. Facility Access Controls
- Data center security requirements
- Physical access logging
- Workstation security

#### 2. Device and Media Controls
- Secure device management
- Data backup and recovery
- Secure data disposal

### Technical Safeguards

#### 1. Access Control
- Unique user identification
- Role-based access control
- Automatic logoff after inactivity
- Encryption of data in transit and at rest

#### 2. Audit Controls
- Comprehensive audit logging
- Regular security assessments
- Incident response procedures

#### 3. Integrity
- Data integrity controls
- Electronic signature requirements
- Version control for medical records

#### 4. Transmission Security
- End-to-end encryption
- Secure communication protocols
- Data transmission controls

### Compliance Features Implementation

```
┌─────────────────────────────────────────────────────────────────┐
│                    HIPAA COMPLIANCE MATRIX                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Administrative Safeguards:                                    │
│  ✅ Security Officer Assignment                                 │
│  ✅ Workforce Training Program                                  │
│  ✅ Information Access Management                               │
│  ✅ Security Awareness Training                                 │
│  ✅ Security Incident Procedures                                │
│  ✅ Contingency Plan                                            │
│  ✅ Regular Security Evaluations                                │
│                                                                 │
│  Physical Safeguards:                                          │
│  ✅ Facility Access Controls                                    │
│  ✅ Workstation Use Restrictions                                │
│  ✅ Device and Media Controls                                   │
│                                                                 │
│  Technical Safeguards:                                         │
│  ✅ Access Control (Unique User ID)                            │
│  ✅ Audit Controls (Comprehensive Logging)                     │
│  ✅ Integrity (Data Integrity Controls)                        │
│  ✅ Person or Entity Authentication                             │
│  ✅ Transmission Security (Encryption)                         │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## Deployment Guide

### Docker Deployment Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    DOCKER COMPOSE SETUP                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────┐    ┌─────────────────┐                    │
│  │   nginx-proxy   │    │   certbot       │                    │
│  │   (Port 80/443) │    │   (SSL Certs)   │                    │
│  │                 │    │                 │                    │
│  └─────────┬───────┘    └─────────────────┘                    │
│            │                                                    │
│            ▼                                                    │
│  ┌─────────────────┐    ┌─────────────────┐                    │
│  │   frontend      │    │    backend      │                    │
│  │   (React App)   │    │   (Go Server)   │                    │
│  │   Port 3000     │    │   Port 8080     │                    │
│  └─────────────────┘    └─────────┬───────┘                    │
│                                    │                            │
│                                    ▼                            │
│  ┌─────────────────┐    ┌─────────────────┐                    │
│  │   postgresql    │    │     redis       │                    │
│  │   (Database)    │    │   (Cache/Sess)  │                    │
│  │   Port 5432     │    │   Port 6379     │                    │
│  └─────────────────┘    └─────────────────┘                    │
│                                                                 │
│  ┌─────────────────┐    ┌─────────────────┐                    │
│  │   prometheus    │    │    grafana      │                    │
│  │  (Monitoring)   │    │ (Dashboards)    │                    │
│  │   Port 9090     │    │   Port 3001     │                    │
│  └─────────────────┘    └─────────────────┘                    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Environment Configuration

#### Production Environment Variables

```bash
# Database Configuration
DB_HOST=postgresql
DB_PORT=5432
DB_NAME=healthsecure
DB_USER=healthsecure_user
DB_PASSWORD=<secure-password>
DB_SSL_MODE=require

# Redis Configuration  
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=<redis-password>

# JWT Configuration
JWT_SECRET=<jwt-secret-key>
JWT_EXPIRATION=24h
JWT_REFRESH_EXPIRATION=7d

# OAuth Configuration
GOOGLE_CLIENT_ID=<google-client-id>
GOOGLE_CLIENT_SECRET=<google-client-secret>
MICROSOFT_CLIENT_ID=<microsoft-client-id>
MICROSOFT_CLIENT_SECRET=<microsoft-client-secret>

# Security Configuration
ENCRYPTION_KEY=<encryption-key>
CORS_ALLOWED_ORIGINS=https://yourdomain.com
RATE_LIMIT_REQUESTS=1000
RATE_LIMIT_WINDOW=1h

# Monitoring Configuration
PROMETHEUS_ENABLED=true
GRAFANA_ADMIN_PASSWORD=<grafana-password>
```

### Deployment Steps

#### 1. Prerequisites
```bash
# Install Docker and Docker Compose
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

#### 2. Clone and Setup
```bash
# Clone repository
git clone https://github.com/vedantparmar12/HealthSecure.git
cd HealthSecure

# Copy environment files
cp .env.example .env.production
# Edit .env.production with your configuration

# Generate SSL certificates (if not using automated)
./scripts/generate-ssl.sh
```

#### 3. Deploy
```bash
# Build and start services
docker-compose -f docker-compose.prod.yml up -d

# Run database migrations
docker-compose exec backend ./migrate up

# Verify deployment
docker-compose ps
```

### Scaling Configuration

#### Horizontal Scaling
```yaml
# docker-compose.scale.yml
version: '3.8'
services:
  backend:
    deploy:
      replicas: 3
      resources:
        limits:
          memory: 512M
          cpus: '0.5'
        reservations:
          memory: 256M
          cpus: '0.25'
  
  frontend:
    deploy:
      replicas: 2
      resources:
        limits:
          memory: 256M
          cpus: '0.25'
```

---

## Testing

### Test Architecture

```
tests/
├── backend/
│   ├── unit/              # Unit tests
│   │   ├── models_test.go
│   │   ├── services_test.go
│   │   └── handlers_test.go
│   ├── integration/       # Integration tests
│   │   ├── api_test.go
│   │   └── database_test.go
│   └── security/          # Security tests
│       ├── auth_test.go
│       └── rbac_test.go
└── frontend/
    ├── unit/              # Component tests
    │   ├── components/
    │   ├── contexts/
    │   └── services/
    ├── integration/       # Integration tests
    │   ├── pages/
    │   └── workflows/
    └── e2e/               # End-to-end tests
        ├── login.spec.js
        ├── patient-management.spec.js
        └── emergency-access.spec.js
```

### Test Coverage Requirements

#### Backend Testing (Go)
- **Unit Tests**: >90% coverage
- **Integration Tests**: API endpoints
- **Security Tests**: Authentication, authorization
- **Performance Tests**: Load testing

#### Frontend Testing (React)
- **Unit Tests**: Component testing
- **Integration Tests**: Page workflows
- **E2E Tests**: Critical user journeys
- **Accessibility Tests**: WCAG compliance

### Running Tests

#### Backend Tests
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test suite
go test ./internal/auth -v

# Run security tests
go test ./tests/backend/security_test.go -v
```

#### Frontend Tests
```bash
# Run unit tests
npm test

# Run with coverage
npm test -- --coverage

# Run E2E tests
npm run test:e2e

# Run accessibility tests
npm run test:a11y
```

---

## Development Setup

### Local Development Environment

#### Prerequisites
- Go 1.21+
- Node.js 18+
- PostgreSQL 14+
- Redis 6+
- Docker (optional)

#### Backend Setup
```bash
# Clone repository
git clone https://github.com/vedantparmar12/HealthSecure.git
cd HealthSecure/backend

# Install dependencies
go mod download

# Setup environment
cp configs/.env.example configs/.env.local
# Edit configs/.env.local

# Run database migrations
go run cmd/migrate/main.go up

# Start development server
go run cmd/server/main.go
```

#### Frontend Setup
```bash
# Navigate to frontend
cd ../frontend

# Install dependencies
npm install

# Setup environment
cp .env.example .env.local
# Edit .env.local

# Start development server
npm start
```

### Development Tools

#### Code Quality
```bash
# Backend linting
golangci-lint run

# Frontend linting
npm run lint

# Format code
gofmt -w .
npm run format
```

#### Database Management
```bash
# Create migration
go run cmd/migrate/main.go create migration_name

# Run migrations
go run cmd/migrate/main.go up

# Rollback migrations
go run cmd/migrate/main.go down
```

### Git Workflow

#### Branch Strategy
- **main**: Production-ready code
- **develop**: Development branch
- **feature/***: Feature branches
- **hotfix/***: Emergency fixes
- **release/***: Release preparation

#### Commit Standards
```bash
# Conventional commits
feat: add emergency access feature
fix: resolve authentication bug
docs: update API documentation
test: add patient service tests
refactor: optimize database queries
```

---

## Architecture Diagrams

### System Context Diagram

```
                    ┌─────────────────┐
                    │   Healthcare    │
                    │   Providers     │
                    │                 │
                    └────────┬────────┘
                             │
                             ▼
    ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
    │   Patients      │  │  HealthSecure   │  │   Regulators    │
    │                 │──│    System       │──│   (HIPAA)       │
    │                 │  │                 │  │                 │
    └─────────────────┘  └─────────────────┘  └─────────────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │   External      │
                    │   Systems       │
                    │   (EHR, Labs)   │
                    └─────────────────┘
```

### Component Interaction Diagram

```
Frontend (React)
    │
    │ HTTP/REST API
    ▼
API Gateway/Router
    │
    ├─── Authentication Service
    │    │
    │    ├─── JWT Handler
    │    └─── OAuth Handler
    │
    ├─── Authorization Service (RBAC)
    │    │
    │    ├─── Role Manager
    │    └─── Permission Checker
    │
    ├─── Business Services
    │    │
    │    ├─── Patient Service
    │    ├─── Medical Record Service
    │    ├─── Emergency Service
    │    └─── Audit Service
    │
    └─── Data Layer
         │
         ├─── PostgreSQL (Primary Data)
         └─── Redis (Cache/Sessions)
```

This comprehensive documentation provides a complete overview of the HealthSecure system, including detailed architecture diagrams, security implementations, deployment guides, and development workflows. The system is designed with HIPAA compliance and healthcare security best practices as core principles.

---

**Generated**: September 4, 2025  
**Version**: 1.0.0  
**Repository**: https://github.com/vedantparmar12/HealthSecure
