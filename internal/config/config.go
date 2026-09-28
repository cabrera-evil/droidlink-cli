// Package config centralizes droidlink's runtime defaults. It is env-var
// based, not file based: values here only seed command flag defaults, which
// always take precedence when explicitly set.
package config

import (
	"os"
	"strconv"
	"time"
)

const (
	defaultMetroPort        = 8081
	defaultDiscoveryTimeout = 5 * time.Second
	defaultLogLevel         = "info"
)

// Config holds runtime defaults sourced from the environment.
type Config struct {
	DefaultMetroPort        int
	DefaultDiscoveryTimeout time.Duration
	DefaultLogLevel         string
}

// Load reads runtime defaults from the environment, falling back to
// built-in defaults for anything unset or invalid.
func Load() Config {
	return Config{
		DefaultMetroPort:        intEnv("DROIDLINK_METRO_PORT", defaultMetroPort),
		DefaultDiscoveryTimeout: durationEnv("DROIDLINK_DISCOVERY_TIMEOUT", defaultDiscoveryTimeout),
		DefaultLogLevel:         stringEnv("DROIDLINK_LOG_LEVEL", defaultLogLevel),
	}
}

func stringEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

func intEnv(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}

	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}

	return n
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}

	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}

	return d
}
