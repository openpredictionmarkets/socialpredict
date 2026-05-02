package runtime

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	// ReadinessDrainWindowEnv configures how long the process waits after
	// closing readiness before HTTP shutdown begins.
	ReadinessDrainWindowEnv = "BACKEND_READINESS_DRAIN_SECONDS"

	// ShutdownTimeoutEnv configures how long the HTTP server may drain
	// in-flight requests after shutdown begins.
	ShutdownTimeoutEnv = "BACKEND_SHUTDOWN_TIMEOUT_SECONDS"

	DefaultReadinessDrainWindow = 5 * time.Second
	DefaultShutdownTimeout      = 10 * time.Second
)

// ShutdownConfig describes runtime behavior after a termination signal.
type ShutdownConfig struct {
	ReadinessDrainWindow time.Duration
	ShutdownTimeout      time.Duration
}

func LoadShutdownConfigFromEnv() (ShutdownConfig, error) {
	readinessDrainWindow, err := loadPositiveSecondsFromEnv(ReadinessDrainWindowEnv, DefaultReadinessDrainWindow)
	if err != nil {
		return ShutdownConfig{}, err
	}

	shutdownTimeout, err := loadPositiveSecondsFromEnv(ShutdownTimeoutEnv, DefaultShutdownTimeout)
	if err != nil {
		return ShutdownConfig{}, err
	}

	return ShutdownConfig{
		ReadinessDrainWindow: readinessDrainWindow,
		ShutdownTimeout:      shutdownTimeout,
	}, nil
}

func loadPositiveSecondsFromEnv(name string, defaultValue time.Duration) (time.Duration, error) {
	value := os.Getenv(name)
	if value == "" {
		return defaultValue, nil
	}
	seconds, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a positive integer number of seconds: %w", name, err)
	}
	if seconds <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer number of seconds", name)
	}

	return time.Duration(seconds) * time.Second, nil
}

func NormalizeShutdownConfig(config ShutdownConfig) ShutdownConfig {
	if config.ReadinessDrainWindow <= 0 {
		config.ReadinessDrainWindow = DefaultReadinessDrainWindow
	}
	if config.ShutdownTimeout <= 0 {
		config.ShutdownTimeout = DefaultShutdownTimeout
	}
	return config
}
