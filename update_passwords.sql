-- Update all user passwords to "Doctor123" 
-- Hash generated using bcrypt cost 12: $2a$12$rQzPqVCfbSl8GvXqNPOjc.TnbrQKOGP8cNWlv4P2RcNkXhEGOXhLy

UPDATE users SET password = '$2a$12$rQzPqVCfbSl8GvXqNPOjc.TnbrQKOGP8cNWlv4P2RcNkXhEGOXhLy' WHERE email = 'admin@healthsecure.local';
UPDATE users SET password = '$2a$12$rQzPqVCfbSl8GvXqNPOjc.TnbrQKOGP8cNWlv4P2RcNkXhEGOXhLy' WHERE email = 'dr.smith@hospital.local';  
UPDATE users SET password = '$2a$12$rQzPqVCfbSl8GvXqNPOjc.TnbrQKOGP8cNWlv4P2RcNkXhEGOXhLy' WHERE email = 'dr.johnson@hospital.local';
UPDATE users SET password = '$2a$12$rQzPqVCfbSl8GvXqNPOjc.TnbrQKOGP8cNWlv4P2RcNkXhEGOXhLy' WHERE email = 'nurse.wilson@hospital.local';
UPDATE users SET password = '$2a$12$rQzPqVCfbSl8GvXqNPOjc.TnbrQKOGP8cNWlv4P2RcNkXhEGOXhLy' WHERE email = 'nurse.brown@hospital.local'; 
UPDATE users SET password = '$2a$12$rQzPqVCfbSl8GvXqNPOjc.TnbrQKOGP8cNWlv4P2RcNkXhEGOXhLy' WHERE email = 'admin.davis@hospital.local';

-- Verify the updates
SELECT email, 'Password updated to Doctor123' as status FROM users;