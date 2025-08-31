-- Migration: 001_init.sql
-- Description: Create initial users table

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    phone VARCHAR(20) UNIQUE NOT NULL,
    registered_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone);
CREATE INDEX IF NOT EXISTS idx_users_registered_at ON users(registered_at);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

-- Add comments for documentation
COMMENT ON TABLE users IS 'User accounts for OTP authentication system';
COMMENT ON COLUMN users.id IS 'Unique identifier for the user';
COMMENT ON COLUMN users.phone IS 'Phone number in E.164 format (unique)';
COMMENT ON COLUMN users.registered_at IS 'Timestamp when user was registered';
COMMENT ON COLUMN users.created_at IS 'Timestamp when record was created';
COMMENT ON COLUMN users.updated_at IS 'Timestamp when record was last updated';
