package health

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"pg-backup/internal/logger"
)

type BackupService interface {
	BackupAll() (int, error)
}

type Status struct {
	Status        string `json:"status"`
	LastBackup    string `json:"last_backup"`
	NextBackup    string `json:"next_backup"`
	Uptime        string `json:"uptime"`
	BackupCount   int    `json:"backup_count"`
	DatabaseCount int    `json:"database_count"`
}

type Service struct {
	logger        *logger.Logger
	startTime     time.Time
	lastBackup    time.Time
	nextBackup    time.Time
	backupCount   int
	databaseCount int
	backupService BackupService
}

func NewService(logger *logger.Logger, databaseCount int) *Service {
	return &Service{
		logger:        logger,
		startTime:     time.Now(),
		databaseCount: databaseCount,
	}
}

func (s *Service) SetBackupService(backupService BackupService) {
	s.backupService = backupService
}

func (s *Service) UpdateBackupStats(lastBackup, nextBackup time.Time, count, databaseCount int) {
	s.lastBackup = lastBackup
	s.nextBackup = nextBackup
	s.backupCount = count
	s.databaseCount = databaseCount
}

func (s *Service) Start(port int) {
	http.HandleFunc("/health", s.healthHandler)
	http.HandleFunc("/status", s.statusHandler)
	http.HandleFunc("/trigger", s.triggerHandler)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	s.logger.Info("Health check server starting on port %d", port)
	if err := server.ListenAndServe(); err != nil {
		s.logger.Error("Health check server failed: %v", err)
	}
}

func (s *Service) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func (s *Service) statusHandler(w http.ResponseWriter, r *http.Request) {
	status := Status{
		Status:        "running",
		Uptime:        time.Since(s.startTime).String(),
		BackupCount:   s.backupCount,
		DatabaseCount: s.databaseCount,
	}

	if !s.lastBackup.IsZero() {
		status.LastBackup = s.lastBackup.Format("2006-01-02 15:04:05")
	} else {
		status.LastBackup = "never"
	}

	if !s.nextBackup.IsZero() {
		status.NextBackup = s.nextBackup.Format("2006-01-02 15:04:05")
	} else {
		status.NextBackup = "not scheduled"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}

func (s *Service) triggerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Only POST method is allowed",
		})
		return
	}

	if s.backupService == nil {
		s.logger.Error("Backup service not available for manual trigger")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "Backup service not available",
			"message": "The backup service is not initialized",
		})
		return
	}

	s.logger.Info("Manual backup triggered via HTTP endpoint")
	start := time.Now()

	go func() {
		dbCount, err := s.backupService.BackupAll()
		if err != nil {
			s.logger.Error("Manual backup failed: %v", err)
		} else {
			s.logger.Info("Manual backup completed successfully for %d databases", dbCount)
			s.UpdateBackupStats(time.Now(), time.Time{}, s.backupCount+1, dbCount)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status":     "accepted",
		"message":    "Backup started successfully",
		"started_at": start.Format("2006-01-02 15:04:05"),
	})
}
