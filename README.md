# PostgreSQL Backup Application

A robust Go application for automatically backing up PostgreSQL databases with flexible storage options and comprehensive monitoring.

## Features

- **Flexible Database Selection**: Backup specific databases or automatically discover and backup ALL databases
- **Full Dump Mode**: Create a single backup file containing all databases, roles, and tablespaces using pg_dumpall
- **Multiple Storage Options**: Local filesystem or AWS S3 bucket
- **Scheduled Backups**: Configurable cron-based scheduling
- **Compression**: Automatic gzip compression of backup files
- **Health Monitoring**: HTTP endpoints for health checks and status monitoring
- **Manual Backup Trigger**: HTTP API to trigger backups on-demand
- **Comprehensive Logging**: Detailed logs with timestamps and operation tracking
- **Docker Support**: Ready for containerized deployment
- **CLI Interface**: Command-line options for manual operations

## Quick Start

1. **Build the application:**

   ```bash
   make build
   # or
   go build -o pg-backup .
   ```

2. **Configure your backup:**

   ```bash
   cp config.example.yaml config.yaml
   # Edit config.yaml with your settings
   ```

3. **Run a test backup:**

   ```bash
   ./pg-backup -once
   ```

4. **Start the scheduler:**
   ```bash
   ./pg-backup
   ```

## Configuration

### Specific Databases

```yaml
database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "postgres"
  databases:
    - "database1"
    - "database2"
```

### Full Server Backup (Individual Files)

```yaml
database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "postgres"
  databases: [] # Empty = backup ALL databases
full_dump: false # Each database backed up separately
```

### Full Server Backup (Single File)

```yaml
database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "postgres"
  databases: [] # Ignored when full_dump is true
full_dump: true # Creates single file with all databases, roles, and tablespaces
```

### Storage Options

**Local Storage:**

```yaml
storage:
  type: "local"
  local:
    path: "./backups"
```

**S3 Storage:**

```yaml
storage:
  type: "s3"
  s3:
    bucket: "my-backup-bucket"
    region: "us-east-1"
    access_key: "ACCESS_KEY"
    secret_key: "SECRET_KEY"
```

## Commands

- `./pg-backup -list` - List configured databases
- `./pg-backup -once` - Run backup once and exit
- `./pg-backup -config custom.yaml` - Use custom configuration
- `./pg-backup -h` - Show help

## Manual Backup Trigger

Trigger a backup manually via HTTP API while the scheduler is running:

```bash
# Using curl
curl -X POST http://localhost:8080/trigger

# Using the provided script
./trigger-backup.sh
```

**Response on success (202 Accepted):**

```json
{
  "status": "accepted",
  "message": "Backup started successfully",
  "started_at": "2024-08-05 18:30:45"
}
```

The backup runs asynchronously in the background. Check the logs or use the `/status` endpoint to monitor progress.

## Health Monitoring

When running, the application provides HTTP endpoints:

- `http://localhost:8080/health` - Basic health check
- `http://localhost:8080/status` - Detailed status information
- `http://localhost:8080/trigger` - Manually trigger a backup (POST only)

## Docker Deployment

```bash
make docker-build
make docker-run
```

## Configuration Files

- `config.example.yaml` - Example with specific databases
- `config.full-dump.example.yaml` - Example for full server backup (individual files per database)
- `config.full-dump-single-file.example.yaml` - Example for full server backup (single file with pg_dumpall)
- `config.s3.example.yaml` - Example with S3 storage

## Backup Modes

### Individual Database Backups

- Uses `pg_dump` for each database
- Creates separate `.sql.gz` files for each database
- Allows selective restoration of specific databases
- Faster for partial restores

### Full Dump Mode

- Uses `pg_dumpall` to create a complete cluster backup
- Creates a single `.sql.gz` file containing everything
- Includes all databases, roles, tablespaces, and global objects
- Required for complete server restoration including user roles and permissions
- Enabled by setting `full_dump: true` in configuration

## Troubleshooting

### PostgreSQL Client Tools Not Found

If you encounter errors like:

```
pg_dumpall: error: program "pg_dump" is needed by pg_dumpall but was not found
```

**For Docker deployments:**

1. Rebuild the Docker image to ensure PostgreSQL client tools are properly installed
2. The Dockerfile uses `postgresql15-client` package which should include all necessary tools

**For manual installations:**

1. Install PostgreSQL client tools:

   ```bash
   # Ubuntu/Debian
   sudo apt-get install postgresql-client

   # CentOS/RHEL
   sudo yum install postgresql

   # macOS
   brew install postgresql
   ```

2. Ensure `pg_dump` and `pg_dumpall` are in your PATH:
   ```bash
   which pg_dump
   which pg_dumpall
   ```

**Workaround:**
If `pg_dumpall` is not available, you can disable full dump mode by setting `full_dump: false` in your configuration. This will backup each database individually using `pg_dump`.

### Connection Issues

- Verify database host, port, username, and password in configuration
- Ensure the PostgreSQL server allows connections from your backup location
- Check firewall settings and pg_hba.conf configuration
