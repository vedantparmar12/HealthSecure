# HealthSecure Database Schema

## Auto-Migration (GORM)

When the backend starts, it will **automatically create** all these tables:

### 1. `users` Table
Stores user accounts (doctors, nurses, admins)
```sql
CREATE TABLE users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(191) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,        -- bcrypt hashed
    role VARCHAR(20) NOT NULL,             -- 'doctor', 'nurse', 'admin'
    name VARCHAR(100) NOT NULL,
    active BOOLEAN DEFAULT TRUE,
    last_login DATETIME,
    created_at DATETIME,
    updated_at DATETIME,
    INDEX idx_users_email (email)
);
```

**Demo Users Created:**
- `dr.smith@hospital.local` / `Doctor123` (doctor)
- `nurse.jane@hospital.local` / `Nurse123` (nurse)
- `admin@hospital.local` / `admin123` (admin)

### 2. `patients` Table
Patient records
```sql
CREATE TABLE patients (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    date_of_birth DATE,
    gender VARCHAR(20),
    contact_number VARCHAR(20),
    email VARCHAR(191),
    address TEXT,
    emergency_contact VARCHAR(200),
    medical_history TEXT,
    allergies TEXT,
    blood_type VARCHAR(10),
    created_at DATETIME,
    updated_at DATETIME
);
```

### 3. `medical_records` Table
Medical records for patients
```sql
CREATE TABLE medical_records (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    patient_id BIGINT UNSIGNED NOT NULL,
    doctor_id BIGINT UNSIGNED NOT NULL,
    diagnosis TEXT,
    treatment TEXT,
    notes TEXT,
    medications TEXT,
    severity VARCHAR(20),
    created_at DATETIME,
    updated_at DATETIME,
    INDEX idx_medical_records_patient_id (patient_id),
    INDEX idx_medical_records_doctor_id (doctor_id)
);
```

### 4. `audit_logs` Table
Security and audit trail
```sql
CREATE TABLE audit_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id BIGINT UNSIGNED,
    ip_address VARCHAR(50),
    user_agent TEXT,
    details TEXT,
    created_at DATETIME,
    INDEX idx_audit_logs_user_id (user_id),
    INDEX idx_audit_logs_created_at (created_at)
);
```

### 5. `emergency_access` Table
Emergency access requests
```sql
CREATE TABLE emergency_access (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    requester_id BIGINT UNSIGNED NOT NULL,
    patient_id BIGINT UNSIGNED NOT NULL,
    reason TEXT NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    activated_at DATETIME,
    expires_at DATETIME,
    revoked_at DATETIME,
    created_at DATETIME,
    updated_at DATETIME
);
```

### 6. `blacklisted_tokens` Table
Invalidated JWT tokens
```sql
CREATE TABLE blacklisted_tokens (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME,
    INDEX idx_blacklisted_tokens_expires_at (expires_at)
);
```

### 7. `chat_threads` Table
AI chat conversation threads
```sql
CREATE TABLE chat_threads (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    patient_id BIGINT UNSIGNED,
    title VARCHAR(255),
    created_at DATETIME,
    updated_at DATETIME
);
```

### 8. `chat_messages` Table
Individual chat messages
```sql
CREATE TABLE chat_messages (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    thread_id BIGINT UNSIGNED NOT NULL,
    content TEXT NOT NULL,
    role VARCHAR(20) NOT NULL,  -- 'user', 'assistant'
    created_at DATETIME
);
```

### 9. `chat_feedback` Table
User feedback on AI responses
```sql
CREATE TABLE chat_feedback (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    message_id BIGINT UNSIGNED,
    user_id BIGINT UNSIGNED NOT NULL,
    rating INT,
    feedback_text TEXT,
    created_at DATETIME
);
```

### 10. `system_settings` Table
Application configuration
```sql
CREATE TABLE system_settings (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    key VARCHAR(100) UNIQUE NOT NULL,
    value TEXT,
    description TEXT,
    updated_by BIGINT UNSIGNED,
    created_at DATETIME,
    updated_at DATETIME
);
```

---

## What Happens on First Start

1. **Backend connects to Railway MySQL**
2. **GORM Auto-Migration runs**:
   - Checks which tables exist
   - Creates missing tables with proper schemas
   - Adds indexes for performance
3. **Demo users are seeded**:
   - Three users created with hashed passwords
   - One doctor, one nurse, one admin
4. **Database is ready to use!**

---

## Verification

After starting the backend, you should see:
```
✅ Loaded configuration from: ../../../configs/.env
✅ Database connection established successfully
✅ Running database migrations...
✅ Database migrations completed successfully
✅ Seeding demo users for development...
✅ Demo users seeded successfully
✅ Listening and serving HTTP on :8081
```

---

## Manual Schema Verification (Optional)

If you want to verify the schema in Railway:

1. Go to Railway dashboard
2. Open your MySQL database
3. Click "Query" or connect via MySQL client
4. Run: `SHOW TABLES;`
5. Should see all 10 tables listed above

Or test via API:
```bash
curl http://localhost:8081/health
```

Should return:
```json
{
  "status": "healthy",
  "database": "connected",
  "version": "1.0.0",
  "environment": "development"
}
```
