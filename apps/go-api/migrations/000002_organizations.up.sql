/*
 * Organizations
 */
CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    name VARCHAR(255) NOT NULL,

    address TEXT NOT NULL,
    city VARCHAR(100) NOT NULL,
    zip VARCHAR(20) NOT NULL,
    country VARCHAR(70) NOT NULL,
    place_id VARCHAR(255) NOT NULL,
    geo GEOMETRY(Point,4326) NOT NULL,
    timezone VARCHAR(100) NOT NULL,

    billing_email CITEXT NOT NULL,
    custom_subscription BOOLEAN NOT NULL DEFAULT FALSE,
    subscription_expires_at TIMESTAMP WITH TIME ZONE,
    subscription_type SMALLINT NOT NULL,
    subscription_seats INT NOT NULL DEFAULT 1,
    subscription_payment_interval SMALLINT NOT NULL DEFAULT 0, -- 0=monthly, 1=annual
    subscription_in_trial BOOLEAN NOT NULL DEFAULT FALSE,

    -- Paddle
    paddle_subscription_id VARCHAR(255),
    paddle_customer_id VARCHAR(255) NOT NULL,
    -- When set: org keeps current plan until subscription_expires_at, then switches to this Paddle price (e.g. annual→monthly).
    scheduled_plan_price_id VARCHAR(255)
);

CREATE INDEX IF NOT EXISTS idx_organizations_deleted_at ON organizations(deleted_at);
CREATE INDEX IF NOT EXISTS idx_organizations_geo ON organizations USING GIST (geo) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_organizations_paddle_customer_id ON organizations(paddle_customer_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_organizations_paddle_subscription_id ON organizations(paddle_subscription_id) WHERE paddle_subscription_id IS NOT NULL AND deleted_at IS NULL;

/* 
* Roles
*/
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),

    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    name VARCHAR(255) NOT NULL,
    description TEXT,
    permissions JSONB NOT NULL DEFAULT '[]'::jsonb,
    locked BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_roles_org ON roles(organization_id);

/* 
* Teams
*/
CREATE TABLE IF NOT EXISTS teams (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    name VARCHAR(255) NOT NULL,
    description TEXT,

    CONSTRAINT uniq_team_name_org UNIQUE (organization_id, name)
);

CREATE INDEX IF NOT EXISTS idx_teams_org ON teams(organization_id);

/*
 * Organization Members
 */
CREATE TABLE IF NOT EXISTS organization_members (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    
    CONSTRAINT org_members_unique UNIQUE (organization_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_org_members_org ON organization_members(organization_id);
CREATE INDEX IF NOT EXISTS idx_org_members_user ON organization_members(user_id);
CREATE INDEX IF NOT EXISTS idx_org_members_role ON organization_members(role_id);

/*
 * Team Members
 */
CREATE TABLE IF NOT EXISTS team_members (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    CONSTRAINT team_members_unique UNIQUE (team_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_team_members_org ON team_members(organization_id);
CREATE INDEX IF NOT EXISTS idx_team_members_team ON team_members(team_id);
CREATE INDEX IF NOT EXISTS idx_team_members_user ON team_members(user_id);

/*
 * Invitations
 */
CREATE TABLE IF NOT EXISTS invitations (
  id UUID PRIMARY KEY,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
  expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
  accepted_at TIMESTAMP WITH TIME ZONE,
  revoked_at TIMESTAMP WITH TIME ZONE,
  
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  email CITEXT NOT NULL,
  
  invited_by_id UUID REFERENCES users(id) ON DELETE SET NULL,
  invited_by_email TEXT NOT NULL,
  token_hash TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_invitations_token ON invitations(token_hash);
CREATE INDEX IF NOT EXISTS idx_invitations_org ON invitations(organization_id) WHERE accepted_at IS NULL AND revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_invitations_email ON invitations(email) WHERE accepted_at IS NULL AND revoked_at IS NULL;

/*
 * API Keys
 */
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,

    name VARCHAR(255) NOT NULL,
    key_prefix VARCHAR(20) NOT NULL,
    key_hash TEXT NOT NULL,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    permissions JSONB NOT NULL DEFAULT '[]'::jsonb
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_api_keys_hash ON api_keys(key_hash);
CREATE INDEX IF NOT EXISTS idx_api_keys_org ON api_keys(organization_id);

/*
 * Audit Logs
 */
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    actor_id UUID NOT NULL, -- User ID or API Key ID
    actor_info TEXT NOT NULL, -- User email or API Key name
    impersonated_by_id UUID REFERENCES users(id) ON DELETE SET NULL,
    impersonated_by_email TEXT NOT NULL,
    
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id UUID NOT NULL,
    resource_name TEXT,

    data JSONB,

    CONSTRAINT audit_logs_action_check CHECK (action <> '')
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_org ON audit_logs(organization_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor ON audit_logs(actor_id, created_at DESC) WHERE actor_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_audit_logs_impersonated_by ON audit_logs(impersonated_by_id, created_at DESC) WHERE impersonated_by_id IS NOT NULL;
