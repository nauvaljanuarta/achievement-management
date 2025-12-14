-- 1. Roles
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Insert default roles
INSERT INTO roles (id, name, description)
VALUES 
    (gen_random_uuid(), 'Admin', 'Administrator system'),
    (gen_random_uuid(), 'Mahasiswa', 'Student role'),
    (gen_random_uuid(), 'Dosen Wali', 'Advisor role')
ON CONFLICT (name) DO NOTHING;

-- 2. Users
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    role_id UUID REFERENCES roles(id),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 3. Permissions
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT
);

INSERT INTO permissions (id, name, resource, action, description)
VALUES
  (gen_random_uuid(), 'achievement:create', 'achievement', 'create', 'Buat prestasi baru'),
  (gen_random_uuid(), 'achievement:read', 'achievement', 'read', 'Baca prestasi'),
  (gen_random_uuid(), 'achievement:update', 'achievement', 'update', 'Update prestasi'),
  (gen_random_uuid(), 'achievement:delete', 'achievement', 'delete', 'Hapus prestasi'),
  (gen_random_uuid(), 'achievement:verify', 'achievement', 'verify', 'Verifikasi prestasi'),
  (gen_random_uuid(), 'user:manage', 'user', 'manage', 'Manage user')
ON CONFLICT (id) DO NOTHING;

-- 4. Role_Permissions
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id UUID REFERENCES roles(id),
    permission_id UUID REFERENCES permissions(id),
    PRIMARY KEY (role_id, permission_id)
);

-- permission mapping 
INSERT INTO role_permissions (role_id, permission_id)
SELECT 
    r.id AS role_id,
    p.id AS permission_id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'Admin';

INSERT INTO role_permissions (role_id, permission_id)
SELECT 
    r.id AS role_id,
    p.id AS permission_id
FROM roles r, permissions p
WHERE r.name = 'Mahasiswa' 
AND p.name IN ('achievement:create', 'achievement:read', 'achievement:update', 'achievement:delete');

INSERT INTO role_permissions (role_id, permission_id)
SELECT 
    r.id AS role_id,
    p.id AS permission_id
FROM roles r, permissions p
WHERE r.name = 'Dosen Wali' 
AND p.name IN ('achievement:read', 'achievement:verify');