package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database struct {
		Host      string   `yaml:"host"`
		Port      int      `yaml:"port"`
		User      string   `yaml:"user"`
		Password  string   `yaml:"password"`
		Databases []string `yaml:"databases"`
	} `yaml:"database"`

	Storage struct {
		Type  string `yaml:"type"`
		Local struct {
			Path string `yaml:"path"`
		} `yaml:"local"`
		S3 struct {
			Bucket    string `yaml:"bucket"`
			Region    string `yaml:"region"`
			Endpoint  string `yaml:"endpoint"`
			AccessKey string `yaml:"access_key"`
			SecretKey string `yaml:"secret_key"`
		} `yaml:"s3"`
	} `yaml:"storage"`

	Schedule        string `yaml:"schedule"`
	LogFile         string `yaml:"log_file"`
	RunOnStart      bool   `yaml:"run_on_start"`
	RetentionDays   int    `yaml:"retention_days"`
	HealthCheckPort int    `yaml:"health_check_port"`
}

func Load(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	setDefaults(&config)

	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

func setDefaults(config *Config) {
	if config.Database.Port == 0 {
		config.Database.Port = 5432
	}
	if config.RetentionDays == 0 {
		config.RetentionDays = 30
	}
	if config.HealthCheckPort == 0 {
		config.HealthCheckPort = 8080
	}
}

func validate(config *Config) error {
	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if config.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if config.Storage.Type == "" {
		return fmt.Errorf("storage type is required")
	}
	if config.Storage.Type == "local" && config.Storage.Local.Path == "" {
		return fmt.Errorf("local storage path is required")
	}
	if config.Storage.Type == "s3" && (config.Storage.S3.Bucket == "" || config.Storage.S3.Endpoint == "" || config.Storage.S3.AccessKey == "" || config.Storage.S3.SecretKey == "") {
		return fmt.Errorf("s3 bucket, endpoint, access key and secret key are required")
	}
	if config.Schedule == "" {
		return fmt.Errorf("schedule is required")
	}
	if config.LogFile == "" {
		return fmt.Errorf("log file is required")
	}

	return nil
}
