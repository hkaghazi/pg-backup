package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"pg-backup/internal/backup"
	"pg-backup/internal/config"
	"pg-backup/internal/health"
	"pg-backup/internal/logger"
	"pg-backup/internal/storage"

	"github.com/robfig/cron/v3"
)

func main() {
	var (
		configFile = flag.String("config", "config.yaml", "Configuration file path")
		runOnce    = flag.Bool("once", false, "Run backup once and exit")
		listDbs    = flag.Bool("list", false, "List configured databases and exit")
	)
	flag.Parse()

	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	if *listDbs {
		fmt.Println("Configured databases:")
		for i, db := range cfg.Database.Databases {
			fmt.Printf("  %d. %s\n", i+1, db)
		}
		return
	}

	appLogger := logger.New(cfg.LogFile)
	defer appLogger.Close()

	var storageProvider storage.Provider
	switch cfg.Storage.Type {
	case "local":
		storageProvider = storage.NewLocal(cfg.Storage.Local.Path)
	case "s3":
		s3Provider, err := storage.NewS3(cfg.Storage.S3.Bucket, cfg.Storage.S3.Region, cfg.Storage.S3.AccessKey, cfg.Storage.S3.SecretKey)
		if err != nil {
			appLogger.Error("Failed to initialize S3 storage: %v", err)
			os.Exit(1)
		}
		storageProvider = s3Provider
	default:
		appLogger.Error("Invalid storage type: %s", cfg.Storage.Type)
		os.Exit(1)
	}

	backupService := backup.NewService(cfg, storageProvider, appLogger)
	healthService := health.NewService(appLogger, len(cfg.Database.Databases))
	healthService.SetBackupService(backupService)

	go healthService.Start(cfg.HealthCheckPort)

	if *runOnce {
		appLogger.Info("Running one-time backup")
		dbCount, err := backupService.BackupAll()
		if err != nil {
			appLogger.Error("Backup failed: %v", err)
			os.Exit(1)
		}
		appLogger.Info("Backup completed successfully for %d databases", dbCount)
		return
	}

	runScheduler(cfg, backupService, appLogger, healthService)
}

func runScheduler(cfg *config.Config, backupService *backup.Service, appLogger *logger.Logger, healthService *health.Service) {
	appLogger.Info("Starting pg-backup scheduler")

	c := cron.New()

	_, err := c.AddFunc(cfg.Schedule, func() {
		appLogger.Info("Starting scheduled backup")
		start := time.Now()
		dbCount, err := backupService.BackupAll()
		if err != nil {
			appLogger.Error("Backup failed: %v", err)
		} else {
			appLogger.Info("Backup completed successfully for %d databases", dbCount)
			healthService.UpdateBackupStats(time.Now(), start.Add(24*time.Hour), 1, dbCount)
		}
	})

	if err != nil {
		appLogger.Error("Failed to schedule backup: %v", err)
		return
	}

	c.Start()
	appLogger.Info("Backup scheduler started with cron: %s", cfg.Schedule)

	if cfg.RunOnStart {
		appLogger.Info("Running initial backup")
		start := time.Now()
		dbCount, err := backupService.BackupAll()
		if err != nil {
			appLogger.Error("Initial backup failed: %v", err)
		} else {
			appLogger.Info("Initial backup completed successfully for %d databases", dbCount)
			healthService.UpdateBackupStats(time.Now(), start.Add(24*time.Hour), 1, dbCount)
		}
	}

	select {}
}
