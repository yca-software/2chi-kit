/* 
 * Extensions 
 */
CREATE EXTENSION IF NOT EXISTS "postgis" WITH SCHEMA public;
CREATE EXTENSION IF NOT EXISTS "citext" WITH SCHEMA public;

/* 
 * Users 
 */
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),

    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    language VARCHAR(2) NOT NULL DEFAULT 'en',
    avatar_url TEXT NOT NULL DEFAULT '',

    email CITEXT NOT NULL,
    email_verified_at TIMESTAMP WITH TIME ZONE,
    password TEXT,
    google_id VARCHAR(255),

    terms_accepted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    terms_version VARCHAR(255) NOT NULL DEFAULT '1.0.0',
    
    CONSTRAINT chk_user_has_auth CHECK (
        (password IS NOT NULL) OR (google_id IS NOT NULL)
    )
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_google_id 
    ON users(google_id) WHERE google_id IS NOT NULL;

/* 
 * Admin Access 
 */
CREATE TABLE IF NOT EXISTS admin_access (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

/* 
 * User Refresh Tokens 
 */
CREATE TABLE IF NOT EXISTS user_refresh_tokens (
    id UUID PRIMARY KEY NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked_at TIMESTAMP WITH TIME ZONE,

    ip INET NOT NULL,
    user_agent TEXT NOT NULL,
    token_hash TEXT NOT NULL,
    impersonated_by UUID REFERENCES users(id) ON DELETE SET NULL,

    CONSTRAINT chk_refresh_expires_after_created CHECK (expires_at > created_at)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_refresh_tokens_hash 
    ON user_refresh_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_user_refresh_tokens_user 
    ON user_refresh_tokens(user_id)
    WHERE revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_user_refresh_tokens_impersonated_by 
    ON user_refresh_tokens(impersonated_by)
    WHERE impersonated_by IS NOT NULL;

/* 
 * User Password Reset Tokens 
*/
CREATE TABLE IF NOT EXISTS user_password_reset_tokens (
    id UUID PRIMARY KEY NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    
    token_hash TEXT NOT NULL,
    
    CONSTRAINT chk_reset_expires_after_created CHECK (expires_at > created_at)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_password_reset_tokens_hash 
    ON user_password_reset_tokens(token_hash);

/* 
 * User Email Verification Tokens 
*/
CREATE TABLE IF NOT EXISTS user_email_verification_tokens (
    id UUID PRIMARY KEY NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    
    token_hash TEXT NOT NULL,
    
    CONSTRAINT chk_email_verification_expires_after_created CHECK (expires_at > created_at)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_email_verification_tokens_hash 
    ON user_email_verification_tokens(token_hash);