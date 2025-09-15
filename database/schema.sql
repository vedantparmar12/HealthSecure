-- HealthSecure Database Schema
-- HIPAA-compliant medical data access control system

-- Create database (run separately if needed)
-- CREATE DATABASE IF NOT EXISTS healthsecure CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
-- USE healthsecure;

-- Users table with role-based access
CREATE TABLE IF NOT EXISTS users (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    role ENUM('doctor', 'nurse', 'admin') NOT NULL,
    name VARCHAR(255) NOT NULL,
    active BOOLEAN DEFAULT TRUE,
    last_login TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_users_email (email),
    INDEX idx_users_role (role),
    INDEX idx_users_active (active)
);

-- Patients table with sensitive data protection
CREATE TABLE IF NOT EXISTS patients (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    date_of_birth DATE NOT NULL,
    ssn VARCHAR(11) UNIQUE, -- Sensitive data, doctor access only
    phone VARCHAR(20),
    address TEXT,
    emergency_contact VARCHAR(255),
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_patients_name (last_name, first_name),
    INDEX idx_patients_dob (date_of_birth),
    INDEX idx_patients_ssn (ssn)
);

-- Medical records with severity-based access control
CREATE TABLE IF NOT EXISTS medical_records (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    patient_id INT UNSIGNED NOT NULL,
    doctor_id INT UNSIGNED NOT NULL,
    diagnosis TEXT,
    treatment TEXT,
    notes TEXT,
    medications TEXT,
    severity ENUM('low', 'medium', 'high', 'critical') DEFAULT 'low',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (patient_id) REFERENCES patients(id) ON DELETE CASCADE,
    FOREIGN KEY (doctor_id) REFERENCES users(id) ON DELETE RESTRICT,
    
    INDEX idx_medical_patient (patient_id),
    INDEX idx_medical_doctor (doctor_id),
    INDEX idx_medical_severity (severity),
    INDEX idx_medical_created (created_at)
);

-- Comprehensive audit logging for HIPAA compliance
CREATE TABLE IF NOT EXISTS audit_logs (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id INT UNSIGNED NOT NULL,
    patient_id INT UNSIGNED NULL,
    record_id INT UNSIGNED NULL,
    action ENUM('LOGIN', 'LOGOUT', 'VIEW', 'CREATE', 'UPDATE', 'DELETE', 
                'EMERGENCY_REQUEST', 'EMERGENCY_ACCESS', 'UNAUTHORIZED_ACCESS') NOT NULL,
    resource VARCHAR(255) NOT NULL,
    ip_address VARCHAR(45) NOT NULL, -- IPv6 compatible
    user_agent TEXT,
    emergency_use BOOLEAN DEFAULT FALSE,
    reason TEXT,
    success BOOLEAN DEFAULT TRUE,
    error_message TEXT,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT,
    FOREIGN KEY (patient_id) REFERENCES patients(id) ON DELETE CASCADE,
    FOREIGN KEY (record_id) REFERENCES medical_records(id) ON DELETE CASCADE,
    
    INDEX idx_audit_user (user_id),
    INDEX idx_audit_patient (patient_id),
    INDEX idx_audit_action (action),
    INDEX idx_audit_timestamp (timestamp),
    INDEX idx_audit_emergency (emergency_use),
    INDEX idx_audit_success (success),
    INDEX idx_audit_ip (ip_address)
);

-- Emergency access control with time limits
CREATE TABLE IF NOT EXISTS emergency_access (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id INT UNSIGNED NOT NULL,
    patient_id INT UNSIGNED NOT NULL,
    reason TEXT NOT NULL,
    access_token VARCHAR(255) NOT NULL UNIQUE,
    status ENUM('pending', 'active', 'used', 'expired', 'revoked') DEFAULT 'pending',
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP NULL,
    revoked_at TIMESTAMP NULL,
    revoked_by INT UNSIGNED NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT,
    FOREIGN KEY (patient_id) REFERENCES patients(id) ON DELETE CASCADE,
    FOREIGN KEY (revoked_by) REFERENCES users(id) ON DELETE SET NULL,
    
    INDEX idx_emergency_user (user_id),
    INDEX idx_emergency_patient (patient_id),
    INDEX idx_emergency_token (access_token),
    INDEX idx_emergency_status (status),
    INDEX idx_emergency_expires (expires_at),
    INDEX idx_emergency_created (created_at)
);

-- Session management for JWT token blacklisting
CREATE TABLE IF NOT EXISTS blacklisted_tokens (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    user_id INT UNSIGNED NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    
    INDEX idx_blacklist_token (token_hash),
    INDEX idx_blacklist_expires (expires_at)
);

-- User sessions for enhanced security tracking
CREATE TABLE IF NOT EXISTS user_sessions (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id INT UNSIGNED NOT NULL,
    session_id VARCHAR(255) NOT NULL UNIQUE,
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    
    INDEX idx_sessions_user (user_id),
    INDEX idx_sessions_id (session_id),
    INDEX idx_sessions_expires (expires_at),
    INDEX idx_sessions_activity (last_activity)
);

-- System settings for configuration
CREATE TABLE IF NOT EXISTS system_settings (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    setting_key VARCHAR(255) NOT NULL UNIQUE,
    setting_value TEXT,
    description TEXT,
    updated_by INT UNSIGNED,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL,
    
    INDEX idx_settings_key (setting_key)
);

-- Security events for enhanced monitoring
CREATE TABLE IF NOT EXISTS security_events (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    event_type ENUM('FAILED_LOGIN', 'SUSPICIOUS_ACTIVITY', 'UNAUTHORIZED_ACCESS', 
                    'EMERGENCY_ACCESS', 'DATA_BREACH', 'SYSTEM_ALERT') NOT NULL,
    severity ENUM('LOW', 'MEDIUM', 'HIGH', 'CRITICAL') DEFAULT 'MEDIUM',
    user_id INT UNSIGNED NULL,
    ip_address VARCHAR(45),
    description TEXT NOT NULL,
    details JSON,
    resolved BOOLEAN DEFAULT FALSE,
    resolved_by INT UNSIGNED NULL,
    resolved_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (resolved_by) REFERENCES users(id) ON DELETE SET NULL,
    
    INDEX idx_security_type (event_type),
    INDEX idx_security_severity (severity),
    INDEX idx_security_user (user_id),
    INDEX idx_security_resolved (resolved),
    INDEX idx_security_created (created_at)
);

-- Insert default admin user (password should be changed immediately)
INSERT IGNORE INTO users (email, password, role, name, active) VALUES 
('admin@healthsecure.local', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj.2CRBPqe5e', 'admin', 'System Administrator', TRUE);

-- Insert system settings with default values
INSERT IGNORE INTO system_settings (setting_key, setting_value, description) VALUES 
('jwt_expires_minutes', '15', 'JWT token expiration time in minutes'),
('refresh_token_expires_days', '7', 'Refresh token expiration time in days'),
('emergency_access_duration_hours', '1', 'Emergency access duration in hours'),
('max_failed_login_attempts', '5', 'Maximum failed login attempts before lockout'),
('session_timeout_minutes', '30', 'User session timeout in minutes'),
('audit_retention_days', '2555', 'Audit log retention period in days (7 years for HIPAA)'),
('password_min_length', '8', 'Minimum password length requirement'),
('require_password_complexity', 'true', 'Require complex passwords');

-- Create stored procedures for common operations

DELIMITER //

-- Procedure to clean up expired sessions and tokens
CREATE PROCEDURE CleanupExpiredSessions()
BEGIN
    DECLARE done INT DEFAULT FALSE;
    DECLARE cleanup_count INT DEFAULT 0;
    
    -- Clean expired blacklisted tokens
    DELETE FROM blacklisted_tokens WHERE expires_at < NOW();
    SET cleanup_count = cleanup_count + ROW_COUNT();
    
    -- Clean expired user sessions
    DELETE FROM user_sessions WHERE expires_at < NOW();
    SET cleanup_count = cleanup_count + ROW_COUNT();
    
    -- Update expired emergency access
    UPDATE emergency_access 
    SET status = 'expired' 
    WHERE expires_at < NOW() AND status NOT IN ('expired', 'revoked');
    
    SELECT cleanup_count as cleaned_records;
END //

-- Procedure to get user permissions summary
CREATE PROCEDURE GetUserPermissions(IN user_id INT)
BEGIN
    SELECT 
        u.id,
        u.email,
        u.name,
        u.role,
        u.active,
        u.last_login,
        COUNT(DISTINCT al.id) as audit_log_count,
        COUNT(DISTINCT ea.id) as emergency_access_count,
        MAX(al.timestamp) as last_audit_activity
    FROM users u
    LEFT JOIN audit_logs al ON u.id = al.user_id
    LEFT JOIN emergency_access ea ON u.id = ea.user_id
    WHERE u.id = user_id
    GROUP BY u.id;
END //

DELIMITER ;

-- Create views for common queries

-- View for active emergency access sessions
CREATE VIEW active_emergency_sessions AS
SELECT 
    ea.*,
    u.name as user_name,
    u.email as user_email,
    u.role as user_role,
    p.first_name,
    p.last_name,
    TIMESTAMPDIFF(MINUTE, NOW(), ea.expires_at) as minutes_remaining
FROM emergency_access ea
JOIN users u ON ea.user_id = u.id
JOIN patients p ON ea.patient_id = p.id
WHERE ea.status = 'active' 
  AND ea.expires_at > NOW() 
  AND ea.revoked_at IS NULL;

-- View for recent security events
CREATE VIEW recent_security_events AS
SELECT 
    se.*,
    u.name as user_name,
    u.email as user_email,
    ru.name as resolved_by_name
FROM security_events se
LEFT JOIN users u ON se.user_id = u.id
LEFT JOIN users ru ON se.resolved_by = ru.id
WHERE se.created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)
ORDER BY se.created_at DESC;

-- View for audit summary
CREATE VIEW audit_summary AS
SELECT 
    DATE(al.timestamp) as audit_date,
    al.action,
    COUNT(*) as action_count,
    COUNT(DISTINCT al.user_id) as unique_users,
    COUNT(DISTINCT al.patient_id) as unique_patients,
    SUM(CASE WHEN al.emergency_use THEN 1 ELSE 0 END) as emergency_actions,
    SUM(CASE WHEN al.success = FALSE THEN 1 ELSE 0 END) as failed_actions
FROM audit_logs al
WHERE al.timestamp >= DATE_SUB(NOW(), INTERVAL 30 DAY)
GROUP BY DATE(al.timestamp), al.action
ORDER BY audit_date DESC, action_count DESC;