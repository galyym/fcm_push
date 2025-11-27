package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Server ServerConfig
	FCM    FCMConfig
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
	}

	if cfg.FCM.CredentialsPath == "" {
		return nil, fmt.Errorf("FCM_CREDENTIALS_PATH is required")
	}
	if cfg.FCM.ProjectID == "" {
		return nil, fmt.Errorf("FCM_PROJECT_ID is required")
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
