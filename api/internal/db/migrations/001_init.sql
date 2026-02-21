-- api/internal/db/migrations/001_init.sql

-- ==============================================================================
-- 1. Extensions & Global Functions
-- ==============================================================================

-- Ensure UUID generation is available (Native in PG13+, but safe to declare)
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Automated updated_at trigger function
-- This guarantees the DB enforces mutation timestamps, even if the Go API forgets.
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ==============================================================================
-- 2. Identity & Access Management (Dynamic RBAC)
-- ==============================================================================

CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource VARCHAR(100) NOT NULL, -- e.g., 'applications', 'server'
    action VARCHAR(100) NOT NULL,   -- e.g., 'read', 'write', 'deploy'
    UNIQUE(resource, action)
);

CREATE TABLE role_permissions (
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role_id UUID REFERENCES roles(id) ON DELETE RESTRICT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ==============================================================================
-- 3. Web & Infrastructure State
-- ==============================================================================

CREATE TABLE domains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    domain_name VARCHAR(255) UNIQUE NOT NULL,
    -- 🛡️ Flexible State Constraints: Safer than hardcoded Postgres ENUMs
    ssl_status VARCHAR(50) NOT NULL DEFAULT 'none' 
        CHECK (ssl_status IN ('none', 'active', 'renewing', 'failed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    domain_id UUID REFERENCES domains(id) ON DELETE RESTRICT,
    
    -- GitOps Configuration
    repo_url VARCHAR(1024) NOT NULL,
    branch VARCHAR(100) NOT NULL DEFAULT 'main',
    build_command VARCHAR(500) NOT NULL,
    start_command VARCHAR(500) NOT NULL,
    
    -- Networking & Jail Identity
    local_port INTEGER NOT NULL CHECK (local_port > 1024 AND local_port < 65536),
    app_user VARCHAR(100) UNIQUE NOT NULL, -- e.g., 'kari-app-123'
    
    -- 🛡️ JSONB Environment Variables
    env_vars JSONB NOT NULL DEFAULT '{}'::jsonb,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ==============================================================================
-- 4. Observability & GitOps Workflows
-- ==============================================================================

CREATE TABLE deployments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id UUID REFERENCES applications(id) ON DELETE CASCADE,
    trace_id VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending' 
        CHECK (status IN ('pending', 'building', 'success', 'failed')),
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

-- The Action Center / System Alerts Table
CREATE TABLE system_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    severity VARCHAR(50) NOT NULL 
        CHECK (severity IN ('info', 'warning', 'critical', 'fatal')),
    category VARCHAR(100) NOT NULL, -- e.g., 'ssl', 'deployment', 'systemd'
    resource_id VARCHAR(255),       -- Optional link to a specific app/domain
    message TEXT NOT NULL,
    is_resolved BOOLEAN NOT NULL DEFAULT false,
    
    -- 🛡️ JSONB Metadata (Holds trace_ids, raw Rust Agent error outputs, etc.)
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

-- ==============================================================================
-- 5. Performance Optimization (SLA Indices)
-- ==============================================================================

-- 🛡️ GIN Index for the Action Center
-- This allows the Go Brain to execute `metadata @> jsonb_build_object('trace_id', 'X')`
-- and scan 100,000+ alerts in sub-millisecond time without triggering sequential table scans.
CREATE INDEX idx_system_alerts_metadata ON system_alerts USING GIN (metadata);

-- Action Center UI Filters (Compound index for "Show me all unresolved critical alerts")
CREATE INDEX idx_system_alerts_status ON system_alerts (is_resolved, severity, created_at DESC);

-- GIN Index for Application Environment Variables (Allows searching for apps sharing a specific API key)
CREATE INDEX idx_applications_env_vars ON applications USING GIN (env_vars);

-- ==============================================================================
-- 6. Update Triggers
-- ==============================================================================

CREATE TRIGGER set_timestamp_users
BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();

CREATE TRIGGER set_timestamp_roles
BEFORE UPDATE ON roles FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();

CREATE TRIGGER set_timestamp_domains
BEFORE UPDATE ON domains FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();

CREATE TRIGGER set_timestamp_applications
BEFORE UPDATE ON applications FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();
