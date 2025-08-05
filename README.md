# PostgreSQL Backup Application

A robust Go application for automatically backing up PostgreSQL databases with flexible storage options and comprehensive monitoring.

## Features

- **Flexible Database Selection**: Backup specific databases or automatically discover and backup ALL databases
- **Multiple Storage Options**: Local filesystem or AWS S3 bucket
- **Scheduled Backups**: Configurable cron-based scheduling
- **Compression**: Automatic gzip compression of backup files
- **Health Monitoring**: HTTP endpoints for health checks and status monitoring
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

### Full Server Backup

```yaml
database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "postgres"
  databases: [] # Empty = backup ALL databases
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

## Health Monitoring

When running, the application provides HTTP endpoints:

- `http://localhost:8080/health` - Basic health check
- `http://localhost:8080/status` - Detailed status information

## Docker Deployment

```bash
make docker-build
make docker-run
```

## Configuration Files

- `config.example.yaml` - Example with specific databases
- `config.full-dump.example.yaml` - Example for full server backup
- `config.s3.example.yaml` - Example with S3 storage
