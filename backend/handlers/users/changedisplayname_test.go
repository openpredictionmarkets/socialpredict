package usershandlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"socialpredict/security"
	"strings"
	"testing"

	"socialpredict/handlers"
	"socialpredict/handlers/users/dto"
	dusers "socialpredict/internal/domain/users"
	"socialpredict/models/modelstesting"
)

func TestDisplayNameValidation(t *testing.T) {
	tests := []struct {
		name          string
		displayName   string
		expectedValid bool
		expectedError string
	}{
		{
			name:          "Valid display name",
			displayName:   "Valid Name",
			expectedValid: true,
		},
		{
			name:          "Empty display name",
			displayName:   "",
			expectedValid: false,
			expectedError: "must be between 1 and 50 characters",
		},
		{
			name:          "Too long display name",
			displayName:   strings.Repeat("a", 51),
			expectedValid: false,
			expectedError: "must be between 1 and 50 characters",
		},
		{
			name:          "Display name with XSS",
			displayName:   "Test<script>alert('xss')</script>",
			expectedValid: false, // Should be rejected due to containsSuspiciousPatterns
		},
		{
			name:          "Display name with HTML",
			displayName:   "Test<b>bold</b>",
			expectedValid: true, // Basic HTML tags like <b> are allowed by strict policy after sanitization
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test length validation first
			if len(tt.displayName) > 50 || len(tt.displayName) < 1 {
				if tt.expectedValid {
					t.Errorf("Expected valid display name but length validation failed")
				}
				return
			}

			// Test sanitization
			securityService := security.NewSecurityService()
			sanitized, err := securityService.Sanitizer.SanitizeDisplayName(tt.displayName)

			if tt.expectedValid {
				if err != nil {
					t.Errorf("Expected valid display name but got error: %v", err)
				}

				// Check that XSS was sanitized
				if strings.Contains(tt.displayName, "<script>") && strings.Contains(sanitized, "<script>") {
					t.Error("XSS script tag was not sanitized")
				}
			} else {
				if err == nil && !strings.Contains(tt.expectedError, "characters") {
					t.Errorf("Expected error but validation passed")
				}
			}
		})
	}
}

func TestDisplayNameXSSPrevention(t *testing.T) {
	securityService := security.NewSecurityService()

	xssPayloads := []string{
		"<script>alert('xss')</script>",
		"javascript:alert('xss')",
		"<img src=x onerror=alert('xss')>",
		"<svg onload=alert('xss')>",
		"<iframe src=javascript:alert('xss')></iframe>",
		"<object data=javascript:alert('xss')>",
		"<embed src=javascript:alert('xss')>",
		"<link rel=stylesheet href=javascript:alert('xss')>",
		"<meta http-equiv=refresh content=0;url=javascript:alert('xss')>",
		"<form><button formaction=javascript:alert('xss')>",
	}

	for _, payload := range xssPayloads {
		t.Run("XSS_Prevention_"+payload, func(t *testing.T) {
			sanitized, err := securityService.Sanitizer.SanitizeDisplayName(payload)

			if err != nil {
				// Some payloads might be completely rejected, which is fine
				return
			}

			// Verify the dangerous content was sanitized
			if sanitized == payload {
				t.Errorf("Dangerous payload was not sanitized: %s", payload)
			}

			// Verify no script tags remain
			if strings.Contains(sanitized, "<script>") {
				t.Errorf("Script tag remained after sanitization: %s -> %s", payload, sanitized)
			}

			// Verify no javascript: protocols remain
			if strings.Contains(sanitized, "javascript:") {
				t.Errorf("JavaScript protocol remained after sanitization: %s -> %s", payload, sanitized)
			}
		})
	}
}

type displayNameServiceMock struct {
	user    *dusers.User
	updated *dusers.User
	err     error
}

func (m *displayNameServiceMock) GetPublicUser(context.Context, string) (*dusers.PublicUser, error) {
	return nil, nil
}

func (m *displayNameServiceMock) GetUser(context.Context, string) (*dusers.User, error) {
	return m.user, nil
}

func (m *displayNameServiceMock) GetPrivateProfile(context.Context, string) (*dusers.PrivateProfile, error) {
	return nil, nil
}

func (m *displayNameServiceMock) ApplyTransaction(context.Context, string, int64, string) error {
	return nil
}

func (m *displayNameServiceMock) GetUserCredit(context.Context, string, int64) (int64, error) {
	return 0, nil
}

func (m *displayNameServiceMock) GetUserPortfolio(context.Context, string) (*dusers.Portfolio, error) {
	return nil, nil
}

func (m *displayNameServiceMock) GetUserFinancials(context.Context, string) (map[string]int64, error) {
	return nil, nil
}

func (m *displayNameServiceMock) ListUserMarkets(context.Context, int64) ([]*dusers.UserMarket, error) {
	return nil, nil
}

func (m *displayNameServiceMock) UpdateDescription(context.Context, string, string) (*dusers.User, error) {
	return nil, nil
}

func (m *displayNameServiceMock) UpdateDisplayName(context.Context, string, string) (*dusers.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.updated, nil
}

func (m *displayNameServiceMock) UpdateEmoji(context.Context, string, string) (*dusers.User, error) {
	return nil, nil
}

func (m *displayNameServiceMock) UpdatePersonalLinks(context.Context, string, dusers.PersonalLinks) (*dusers.User, error) {
	return nil, nil
}

func (m *displayNameServiceMock) ChangePassword(context.Context, string, string, string) error {
	return nil
}

func TestChangeDisplayNameHandler_SuccessEnvelope(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")

	svc := &displayNameServiceMock{
		user: &dusers.User{Username: "alice"},
		updated: &dusers.User{
			ID: 1, Username: "alice", DisplayName: "Alice New", UserType: "REGULAR",
		},
	}
	payload, _ := json.Marshal(dto.ChangeDisplayNameRequest{DisplayName: "Alice New"})
	req := httptest.NewRequest(http.MethodPost, "/v0/profilechange/displayname", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+modelstesting.GenerateValidJWT("alice"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	ChangeDisplayNameHandler(svc).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp handlers.SuccessEnvelope[dto.PrivateUserResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode success envelope: %v", err)
	}
	if !resp.OK {
		t.Fatalf("expected ok=true, got false")
	}
	if resp.Result.DisplayName != "Alice New" {
		t.Fatalf("expected display name Alice New, got %q", resp.Result.DisplayName)
	}
}

func TestWriteProfileError_SanitizesValidationFailure(t *testing.T) {
	rec := httptest.NewRecorder()

	writeProfileError(rec, errors.New("display name must be between 1 and 50 characters"), "display name")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	var resp handlers.FailureEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode failure envelope: %v", err)
	}
	if resp.Reason != string(handlers.ReasonValidationFailed) {
		t.Fatalf("expected reason %q, got %q", handlers.ReasonValidationFailed, resp.Reason)
	}
}
