package runtime

import (
	"fmt"
	"os"
	"strings"
)

const (
	// StartupWriterEnv enables the one process role allowed to perform startup-owned DB writes.
	StartupWriterEnv = "STARTUP_WRITER"

	legacyStartupWriterEnv = "SP_STARTUP_WRITER"
)

// StartupMutationMode identifies whether this process may run startup-owned DB writes.
type StartupMutationMode struct {
	Writer bool
	Source string
}

// LoadStartupMutationModeFromEnv reads the explicit startup writer seam.
func LoadStartupMutationModeFromEnv() (StartupMutationMode, error) {
	for _, key := range []string{StartupWriterEnv, legacyStartupWriterEnv} {
		value, ok := os.LookupEnv(key)
		if !ok {
			continue
		}

		writer, err := parseStartupWriterFlag(value)
		if err != nil {
			return StartupMutationMode{}, fmt.Errorf("%s: %w", key, err)
		}
		return StartupMutationMode{Writer: writer, Source: key}, nil
	}

	return StartupMutationMode{Writer: false, Source: "default"}, nil
}

func parseStartupWriterFlag(value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true, nil
	case "0", "false", "no", "off":
		return false, nil
	default:
		return false, fmt.Errorf("expected boolean startup writer flag, got %q", value)
	}
}
