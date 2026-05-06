package security

import "strconv"

// ParseBoundedIntParam parses an integer HTTP boundary value with a default and inclusive bounds.
func ParseBoundedIntParam(raw string, defaultValue, minValue, maxValue int) int {
	if raw == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value < minValue || value > maxValue {
		return defaultValue
	}

	return value
}
