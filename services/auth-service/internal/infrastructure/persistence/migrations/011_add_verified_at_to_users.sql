-- Add verified_at column to users table
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name='users' AND column_name='verified_at') THEN
        ALTER TABLE users ADD COLUMN verified_at TIMESTAMP;
    END IF;
END $$;

-- Create index on verified_at for queries
CREATE INDEX IF NOT EXISTS idx_users_verified_at ON users(verified_at);

-- Update existing active users to have verified_at set to created_at
UPDATE users
SET verified_at = created_at
WHERE status = 'active' AND verified_at IS NULL;
