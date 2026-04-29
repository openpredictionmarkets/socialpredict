package authhttp

import (
	"net/http"
	"testing"

	"socialpredict/handlers"
	authsvc "socialpredict/internal/service/auth"
)

func TestAuthFailureMapping(t *testing.T) {
	tests := []struct {
		name       string
		err        *authsvc.AuthError
		wantStatus int
		wantReason handlers.FailureReason
	}{
		{"missing token", &authsvc.AuthError{Kind: authsvc.ErrorKindMissingToken}, http.StatusUnauthorized, handlers.ReasonInvalidToken},
		{"password change required", &authsvc.AuthError{Kind: authsvc.ErrorKindPasswordChangeRequired}, http.StatusForbidden, handlers.ReasonPasswordChangeRequired},
		{"authorization denied", &authsvc.AuthError{Kind: authsvc.ErrorKindAdminRequired}, http.StatusForbidden, handlers.ReasonAuthorizationDenied},
		{"user not found", &authsvc.AuthError{Kind: authsvc.ErrorKindUserNotFound}, http.StatusNotFound, handlers.ReasonUserNotFound},
		{"internal", &authsvc.AuthError{Kind: authsvc.ErrorKindUserLoadFailed}, http.StatusInternalServerError, handlers.ReasonInternalError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StatusCode(tt.err); got != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, got)
			}
			if got := FailureReason(tt.err); got != tt.wantReason {
				t.Fatalf("expected reason %q, got %q", tt.wantReason, got)
			}
		})
	}
}
