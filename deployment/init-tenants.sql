-- Create tenants table based on Tenant model
CREATE TABLE IF NOT EXISTS tenants (
    id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    db_type VARCHAR(50) NOT NULL DEFAULT 'mysql',
    cnn TEXT NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert Tenant1 (pointing to tenant-db - PostgreSQL)
INSERT INTO tenants (id, name, db_type, cnn, is_active)
VALUES (
    'tenant1',
    'Tenant 1',
    'postgresql',
    'host=localhost port=5433 user=postgres password=password dbname=tenant_db sslmode=disable',
    true
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    db_type = EXCLUDED.db_type,
    cnn = EXCLUDED.cnn,
    is_active = EXCLUDED.is_active,
    updated_at = CURRENT_TIMESTAMP;

-- Insert Tenant2 (pointing to tenant-db-1 - MySQL)
INSERT INTO tenants (id, name, db_type, cnn, is_active)
VALUES (
    'tenant2',
    'Tenant 2',
    'mysql',
    'mysqluser:mysqlpass@tcp(localhost:3307)/tenant_db_1?parseTime=true&loc=UTC&allowPublicKeyRetrieval=true',
    true
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    db_type = EXCLUDED.db_type,
    cnn = EXCLUDED.cnn,
    is_active = EXCLUDED.is_active,
    updated_at = CURRENT_TIMESTAMP;

-- Display inserted tenants
SELECT id, name, db_type, is_active, created_at, updated_at FROM tenants;
