package health

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"pg-backup/internal/logger"
)

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
}

func NewService(logger *logger.Logger, databaseCount int) *Service {
	return &Service{
		logger:        logger,
		startTime:     time.Now(),
		databaseCount: databaseCount,
	}
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
