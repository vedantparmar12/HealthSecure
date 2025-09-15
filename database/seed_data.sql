-- HealthSecure Sample Data for Development and Testing
-- This data is for development purposes only and should NOT be used in production

-- Insert sample users with different roles
-- Note: All passwords are 'Doctor123' hashed with bcrypt cost 12
INSERT INTO users (email, password, role, name, active, last_login) VALUES
('admin@healthsecure.local', '$2a$12$rQzPqVCfbSl8GvXqNPOjc.TnbrQKOGP8cNWlv4P2RcNkXhEGOXhLy', 'admin', 'System Administrator', TRUE, NOW()),
('dr.smith@hospital.local', '$2a$12$rQzPqVCfbSl8GvXqNPOjc.TnbrQKOGP8cNWlv4P2RcNkXhEGOXhLy', 'doctor', 'Dr. John Smith', TRUE, DATE_SUB(NOW(), INTERVAL 2 HOUR)),
('dr.johnson@hospital.local', '$2a$12$rQzPqVCfbSl8GvXqNPOjc.TnbrQKOGP8cNWlv4P2RcNkXhEGOXhLy', 'doctor', 'Dr. Sarah Johnson', TRUE, DATE_SUB(NOW(), INTERVAL 1 DAY)),
('nurse.wilson@hospital.local', '$2a$12$rQzPqVCfbSl8GvXqNPOjc.TnbrQKOGP8cNWlv4P2RcNkXhEGOXhLy', 'nurse', 'Nurse Emily Wilson', TRUE, DATE_SUB(NOW(), INTERVAL 30 MINUTE)),
('nurse.brown@hospital.local', '$2a$12$rQzPqVCfbSl8GvXqNPOjc.TnbrQKOGP8cNWlv4P2RcNkXhEGOXhLy', 'nurse', 'Nurse Michael Brown', TRUE, DATE_SUB(NOW(), INTERVAL 4 HOUR)),
('admin.davis@hospital.local', '$2a$12$rQzPqVCfbSl8GvXqNPOjc.TnbrQKOGP8cNWlv4P2RcNkXhEGOXhLy', 'admin', 'Administrator Jane Davis', TRUE, DATE_SUB(NOW(), INTERVAL 6 HOUR));

-- Insert sample patients (using fake data for testing)
INSERT INTO patients (first_name, last_name, date_of_birth, ssn, phone, address, emergency_contact) VALUES
('John', 'Doe', '1985-03-15', '123-45-6789', '+1-555-0123', '123 Main St, Anytown, ST 12345', 'Jane Doe (Wife) - +1-555-0124'),
('Jane', 'Smith', '1978-07-22', '234-56-7890', '+1-555-0234', '456 Oak Ave, Somewhere, ST 23456', 'Bob Smith (Husband) - +1-555-0235'),
('Robert', 'Johnson', '1992-11-08', '345-67-8901', '+1-555-0345', '789 Pine Rd, Elsewhere, ST 34567', 'Mary Johnson (Mother) - +1-555-0346'),
('Emily', 'Davis', '1965-02-14', '456-78-9012', '+1-555-0456', '321 Elm St, Nowhere, ST 45678', 'Tom Davis (Son) - +1-555-0457'),
('Michael', 'Wilson', '1990-09-30', '567-89-0123', '+1-555-0567', '654 Maple Dr, Anyplace, ST 56789', 'Lisa Wilson (Sister) - +1-555-0568'),
('Sarah', 'Brown', '1983-12-05', '678-90-1234', '+1-555-0678', '987 Cedar Ln, Someplace, ST 67890', 'David Brown (Brother) - +1-555-0679'),
('James', 'Miller', '1976-06-18', '789-01-2345', '+1-555-0789', '147 Birch St, Elsewhere, ST 78901', 'Anna Miller (Wife) - +1-555-0790'),
('Linda', 'Garcia', '1988-04-25', '890-12-3456', '+1-555-0890', '258 Spruce Ave, Nowhere, ST 89012', 'Carlos Garcia (Husband) - +1-555-0891');

-- Insert sample medical records
INSERT INTO medical_records (patient_id, doctor_id, diagnosis, treatment, notes, medications, severity) VALUES
-- John Doe's records
(1, 2, 'Hypertension', 'Prescribed ACE inhibitor, lifestyle modifications', 'Patient shows good compliance with medication. Blood pressure improving.', 'Lisinopril 10mg daily', 'medium'),
(1, 2, 'Annual Physical', 'Routine checkup completed', 'Overall health is good. Recommended continued monitoring of blood pressure.', 'Continue current medications', 'low'),

-- Jane Smith's records  
(2, 2, 'Type 2 Diabetes', 'Metformin therapy, dietary counseling', 'HbA1c levels have improved significantly. Patient is managing diet well.', 'Metformin 500mg twice daily', 'medium'),
(2, 3, 'Diabetic Retinopathy Screening', 'Eye examination shows early changes', 'Mild non-proliferative diabetic retinopathy detected. Annual follow-up recommended.', 'No medications prescribed', 'medium'),

-- Robert Johnson's records
(3, 2, 'Acute Bronchitis', 'Antibiotic therapy and rest', 'Patient responded well to treatment. Symptoms resolved within one week.', 'Azithromycin 250mg daily for 5 days', 'low'),
(3, 2, 'Sports Physical', 'Cleared for athletic participation', 'Young athlete in excellent condition. No restrictions on activity.', 'None', 'low'),

-- Emily Davis's records
(4, 3, 'Coronary Artery Disease', 'Cardiac catheterization, stent placement', 'Successful PCI with drug-eluting stent. Patient stable post-procedure.', 'Clopidogrel 75mg daily, Atorvastatin 40mg daily, Metoprolol 50mg twice daily', 'high'),
(4, 3, 'Post-Cardiac Procedure Follow-up', 'Excellent recovery progress', 'Patient doing very well. Exercise tolerance improving. Continue current regimen.', 'Continue current cardiac medications', 'medium'),

-- Michael Wilson's records
(5, 2, 'Anxiety Disorder', 'Cognitive behavioral therapy referral', 'Patient benefits from therapy sessions. Anxiety levels decreased significantly.', 'Sertraline 50mg daily', 'medium'),

-- Sarah Brown's records
(6, 3, 'Migraine Headaches', 'Preventive medication and trigger avoidance', 'Frequency of migraines reduced from weekly to monthly with treatment.', 'Sumatriptan 50mg as needed, Propranolol 40mg daily', 'medium'),

-- James Miller's records
(7, 2, 'Prostate Cancer', 'Radical prostatectomy completed', 'Surgery successful. PSA levels undetectable at 3-month follow-up.', 'No current medications', 'critical'),
(7, 2, 'Post-Surgical Follow-up', 'Excellent surgical recovery', 'Patient recovering well. No signs of complications. Continue monitoring.', 'None required', 'low'),

-- Linda Garcia's records
(8, 3, 'Pregnancy - First Trimester', 'Prenatal care initiated', 'Normal first trimester. All screening tests within normal limits.', 'Prenatal vitamins', 'low'),
(8, 3, 'Gestational Diabetes Screening', 'Glucose tolerance test abnormal', 'Gestational diabetes diagnosed. Nutritionist consultation scheduled.', 'Prenatal vitamins, glucose monitoring supplies', 'medium');

-- Insert sample audit logs to demonstrate the audit trail
INSERT INTO audit_logs (user_id, patient_id, record_id, action, resource, ip_address, user_agent, emergency_use, success, timestamp) VALUES
-- Recent logins
(2, NULL, NULL, 'LOGIN', '/api/auth/login', '192.168.1.100', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36', FALSE, TRUE, DATE_SUB(NOW(), INTERVAL 30 MINUTE)),
(4, NULL, NULL, 'LOGIN', '/api/auth/login', '192.168.1.101', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36', FALSE, TRUE, DATE_SUB(NOW(), INTERVAL 45 MINUTE)),

-- Patient data access
(2, 1, NULL, 'VIEW', '/api/patients/1', '192.168.1.100', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36', FALSE, TRUE, DATE_SUB(NOW(), INTERVAL 25 MINUTE)),
(2, 1, 1, 'VIEW', '/api/patients/1/records/1', '192.168.1.100', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36', FALSE, TRUE, DATE_SUB(NOW(), INTERVAL 24 MINUTE)),
(4, 2, NULL, 'VIEW', '/api/patients/2', '192.168.1.101', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36', FALSE, TRUE, DATE_SUB(NOW(), INTERVAL 40 MINUTE)),

-- Medical record updates
(2, 1, 2, 'CREATE', '/api/patients/1/records', '192.168.1.100', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36', FALSE, TRUE, DATE_SUB(NOW(), INTERVAL 20 MINUTE)),
(3, 4, 6, 'UPDATE', '/api/records/6', '192.168.1.102', 'Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15', FALSE, TRUE, DATE_SUB(NOW(), INTERVAL 15 MINUTE)),

-- Failed access attempts (security events)
(4, 4, NULL, 'VIEW', '/api/patients/4', '192.168.1.101', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36', FALSE, FALSE, DATE_SUB(NOW(), INTERVAL 35 MINUTE)),
(0, NULL, NULL, 'UNAUTHORIZED_ACCESS', '/api/admin/users', '192.168.1.999', 'curl/7.68.0', FALSE, FALSE, DATE_SUB(NOW(), INTERVAL 2 HOUR));

-- Insert a sample emergency access request
INSERT INTO emergency_access (user_id, patient_id, reason, access_token, status, expires_at, created_at) VALUES
(4, 1, 'Patient brought to ER unconscious after car accident. Need immediate access to medical history for emergency treatment.', 
 'emergency_token_sample_123456789', 'pending', DATE_ADD(NOW(), INTERVAL 1 HOUR), NOW());

-- Insert corresponding audit log for emergency access request
INSERT INTO audit_logs (user_id, patient_id, action, resource, ip_address, user_agent, emergency_use, reason, success, timestamp) VALUES
(4, 1, 'EMERGENCY_REQUEST', '/api/emergency/request', '192.168.1.101', 
 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36', TRUE, 
 'Patient brought to ER unconscious after car accident. Need immediate access to medical history for emergency treatment.', 
 TRUE, NOW());

-- Insert sample security events
INSERT INTO security_events (event_type, severity, user_id, ip_address, description, resolved) VALUES
('FAILED_LOGIN', 'MEDIUM', NULL, '192.168.1.200', 'Multiple failed login attempts detected from this IP address', FALSE),
('EMERGENCY_ACCESS', 'HIGH', 4, '192.168.1.101', 'Emergency access requested for unconscious patient', FALSE),
('SUSPICIOUS_ACTIVITY', 'MEDIUM', NULL, '10.0.0.50', 'Unusual access pattern detected - multiple rapid requests', FALSE);

-- Insert some blacklisted tokens (for testing token revocation)
INSERT INTO blacklisted_tokens (token_hash, user_id, expires_at) VALUES
('dummy_token_hash_1', 2, DATE_ADD(NOW(), INTERVAL 15 MINUTE)),
('dummy_token_hash_2', 4, DATE_ADD(NOW(), INTERVAL 15 MINUTE));

-- Insert sample user sessions
INSERT INTO user_sessions (user_id, session_id, ip_address, user_agent, expires_at, last_activity) VALUES
(2, 'session_12345', '192.168.1.100', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36', DATE_ADD(NOW(), INTERVAL 30 MINUTE), NOW()),
(4, 'session_67890', '192.168.1.101', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36', DATE_ADD(NOW(), INTERVAL 30 MINUTE), DATE_SUB(NOW(), INTERVAL 5 MINUTE)),
(3, 'session_11111', '192.168.1.102', 'Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15', DATE_ADD(NOW(), INTERVAL 30 MINUTE), DATE_SUB(NOW(), INTERVAL 10 MINUTE));

-- Update system settings with more realistic values for development
INSERT INTO system_settings (setting_key, setting_value, description, updated_by) VALUES 
('maintenance_mode', 'false', 'Enable maintenance mode to restrict access', 1),
('max_session_duration_hours', '8', 'Maximum session duration in hours', 1),
('password_policy_enabled', 'true', 'Enable password complexity requirements', 1),
('emergency_access_approval_required', 'false', 'Require admin approval for emergency access', 1),
('audit_log_level', 'detailed', 'Level of detail in audit logs (basic, detailed, verbose)', 1)
ON DUPLICATE KEY UPDATE 
setting_value = VALUES(setting_value),
updated_by = VALUES(updated_by),
updated_at = CURRENT_TIMESTAMP;

-- Create some test data summary views
SELECT 'Sample data insertion completed successfully!' as status;

SELECT 'Users created:' as summary, COUNT(*) as count FROM users;
SELECT 'Patients created:' as summary, COUNT(*) as count FROM patients; 
SELECT 'Medical records created:' as summary, COUNT(*) as count FROM medical_records;
SELECT 'Audit logs created:' as summary, COUNT(*) as count FROM audit_logs;
SELECT 'Emergency access requests:' as summary, COUNT(*) as count FROM emergency_access;
SELECT 'Security events:' as summary, COUNT(*) as count FROM security_events;