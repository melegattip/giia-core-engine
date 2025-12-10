-- Create activation_tokens table
CREATE TABLE IF NOT EXISTS activation_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    user_id UUID NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_activation_tokens_user
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_activation_tokens_user_id ON activation_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_activation_tokens_token_hash ON activation_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_activation_tokens_expires_at ON activation_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_activation_tokens_used ON activation_tokens(used) WHERE used = false;

-- Create cleanup function for expired tokens
CREATE OR REPLACE FUNCTION cleanup_expired_activation_tokens()
RETURNS void AS $$
BEGIN
    DELETE FROM activation_tokens WHERE expires_at < NOW() OR used = true;
END;
$$ LANGUAGE plpgsql;
