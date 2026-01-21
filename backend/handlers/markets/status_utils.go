package marketshandlers

import (
	"fmt"
	"strings"
)

// normalizeStatusParam converts arbitrary status input to the canonical value understood by the domain layer.
// Returns "" when no filter should be applied (empty/all), otherwise one of active|closed|resolved.
func normalizeStatusParam(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", nil
	}

	value = strings.ToLower(value)
	switch value {
	case "active", "closed", "resolved":
		return value, nil
	case "all":
		return "", nil
	default:
		return "", fmt.Errorf("invalid status %q", raw)
	}
}
