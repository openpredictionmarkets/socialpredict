package authhttp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
		{"nil error", nil, http.StatusInternalServerError, handlers.ReasonInternalError},
		{"missing token", &authsvc.AuthError{Kind: authsvc.ErrorKindMissingToken}, http.StatusUnauthorized, handlers.ReasonInvalidToken},
		{"invalid token", &authsvc.AuthError{Kind: authsvc.ErrorKindInvalidToken}, http.StatusUnauthorized, handlers.ReasonInvalidToken},
		{"password change required", &authsvc.AuthError{Kind: authsvc.ErrorKindPasswordChangeRequired}, http.StatusForbidden, handlers.ReasonPasswordChangeRequired},
		{"authorization denied", &authsvc.AuthError{Kind: authsvc.ErrorKindAdminRequired}, http.StatusForbidden, handlers.ReasonAuthorizationDenied},
		{"user not found", &authsvc.AuthError{Kind: authsvc.ErrorKindUserNotFound}, http.StatusNotFound, handlers.ReasonUserNotFound},
		{"internal", &authsvc.AuthError{Kind: authsvc.ErrorKindUserLoadFailed}, http.StatusInternalServerError, handlers.ReasonInternalError},
		{"service unavailable", &authsvc.AuthError{Kind: authsvc.ErrorKindServiceUnavailable}, http.StatusInternalServerError, handlers.ReasonInternalError},
		{"unknown", &authsvc.AuthError{Kind: authsvc.ErrorKind("new_kind")}, http.StatusInternalServerError, handlers.ReasonInternalError},
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

func TestAuthFailureWriteEnvelope(t *testing.T) {
	tests := []struct {
		name       string
		failure    *Failure
		wantStatus int
		wantReason handlers.FailureReason
	}{
		{
			name:       "nil failure writes internal envelope",
			failure:    nil,
			wantStatus: http.StatusInternalServerError,
			wantReason: handlers.ReasonInternalError,
		},
		{
			name: "auth failure writes mapped envelope",
			failure: &Failure{
				StatusCode: http.StatusForbidden,
				Reason:     handlers.ReasonPasswordChangeRequired,
				Message:    "Password change required",
			},
			wantStatus: http.StatusForbidden,
			wantReason: handlers.ReasonPasswordChangeRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			if err := tt.failure.Write(rec); err != nil {
				t.Fatalf("Write returned error: %v", err)
			}
			assertFailureEnvelope(t, rec, tt.wantStatus, tt.wantReason)
		})
	}
}

func TestWriteFailureEnvelope(t *testing.T) {
	rec := httptest.NewRecorder()
	err := WriteFailure(rec, &authsvc.AuthError{
		Kind:    authsvc.ErrorKindInvalidToken,
		Message: "Invalid token",
	})
	if err != nil {
		t.Fatalf("WriteFailure returned error: %v", err)
	}
	assertFailureEnvelope(t, rec, http.StatusUnauthorized, handlers.ReasonInvalidToken)
}

func TestFailureErrorMessage(t *testing.T) {
	if got := (*Failure)(nil).Error(); got != "" {
		t.Fatalf("nil Failure.Error() = %q, want empty", got)
	}

	failure := FailureFromAuthError(&authsvc.AuthError{
		Kind:    authsvc.ErrorKindAdminRequired,
		Message: "admin only",
	})
	if got := failure.Error(); got != "admin only" {
		t.Fatalf("Failure.Error() = %q, want custom message", got)
	}

	defaulted := FailureFromAuthError(&authsvc.AuthError{Kind: authsvc.ErrorKindMissingToken})
	if got := defaulted.Error(); got != "authentication failed" {
		t.Fatalf("default Failure.Error() = %q, want authentication failed", got)
	}
}

func assertFailureEnvelope(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, wantReason handlers.FailureReason) {
	t.Helper()

	if rec.Code != wantStatus {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, wantStatus, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}

	var response handlers.FailureEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode failure envelope: %v", err)
	}
	if response.OK || response.Reason != string(wantReason) {
		t.Fatalf("response = %+v, want reason %q", response, wantReason)
	}
}
