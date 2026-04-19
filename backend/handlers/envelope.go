package handlers

import (
	"encoding/json"
	"net/http"
)

type FailureReason string

const (
	ReasonMethodNotAllowed       FailureReason = "METHOD_NOT_ALLOWED"
	ReasonInvalidRequest         FailureReason = "INVALID_REQUEST"
	ReasonInvalidToken           FailureReason = "INVALID_TOKEN"
	ReasonPasswordChangeRequired FailureReason = "PASSWORD_CHANGE_REQUIRED"
	ReasonUserNotFound           FailureReason = "USER_NOT_FOUND"

	ReasonValidationFailed FailureReason = "VALIDATION_FAILED"
	ReasonInternalError    FailureReason = "INTERNAL_ERROR"
)

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
