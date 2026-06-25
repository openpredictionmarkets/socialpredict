package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
)

type FailureReason string

const (
	ReasonMethodNotAllowed            FailureReason = "METHOD_NOT_ALLOWED"
	ReasonInvalidRequest              FailureReason = "INVALID_REQUEST"
	ReasonInvalidToken                FailureReason = "INVALID_TOKEN"
	ReasonAuthorizationDenied         FailureReason = "AUTHORIZATION_DENIED"
	ReasonUserNotApproved             FailureReason = "USER_NOT_APPROVED"
	ReasonPasswordChangeRequired      FailureReason = "PASSWORD_CHANGE_REQUIRED"
	ReasonNotFound                    FailureReason = "NOT_FOUND"
	ReasonRateLimited                 FailureReason = "RATE_LIMITED"
	ReasonLoginRateLimited            FailureReason = "LOGIN_RATE_LIMITED"
	ReasonUserNotFound                FailureReason = "USER_NOT_FOUND"
	ReasonMarketNotFound              FailureReason = "MARKET_NOT_FOUND"
	ReasonInvalidState                FailureReason = "INVALID_STATE"
	ReasonMarketGroupChildUnpublished FailureReason = "MARKET_GROUP_CHILD_UNPUBLISHED"

	ReasonValidationFailed    FailureReason = "VALIDATION_FAILED"
	ReasonMarketClosed        FailureReason = "MARKET_CLOSED"
	ReasonInsufficientBalance FailureReason = "INSUFFICIENT_BALANCE"
	ReasonNoPosition          FailureReason = "NO_POSITION"
	ReasonInsufficientShares  FailureReason = "INSUFFICIENT_SHARES"
	ReasonPositionLocked      FailureReason = "POSITION_LOCKED_AWAITING_EXTERNAL_MARKET_MOVEMENT"
	ReasonDustCapExceeded     FailureReason = "DUST_CAP_EXCEEDED"
	ReasonInternalError       FailureReason = "INTERNAL_ERROR"
)

var publicFailureReasons = []FailureReason{
	ReasonMethodNotAllowed,
	ReasonInvalidRequest,
	ReasonInvalidToken,
	ReasonAuthorizationDenied,
	ReasonUserNotApproved,
	ReasonPasswordChangeRequired,
	ReasonNotFound,
	ReasonRateLimited,
	ReasonLoginRateLimited,
	ReasonUserNotFound,
	ReasonMarketNotFound,
	ReasonInvalidState,
	ReasonMarketGroupChildUnpublished,
	ReasonValidationFailed,
	ReasonMarketClosed,
	ReasonInsufficientBalance,
	ReasonNoPosition,
	ReasonInsufficientShares,
	ReasonPositionLocked,
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
	OK      bool           `json:"ok"`
	Reason  string         `json:"reason"`
	Message string         `json:"message,omitempty"`
	Details map[string]any `json:"details,omitempty"`
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

func WriteFailureWithDetails(w http.ResponseWriter, statusCode int, reason FailureReason, message string, details map[string]any) error {
	return writeJSON(w, statusCode, FailureEnvelope{
		OK:      false,
		Reason:  string(reason),
		Message: strings.TrimSpace(message),
		Details: details,
	})
}

func WriteBusinessFailure(w http.ResponseWriter, reason FailureReason) error {
	return WriteFailure(w, http.StatusOK, reason)
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
