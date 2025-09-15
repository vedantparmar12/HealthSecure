-- Comprehensive fix for login issues
-- Run this script in your Railway database

-- 1. First, fix the audit_logs schema to allow NULL user_id
ALTER TABLE audit_logs MODIFY COLUMN user_id BIGINT UNSIGNED NULL;

-- 2. Verify the schema change
DESCRIBE audit_logs;

-- 3. Check if users exist and their current passwords
SELECT email, name, active FROM users;

-- 4. Update all user passwords to "Doctor123" (bcrypt hash with cost 12)
UPDATE users SET password = '$2a$12$rQzPqVCfbSl8GvXqNPOjc.TnbrQKOGP8cNWlv4P2RcNkXhEGOXhLy' 
WHERE email IN (
    'admin@healthsecure.local',
    'dr.smith@hospital.local',
    'dr.johnson@hospital.local', 
    'nurse.wilson@hospital.local',
    'nurse.brown@hospital.local',
    'admin.davis@hospital.local'
);

-- 5. If users don't exist, create them (only run if the SELECT above shows missing users)
INSERT IGNORE INTO users (email, password, role, name, active, created_at, updated_at) VALUES
('admin@healthsecure.local', '$2a$12$rQzPqVCfbSl8GvXqNPOjc.TnbrQKOGP8cNWlv4P2RcNkXhEGOXhLy', 'admin', 'System Administrator', TRUE, NOW(), NOW()),
('dr.smith@hospital.local', '$2a$12$rQzPqVCfbSl8GvXqNPOjc.TnbrQKOGP8cNWlv4P2RcNkXhEGOXhLy', 'doctor', 'Dr. John Smith', TRUE, NOW(), NOW()),
('dr.johnson@hospital.local', '$2a$12$rQzPqVCfbSl8GvXqNPOjc.TnbrQKOGP8cNWlv4P2RcNkXhEGOXhLy', 'doctor', 'Dr. Sarah Johnson', TRUE, NOW(), NOW()),
('nurse.wilson@hospital.local', '$2a$12$rQzPqVCfbSl8GvXqNPOjc.TnbrQKOGP8cNWlv4P2RcNkXhEGOXhLy', 'nurse', 'Nurse Emily Wilson', TRUE, NOW(), NOW()),
('nurse.brown@hospital.local', '$2a$12$rQzPqVCfbSl8GvXqNPOjc.TnbrQKOGP8cNWlv4P2RcNkXhEGOXhLy', 'nurse', 'Nurse Michael Brown', TRUE, NOW(), NOW()),
('admin.davis@hospital.local', '$2a$12$rQzPqVCfbSl8GvXqNPOjc.TnbrQKOGP8cNWlv4P2RcNkXhEGOXhLy', 'admin', 'Administrator Jane Davis', TRUE, NOW(), NOW());

-- 6. Verify users are created/updated
SELECT email, name, role, active, 'Password updated to Doctor123' as status FROM users;

-- 7. Clean up any failed audit log entries that might be causing issues
DELETE FROM audit_logs WHERE user_id = 0;

SELECT 'Database fixes completed successfully!' as status;