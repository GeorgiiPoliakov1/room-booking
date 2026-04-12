package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port     string
	Env      string
	LogLevel string

	DBHost            string
	DBPort            int
	DBUser            string
	DBPassword        string
	DBName            string
	DBSSLMode         string
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:     getEnv("APP_PORT", "8080"),
		Env:      getEnv("APP_ENV", "prod"),
		LogLevel: getEnv("LOG_LEVEL", "info"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getIntEnv("DB_PORT", 5432),
		DBUser:     getEnv("DB_USER", "roomuser"),
		DBPassword: getEnv("DB_PASSWORD", "roompass"),
		DBName:     getEnv("DB_NAME", "roomdb"),
		DBSSLMode:  getEnv("DB_SSL_MODE", "require"),

		DBMaxOpenConns:    getIntEnv("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:    getIntEnv("DB_MAX_IDLE_CONNS", 5),
		DBConnMaxLifetime: getDurationEnv("DB_CONN_MAX_LIFETIME", 30*time.Minute),
	}
	if cfg.DBPassword == "" && cfg.Env == "production" {
		return nil, fmt.Errorf("DB_PASSWORD is required in production environment")
	}
	if cfg.DBName == "" {
		return nil, fmt.Errorf("DB_NAME is required")
	}

	if cfg.Env == "production" && cfg.DBSSLMode == "disable" {
		return nil, fmt.Errorf("DB_SSL_MODE=disable is not allowed in production")
	}

	return cfg, nil
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
		c.DBSSLMode,
	)
}

func (c *Config) IsDevelopment() bool {
	return c.Env == "development" || c.Env == "dev"
}

func getIntEnv(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intVal, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intVal
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return duration
}
