package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port           string
	DefaultCount   int
	DefaultTimeout time.Duration
	MaxCount       int
	MaxTimeout     time.Duration
	EnableLogging  bool
	
	// PocketBase configuration
	PocketBaseEnabled   bool
	PocketBaseURL       string
	PocketBaseAdminEmail    string
	PocketBaseAdminPassword string
}

func Load() *Config {
	cfg := &Config{
		Port:           getEnv("PORT", "8091"),
		DefaultCount:   getEnvInt("DEFAULT_COUNT", 4),
		DefaultTimeout: getEnvDuration("DEFAULT_TIMEOUT", 15*time.Second),
		MaxCount:       getEnvInt("MAX_COUNT", 20),
		MaxTimeout:     getEnvDuration("MAX_TIMEOUT", 30*time.Second),
		EnableLogging:  getEnvBool("ENABLE_LOGGING", true),

		PocketBaseEnabled:       getEnvBool("POCKETBASE_ENABLED", true),
		PocketBaseURL:           getEnv("POCKETBASE_URL", ""),
		PocketBaseAdminEmail:    getEnv("POCKETBASE_ADMIN_EMAIL", ""),
		PocketBaseAdminPassword: getEnv("POCKETBASE_ADMIN_PASSWORD", ""),
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
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

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}