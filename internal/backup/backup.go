package backup

import (
	"bytes"
	"compress/gzip"
	"database/sql"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"pg-backup/internal/config"
	"pg-backup/internal/logger"
	"pg-backup/internal/storage"

	_ "github.com/lib/pq"
)

type Service struct {
	dbConfig *config.Config
	storage  storage.Provider
	logger   *logger.Logger
}

func NewService(dbConfig *config.Config, storage storage.Provider, logger *logger.Logger) *Service {
	return &Service{
		dbConfig: dbConfig,
		storage:  storage,
		logger:   logger,
	}
}

func (s *Service) BackupAll() (int, error) {
	// Check if full dump is enabled
	if s.dbConfig.FullDump {
		s.logger.Info("Full dump mode enabled, creating single backup file for entire server")
		err := s.backupFullServer()
		if err != nil {
			s.logger.Error("Failed to perform full dump: %v", err)
			return 0, err
		}
		s.logger.Info("Full dump completed successfully")
		return 1, nil
	}

	databases := s.dbConfig.Database.Databases

	// If no databases specified, discover all databases
	if len(databases) == 0 {
		s.logger.Info("No specific databases configured, discovering all databases")
		discoveredDbs, err := s.discoverDatabases()
		if err != nil {
			s.logger.Error("Failed to discover databases: %v", err)
			return 0, err
		}
		databases = discoveredDbs
		s.logger.Info("Discovered %d databases: %s", len(databases), strings.Join(databases, ", "))
	}

	for _, database := range databases {
		s.logger.Info("Starting backup for database: %s", database)

		err := s.backupDatabase(database)
		if err != nil {
			s.logger.Error("Failed to backup database %s: %v", database, err)
			return 0, err
		}

		s.logger.Info("Successfully backed up database: %s", database)
	}

	return len(databases), nil
}

func (s *Service) discoverDatabases() ([]string, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s dbname=postgres sslmode=disable",
		s.dbConfig.Database.Host,
		s.dbConfig.Database.Port,
		s.dbConfig.Database.User,
	)

	if s.dbConfig.Database.Password != "" {
		connStr += fmt.Sprintf(" password=%s", s.dbConfig.Database.Password)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	query := `
		SELECT datname 
		FROM pg_database 
		WHERE datistemplate = false 
		AND datname NOT IN ('postgres')
		ORDER BY datname
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query databases: %w", err)
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbName string
		err := rows.Scan(&dbName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan database name: %w", err)
		}
		databases = append(databases, dbName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating database rows: %w", err)
	}

	return databases, nil
}

func (s *Service) backupDatabase(database string) error {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("%s_%s.sql.gz", database, timestamp)

	cmd := exec.Command("pg_dump",
		"-h", s.dbConfig.Database.Host,
		"-p", fmt.Sprintf("%d", s.dbConfig.Database.Port),
		"-U", s.dbConfig.Database.User,
		"-d", database,
		"--no-password",
	)

	if s.dbConfig.Database.Password != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("PGPASSWORD=%s", s.dbConfig.Database.Password))
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	s.logger.Info("Executing pg_dump for database: %s", database)
	start := time.Now()

	err := cmd.Run()
	if err != nil {
		s.logger.Error("pg_dump failed for database %s: %v, stderr: %s", database, err, stderr.String())
		return fmt.Errorf("pg_dump failed: %w", err)
	}

	duration := time.Since(start)
	s.logger.Info("pg_dump completed for database %s in %v", database, duration)

	if stderr.Len() > 0 {
		s.logger.Warning("pg_dump warnings for database %s: %s", database, stderr.String())
	}

	var compressed bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressed)
	_, err = gzipWriter.Write(stdout.Bytes())
	if err != nil {
		s.logger.Error("Failed to compress backup for database %s: %v", database, err)
		return fmt.Errorf("failed to compress backup: %w", err)
	}
	err = gzipWriter.Close()
	if err != nil {
		s.logger.Error("Failed to close gzip writer for database %s: %v", database, err)
		return fmt.Errorf("failed to close gzip writer: %w", err)
	}

	originalSize := stdout.Len()
	compressedSize := compressed.Len()
	compressionRatio := float64(compressedSize) / float64(originalSize) * 100

	s.logger.Info("Backup compressed: %s (original: %d bytes, compressed: %d bytes, ratio: %.1f%%)",
		filename, originalSize, compressedSize, compressionRatio)

	s.logger.Info("Storing backup file: %s", filename)
	err = s.storage.Store(filename, &compressed)
	if err != nil {
		s.logger.Error("Failed to store backup for database %s: %v", database, err)
		return fmt.Errorf("failed to store backup: %w", err)
	}

	s.logger.Info("Backup stored successfully: %s (%d bytes compressed)", filename, compressedSize)
	return nil
}

func (s *Service) backupFullServer() error {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("full_dump_%s.sql.gz", timestamp)

	// Check if pg_dumpall is available - try multiple locations
	var pgDumpallPath string
	possiblePaths := []string{
		"pg_dumpall",
		"/usr/bin/pg_dumpall",
		"/usr/libexec/postgresql/pg_dumpall",
	}

	for _, path := range possiblePaths {
		if _, err := exec.LookPath(path); err == nil {
			pgDumpallPath = path
			break
		}
	}

	if pgDumpallPath == "" {
		s.logger.Error("pg_dumpall not found in any expected location. Full dump requires PostgreSQL client tools to be installed.")
		return fmt.Errorf("pg_dumpall not available")
	}

	s.logger.Info("Using pg_dumpall from: %s", pgDumpallPath)

	// Use pg_dumpall to create a full cluster dump including all databases, roles, and tablespaces
	cmd := exec.Command(pgDumpallPath,
		"-h", s.dbConfig.Database.Host,
		"-p", fmt.Sprintf("%d", s.dbConfig.Database.Port),
		"-U", s.dbConfig.Database.User,
		"--no-password",
	)

	if s.dbConfig.Database.Password != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("PGPASSWORD=%s", s.dbConfig.Database.Password))
		// Also add the current environment to ensure PATH is preserved
		cmd.Env = append(cmd.Env, "PATH=/usr/libexec/postgresql:/usr/bin:/usr/sbin:/bin:/sbin")
	} else {
		cmd.Env = []string{"PATH=/usr/libexec/postgresql:/usr/bin:/usr/sbin:/bin:/sbin"}
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	s.logger.Info("Executing pg_dumpall for full server dump")
	start := time.Now()

	err := cmd.Run()
	if err != nil {
		s.logger.Error("pg_dumpall failed: %v, stderr: %s", err, stderr.String())
		s.logger.Error("This might indicate missing PostgreSQL client tools or insufficient permissions")
		s.logger.Error("Consider using individual database backups (set full_dump: false) if pg_dumpall is not available")
		return fmt.Errorf("pg_dumpall failed: %w", err)
	}

	duration := time.Since(start)
	s.logger.Info("pg_dumpall completed in %v", duration)

	if stderr.Len() > 0 {
		s.logger.Warning("pg_dumpall warnings: %s", stderr.String())
	}

	var compressed bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressed)
	_, err = gzipWriter.Write(stdout.Bytes())
	if err != nil {
		s.logger.Error("Failed to compress full dump: %v", err)
		return fmt.Errorf("failed to compress full dump: %w", err)
	}
	err = gzipWriter.Close()
	if err != nil {
		s.logger.Error("Failed to close gzip writer for full dump: %v", err)
		return fmt.Errorf("failed to close gzip writer: %w", err)
	}

	originalSize := stdout.Len()
	compressedSize := compressed.Len()
	compressionRatio := float64(compressedSize) / float64(originalSize) * 100

	s.logger.Info("Full dump compressed: %s (original: %d bytes, compressed: %d bytes, ratio: %.1f%%)",
		filename, originalSize, compressedSize, compressionRatio)

	s.logger.Info("Storing full dump file: %s", filename)
	err = s.storage.Store(filename, &compressed)
	if err != nil {
		s.logger.Error("Failed to store full dump: %v", err)
		return fmt.Errorf("failed to store full dump: %w", err)
	}

	s.logger.Info("Full dump stored successfully: %s (%d bytes compressed)", filename, compressedSize)
	return nil
}
