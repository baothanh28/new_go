-- Create tenants table based on Tenant model
CREATE TABLE IF NOT EXISTS tenants (
    id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    db_type VARCHAR(50) NOT NULL DEFAULT 'mysql',
    db_host VARCHAR(255) NOT NULL,
    db_port INTEGER NOT NULL,
    db_name VARCHAR(100) NOT NULL,
    db_user VARCHAR(100) NOT NULL,
    db_password VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert Tenant1 (pointing to tenant-db - PostgreSQL)
INSERT INTO tenants (id, name, db_type, db_host, db_port, db_name, db_user, db_password, is_active)
VALUES (
    'tenant1',
    'Tenant 1',
    'postgresql',
    'localhost',
    5433,
    'tenant_db',
    'postgres',
    'password',
    true
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    db_type = EXCLUDED.db_type,
    db_host = EXCLUDED.db_host,
    db_port = EXCLUDED.db_port,
    db_name = EXCLUDED.db_name,
    db_user = EXCLUDED.db_user,
    db_password = EXCLUDED.db_password,
    is_active = EXCLUDED.is_active,
    updated_at = CURRENT_TIMESTAMP;

-- Insert Tenant2 (pointing to tenant-db-1 - MySQL)
INSERT INTO tenants (id, name, db_type, db_host, db_port, db_name, db_user, db_password, is_active)
VALUES (
    'tenant2',
    'Tenant 2',
    'mysql',
    'localhost',
    3307,
    'tenant_db_1',
    'mysqluser',
    'mysqlpass',
    true
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    db_type = EXCLUDED.db_type,
    db_host = EXCLUDED.db_host,
    db_port = EXCLUDED.db_port,
    db_name = EXCLUDED.db_name,
    db_user = EXCLUDED.db_user,
    db_password = EXCLUDED.db_password,
    is_active = EXCLUDED.is_active,
    updated_at = CURRENT_TIMESTAMP;

-- Display inserted tenants
SELECT id, name, db_type, db_host, db_port, db_name, is_active FROM tenants;
