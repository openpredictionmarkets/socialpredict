package usershandlers

import (
	"socialpredict/security"
	"strings"
	"testing"
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
