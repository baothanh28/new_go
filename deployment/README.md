# Deployment Configuration

This folder contains deployment configurations for the MyApp project.

## Docker Compose

The `docker-compose.yml` file sets up the required PostgreSQL databases for local development and testing.

### Services

#### Master Database (master-db)
- **Image**: `postgres:15-alpine`
- **Container Name**: `myapp-master-db`
- **Port**: `5432`
- **Database**: `master_db`
- **Purpose**: Stores system-wide data (users, authentication, tenants)

#### Tenant Database (tenant-db)
- **Image**: `postgres:15-alpine`
- **Container Name**: `myapp-tenant-db`
- **Port**: `5433` (mapped to internal 5432)
- **Database**: `tenant_db`
- **Purpose**: Stores tenant-specific data (products, orders, etc.)

### Quick Start

From the `src` directory:

```bash
# Start both databases
make docker-up

# Stop both databases
make docker-down

# View logs
make docker-logs
```

Or run directly from this folder:

```bash
# Start databases
docker-compose up -d

# Stop databases
docker-compose down

# View logs
docker-compose logs -f

# Check status
docker-compose ps
```

### Database Credentials

**Default credentials** (change in production):
- **Username**: `postgres`
- **Password**: `password`

### Volumes

Data is persisted in Docker volumes:
- `master-data` - Master database data
- `tenant-data` - Tenant database data

To remove all data:

```bash
docker-compose down -v
```

### Health Checks

Both databases include health checks:
- **Test Command**: `pg_isready -U postgres`
- **Interval**: 10 seconds
- **Timeout**: 5 seconds
- **Retries**: 5

### Connecting to Databases

#### Master Database
```bash
psql -h localhost -p 5432 -U postgres -d master_db
# Password: password
```

#### Tenant Database
```bash
psql -h localhost -p 5433 -U postgres -d tenant_db
# Password: password
```

### Using with Application

Update `src/config/config.yaml` to match these settings:

```yaml
master_database:
  host: "localhost"
  port: 5432
  name: "master_db"
  user: "postgres"
  password: "password"

tenant_database:
  host: "localhost"
  port: 5433
  name: "tenant_db"
  user: "postgres"
  password: "password"
```

### Production Deployment

For production, consider:

1. **Use Environment Variables** for sensitive data
2. **Use stronger passwords**
3. **Configure SSL/TLS** for database connections
4. **Use managed database services** (AWS RDS, Google Cloud SQL, Azure Database)
5. **Implement backup strategies**
6. **Configure monitoring and alerts**

### Troubleshooting

#### Port Already in Use

If ports 5432 or 5433 are already in use:

1. Stop the conflicting service
2. Or change the ports in `docker-compose.yml`:
   ```yaml
   ports:
     - "5434:5432"  # Change host port
   ```

#### Database Connection Issues

1. Check containers are running:
   ```bash
   docker ps
   ```

2. Check container logs:
   ```bash
   docker logs myapp-master-db
   docker logs myapp-tenant-db
   ```

3. Restart containers:
   ```bash
   docker-compose restart
   ```

#### Clean Start

To start fresh with no data:

```bash
docker-compose down -v
docker-compose up -d
```

Then re-run migrations from the application.

### Additional Configuration

#### Custom PostgreSQL Settings

You can add custom PostgreSQL configuration by mounting a config file:

```yaml
services:
  master-db:
    volumes:
      - master-data:/var/lib/postgresql/data
      - ./postgresql.conf:/etc/postgresql/postgresql.conf
    command: postgres -c config_file=/etc/postgresql/postgresql.conf
```

#### Resource Limits

Add resource limits for production:

```yaml
services:
  master-db:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '1'
          memory: 1G
```

## Future Deployment Configurations

This folder can be extended with:

- Kubernetes manifests (k8s/)
- Helm charts
- Terraform configurations
- CI/CD pipeline configurations
- Production docker-compose files
- Nginx/reverse proxy configurations
- SSL certificate configurations
