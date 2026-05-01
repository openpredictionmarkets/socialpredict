package security

import (
	"encoding/json"
	"net/http"
)

const (
	RuntimeReasonInternalError    = "INTERNAL_ERROR"
	RuntimeReasonLoginRateLimited = "LOGIN_RATE_LIMITED"
	RuntimeReasonMethodNotAllowed = "METHOD_NOT_ALLOWED"
	RuntimeReasonRateLimited      = "RATE_LIMITED"

	RuntimeErrorTypeInternal         = "runtime.internal"
	RuntimeErrorTypeMethodNotAllowed = "http.method_not_allowed"
	RuntimeErrorTypePanic            = "exception.panic"
	RuntimeErrorTypeRateLimited      = "http.rate_limited"
)

// RuntimeFailureResponse is the public runtime-boundary failure envelope.
type RuntimeFailureResponse struct {
	OK     bool   `json:"ok"`
	Reason string `json:"reason"`
}

// RuntimeFailureErrorType returns the stable telemetry classification for
// runtime-owned HTTP failures. It is intentionally independent from public
// response reason strings.
func RuntimeFailureErrorType(statusCode int) (string, bool) {
	switch statusCode {
	case http.StatusMethodNotAllowed:
		return RuntimeErrorTypeMethodNotAllowed, true
	case http.StatusTooManyRequests:
		return RuntimeErrorTypeRateLimited, true
	case http.StatusInternalServerError:
		return RuntimeErrorTypeInternal, true
	default:
		return "", false
	}
}

// WriteMethodNotAllowed writes the shared runtime-owned 405 response.
func WriteMethodNotAllowed(w http.ResponseWriter) {
	writeRuntimeFailure(w, http.StatusMethodNotAllowed, RuntimeReasonMethodNotAllowed)
}

// WriteRateLimited writes the shared runtime-owned 429 response.
func WriteRateLimited(w http.ResponseWriter, reason string) {
	if reason == "" {
		reason = RuntimeReasonRateLimited
	}
	writeRuntimeFailure(w, http.StatusTooManyRequests, reason)
}

// WriteInternalServerError writes the shared sanitized runtime-owned 500 response.
func WriteInternalServerError(w http.ResponseWriter) {
	writeRuntimeFailure(w, http.StatusInternalServerError, RuntimeReasonInternalError)
}

func writeRuntimeFailure(w http.ResponseWriter, statusCode int, reason string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(RuntimeFailureResponse{
		OK:     false,
		Reason: reason,
	})
}
