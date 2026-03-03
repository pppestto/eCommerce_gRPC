-- Add password_hash column for JWT auth (Variant A: user-service stores passwords)
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255);
