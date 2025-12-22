package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port string
	MaxFileSize int64
	ProcessingTimeout int
	MaxMemoryMB int
}

func Load() *Config {
	return &Config{
		Port: getEnv("GO_PORT", "8080"),
		MaxFileSize: getEnvInt64("MAX_FILE_SIZE", 10485760), // 10MB
		ProcessingTimeout: getEnvInt("PROCESSING_TIMEOUT", 30),
		MaxMemoryMB: getEnvInt("MAX_MEMORY_MB", 512),
	}
}

func getEnv(key, defaultValue string) string{
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}
