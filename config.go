package main

import (
	"os"
	"time"
)

// AppConfig holds runtime configuration.
type AppConfig struct {
	Port   string
	DBPath string
	// Server timeouts
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func loadConfig() AppConfig {
	return AppConfig{
		Port:              getenv("PORT", "8086"),
		DBPath:            getenv("DB_PATH", dbFile),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}
