package mcpserver

import (
	"errors"

	dmarkets "socialpredict/internal/domain/markets"
)

type ToolError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *ToolError) Error() string {
	if e == nil {
		return ""
	}
	if e.Code == "" {
		return e.Message
	}
	return e.Code + ": " + e.Message
}

func MapError(err error) *ToolError {
	if err == nil {
		return nil
	}
	if toolErr := new(ToolError); errors.As(err, &toolErr) {
		return toolErr
	}
	switch {
	case errors.Is(err, dmarkets.ErrInvalidInput):
		return &ToolError{Code: "validation_error", Message: "invalid input"}
	case errors.Is(err, dmarkets.ErrMarketNotFound), errors.Is(err, dmarkets.ErrMarketGroupNotFound):
		return &ToolError{Code: "not_found", Message: err.Error()}
	case errors.Is(err, dmarkets.ErrInvalidState):
		return &ToolError{Code: "conflict", Message: "requested action is not valid for the current state"}
	case errors.Is(err, dmarkets.ErrUnauthorized):
		return &ToolError{Code: "unauthorized", Message: "authentication required"}
	default:
		return &ToolError{Code: "internal_error", Message: "internal server error"}
	}
}
