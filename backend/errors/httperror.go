package errors

import (
	"encoding/json"
	"log"
	"net/http"
)

// HTTPErrorResponse represents a structured error response.
type HTTPErrorResponse struct {
	Error string `json:"error"`
}

// HandleHTTPError checks for an error and handles it by sending an appropriate HTTP response.
// It logs the error and writes a corresponding HTTP status code and error message to the response writer.
// Returns true if there was an error handled, false otherwise.
func HandleHTTPError(w http.ResponseWriter, err error, statusCode int, userMessage string) bool {
	if err != nil {
		log.Printf("Error: %v", err) // Log the actual error for server-side diagnostics.
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(HTTPErrorResponse{Error: userMessage})
		return true
	}
	return false
}
