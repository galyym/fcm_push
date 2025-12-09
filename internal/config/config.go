package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	FCM      FCMConfig
	Database DatabaseConfig
	Worker   WorkerConfig
}
type ServerConfig struct {
	Port         string
	ReadTimeout  int
	WriteTimeout int
}
type FCMConfig struct {
	CredentialsPath string
	ProjectID       string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type WorkerConfig struct {
	WorkerCount      int
	PollInterval     string
	MaxRetryAttempts int
	RetryIntervals   string
	CleanupAfterDays int
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			ReadTimeout:  getEnvAsInt("SERVER_READ_TIMEOUT", 10),
			WriteTimeout: getEnvAsInt("SERVER_WRITE_TIMEOUT", 10),
		},
		FCM: FCMConfig{
			CredentialsPath: getEnv("FCM_CREDENTIALS_PATH", ""),
			ProjectID:       getEnv("FCM_PROJECT_ID", ""),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "fcm_push_db"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Worker: WorkerConfig{
			WorkerCount:      getEnvAsInt("WORKER_COUNT", 5),
			PollInterval:     getEnv("WORKER_POLL_INTERVAL", "5s"),
			MaxRetryAttempts: getEnvAsInt("MAX_RETRY_ATTEMPTS", 3),
			RetryIntervals:   getEnv("RETRY_INTERVALS", "1m,5m,15m"),
			CleanupAfterDays: getEnvAsInt("CLEANUP_AFTER_DAYS", 30),
		},
	}

	if cfg.FCM.CredentialsPath == "" {
		return nil, fmt.Errorf("FCM_CREDENTIALS_PATH is required")
	}
	if cfg.FCM.ProjectID == "" {
		return nil, fmt.Errorf("FCM_PROJECT_ID is required")
	}
	if cfg.Database.Password == "" {
		return nil, fmt.Errorf("DB_PASSWORD is required")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}
