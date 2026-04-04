package marketshandlers

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func parseBoundedInt(raw string, defaultValue, minValue, maxValue int) int {
	if raw == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value < minValue || value > maxValue {
		return defaultValue
	}

	return value
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(payload)
}
