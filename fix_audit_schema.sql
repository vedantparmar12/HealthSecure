-- Update audit_logs table to allow NULL user_id for failed login attempts
-- This fixes the foreign key constraint violation issue

-- First, update the column to allow NULL
ALTER TABLE audit_logs MODIFY COLUMN user_id BIGINT UNSIGNED NULL;

-- Drop the existing foreign key constraint (if it exists)
-- Find the constraint name first
SELECT CONSTRAINT_NAME 
FROM information_schema.KEY_COLUMN_USAGE 
WHERE TABLE_SCHEMA = DATABASE() 
AND TABLE_NAME = 'audit_logs' 
AND COLUMN_NAME = 'user_id'
AND REFERENCED_TABLE_NAME = 'users';

-- Drop the foreign key constraint (replace 'fk_audit_logs_user' with actual constraint name from above query)
-- ALTER TABLE audit_logs DROP FOREIGN KEY fk_audit_logs_user;

-- Recreate the foreign key constraint to allow NULL values
-- ALTER TABLE audit_logs ADD CONSTRAINT fk_audit_logs_user FOREIGN KEY (user_id) REFERENCES users(id);

-- Verify the schema change
DESCRIBE audit_logs;