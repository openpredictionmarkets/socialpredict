package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
)

type FailureReason string

const (
	ReasonMethodNotAllowed       FailureReason = "METHOD_NOT_ALLOWED"
	ReasonInvalidRequest         FailureReason = "INVALID_REQUEST"
	ReasonInvalidToken           FailureReason = "INVALID_TOKEN"
	ReasonAuthorizationDenied    FailureReason = "AUTHORIZATION_DENIED"
	ReasonPasswordChangeRequired FailureReason = "PASSWORD_CHANGE_REQUIRED"
	ReasonNotFound               FailureReason = "NOT_FOUND"
	ReasonUserNotFound           FailureReason = "USER_NOT_FOUND"
	ReasonMarketNotFound         FailureReason = "MARKET_NOT_FOUND"

	ReasonValidationFailed    FailureReason = "VALIDATION_FAILED"
	ReasonMarketClosed        FailureReason = "MARKET_CLOSED"
	ReasonInsufficientBalance FailureReason = "INSUFFICIENT_BALANCE"
	ReasonNoPosition          FailureReason = "NO_POSITION"
	ReasonInsufficientShares  FailureReason = "INSUFFICIENT_SHARES"
	ReasonDustCapExceeded     FailureReason = "DUST_CAP_EXCEEDED"
	ReasonInternalError       FailureReason = "INTERNAL_ERROR"
)

var publicFailureReasons = []FailureReason{
	ReasonMethodNotAllowed,
	ReasonInvalidRequest,
	ReasonInvalidToken,
	ReasonAuthorizationDenied,
	ReasonPasswordChangeRequired,
	ReasonNotFound,
	ReasonUserNotFound,
	ReasonMarketNotFound,
	ReasonValidationFailed,
	ReasonMarketClosed,
	ReasonInsufficientBalance,
	ReasonNoPosition,
	ReasonInsufficientShares,
	ReasonDustCapExceeded,
	ReasonInternalError,
}

// PublicFailureReasons returns the shared public reason vocabulary for the
// touched API envelope routes.
func PublicFailureReasons() []FailureReason {
	return append([]FailureReason(nil), publicFailureReasons...)
}

type SuccessEnvelope[T any] struct {
	OK     bool `json:"ok"`
	Result T    `json:"result"`
}

type FailureEnvelope struct {
	OK     bool   `json:"ok"`
	Reason string `json:"reason"`
}

func WriteResult[T any](w http.ResponseWriter, statusCode int, result T) error {
	return writeJSON(w, statusCode, SuccessEnvelope[T]{
		OK:     true,
		Result: result,
	})
}

func WriteFailure(w http.ResponseWriter, statusCode int, reason FailureReason) error {
	return writeJSON(w, statusCode, FailureEnvelope{
		OK:     false,
		Reason: string(reason),
	})
}

func WriteBusinessFailure(w http.ResponseWriter, reason FailureReason) error {
	return WriteFailure(w, http.StatusOK, reason)
}

// AuthFailureReason maps auth-layer HTTP outcomes to the shared public reason
// vocabulary used by touched envelope handlers.
func AuthFailureReason(statusCode int, message string) FailureReason {
	switch {
	case statusCode >= http.StatusInternalServerError:
		return ReasonInternalError
	case statusCode == http.StatusUnauthorized:
		return ReasonInvalidToken
	case statusCode == http.StatusForbidden && strings.EqualFold(message, "Password change required"):
		return ReasonPasswordChangeRequired
	case statusCode == http.StatusForbidden:
		return ReasonAuthorizationDenied
	case statusCode == http.StatusNotFound:
		return ReasonUserNotFound
	case statusCode == http.StatusBadRequest:
		return ReasonValidationFailed
	default:
		return ReasonInternalError
	}
}

func IsValidationMessage(message string) bool {
	lower := strings.ToLower(message)
	return strings.Contains(lower, "invalid") ||
		strings.Contains(lower, "exceeds") ||
		strings.Contains(lower, "must") ||
		strings.Contains(lower, "cannot") ||
		strings.Contains(lower, "required")
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, err = w.Write(append(body, '\n'))
	return err
}
