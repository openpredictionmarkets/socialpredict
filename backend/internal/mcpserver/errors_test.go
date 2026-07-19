package mcpserver

import (
	"errors"
	"testing"

	dmarkets "socialpredict/internal/domain/markets"
)

func TestMapError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code string
	}{
		{name: "validation", err: dmarkets.ErrInvalidInput, code: "validation_error"},
		{name: "market missing", err: dmarkets.ErrMarketNotFound, code: "not_found"},
		{name: "group missing", err: dmarkets.ErrMarketGroupNotFound, code: "not_found"},
		{name: "conflict", err: dmarkets.ErrInvalidState, code: "conflict"},
		{name: "internal", err: errors.New("database password leaked here"), code: "internal_error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapError(tt.err)
			if got.Code != tt.code {
				t.Fatalf("code = %q, want %q", got.Code, tt.code)
			}
			if got.Code == "internal_error" && got.Message != "internal server error" {
				t.Fatalf("internal message = %q", got.Message)
			}
		})
	}
}
