-- api/internal/db/migrations/003_dynamic_rbac.sql
-- Karı Dynamic RBAC: Designed Secure, Made Simple.

BEGIN;

-- 1. Roles Table
-- Defines the containers for permissions.
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT false, -- SLA: Protects Super Admin from deletion
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. Permissions Table
-- Defines granular actions on specific resources.
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource VARCHAR(50) NOT NULL, -- e.g., 'applications', 'domains'
    action VARCHAR(50) NOT NULL,   -- e.g., 'deploy', 'read', 'write'
    description TEXT,
    UNIQUE(resource, action)
);

-- 3. Role-Permission Mapping (The Join Table)
-- Connects roles to their granular rights.
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- 4. User-Role Assignment
-- Links users to a specific role.
ALTER TABLE users ADD COLUMN IF NOT EXISTS role_id UUID REFERENCES roles(id);

-- ==============================================================================
-- Seed Data: The "Super Admin" Safety Net
-- ==============================================================================

-- Create the immutable Super Admin role
INSERT INTO roles (name, description, is_system) 
VALUES ('Super Admin', 'Total system authority. Cannot be modified or deleted.', true)
ON CONFLICT (name) DO NOTHING;

-- Seed baseline permissions for the Karı platform
INSERT INTO permissions (resource, action, description) VALUES
    ('applications', 'read', 'View deployment list and status'),
    ('applications', 'write', 'Create and edit application settings'),
    ('applications', 'deploy', 'Trigger manual GitOps deployments'),
    ('domains', 'read', 'View configured virtual hosts'),
    ('domains', 'ssl', 'Provision Let''s Encrypt certificates'),
    ('audit', 'read', 'Access system alerts and tenant logs'),
    ('rbac', 'manage', 'Modify roles and permissions')
ON CONFLICT (resource, action) DO NOTHING;

-- Grant all baseline permissions to the Super Admin role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'Super Admin'
ON CONFLICT DO NOTHING;

COMMIT;
