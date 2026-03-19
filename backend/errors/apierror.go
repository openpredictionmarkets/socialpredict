package errors

import (
	"encoding/json"
	"log"
	"net/http"
)

// WriteJSONError writes a structured JSON error response to the client.
// Use this for errors that are safe to surface directly to the user.
func WriteJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(HTTPErrorResponse{Error: message})
}

// WriteInternalError logs the real error internally and returns a generic 500 to the client.
// Use this for unexpected server failures where internal details must not leak.
func WriteInternalError(w http.ResponseWriter, err error, logContext string) {
	if err != nil {
		log.Printf("internal error [%s]: %v", logContext, err)
	}
	WriteJSONError(w, http.StatusInternalServerError, "An internal error occurred")
}

// WriteNotFound writes a 404 response for a named resource.
func WriteNotFound(w http.ResponseWriter, resource string) {
	if resource == "" {
		resource = "resource"
	}
	WriteJSONError(w, http.StatusNotFound, resource+" not found")
}

// WriteBadRequest writes a 400 response with the provided validation message.
func WriteBadRequest(w http.ResponseWriter, message string) {
	WriteJSONError(w, http.StatusBadRequest, message)
}

// WriteUnauthorized writes a 401 response.
func WriteUnauthorized(w http.ResponseWriter) {
	WriteJSONError(w, http.StatusUnauthorized, "Authentication required")
}

// WriteForbidden writes a 403 response.
func WriteForbidden(w http.ResponseWriter) {
	WriteJSONError(w, http.StatusForbidden, "Access denied")
}
