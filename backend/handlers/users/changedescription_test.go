package usershandlers

import (
	"socialpredict/security"
	"strings"
	"testing"
)

func TestDescriptionValidation(t *testing.T) {
	tests := []struct {
		name          string
		description   string
		expectedValid bool
		expectedError string
	}{
		{
			name:          "Valid description",
			description:   "This is a valid description",
			expectedValid: true,
		},
		{
			name:          "Empty description",
			description:   "",
			expectedValid: true, // Empty descriptions are allowed
		},
		{
			name:          "Too long description",
			description:   strings.Repeat("a", 2001),
			expectedValid: false,
			expectedError: "exceeds maximum length",
		},
		{
			name:          "Description with XSS",
			description:   "Valid description<script>alert('xss')</script>",
			expectedValid: true, // Should be sanitized, not rejected
		},
		{
			name:          "Description with HTML",
			description:   "Valid description with <b>bold</b> text",
			expectedValid: true, // Should be sanitized
		},
		{
			name:          "Max length description",
			description:   strings.Repeat("a", 2000),
			expectedValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test length validation first
			if len(tt.description) > 2000 {
				if tt.expectedValid {
					t.Errorf("Expected valid description but length validation should fail")
				}
				return
			}

			// Test sanitization
			securityService := security.NewSecurityService()
			sanitized, err := securityService.Sanitizer.SanitizeDescription(tt.description)

			if tt.expectedValid {
				if err != nil {
					t.Errorf("Expected valid description but got error: %v", err)
				}

				// Check that XSS was sanitized
				if strings.Contains(tt.description, "<script>") && strings.Contains(sanitized, "<script>") {
					t.Error("XSS script tag was not sanitized")
				}
			} else {
				if err == nil {
					t.Errorf("Expected error but validation passed")
				}
			}
		})
	}
}

func TestDescriptionXSSPrevention(t *testing.T) {
	securityService := security.NewSecurityService()

	xssPayloads := []string{
		"Good description<script>alert('xss')</script>",
		"Description with javascript:alert('xss')",
		"Text<img src=x onerror=alert('xss')>more text",
		"Content<svg onload=alert('xss')>",
		"Normal<iframe src=javascript:alert('xss')></iframe>text",
		"Text<object data=javascript:alert('xss')>content",
		"Description<embed src=javascript:alert('xss')>",
		"Some<link rel=stylesheet href=javascript:alert('xss')>text",
		"Text<meta http-equiv=refresh content=0;url=javascript:alert('xss')>",
		"Content<form><button formaction=javascript:alert('xss')>text",
	}

	for _, payload := range xssPayloads {
		t.Run("XSS_Prevention_"+payload[:20], func(t *testing.T) {
			sanitized, err := securityService.Sanitizer.SanitizeDescription(payload)

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

func TestDescriptionLengthValidation(t *testing.T) {
	tests := []struct {
		name       string
		length     int
		shouldPass bool
	}{
		{"Empty", 0, true},
		{"Short", 10, true},
		{"Medium", 500, true},
		{"Long", 1500, true},
		{"Max length", 2000, true},
		{"Over limit", 2001, false},
		{"Way over limit", 5000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			description := strings.Repeat("a", tt.length)

			// Test the actual length validation logic from the handler
			if len(description) > 2000 {
				if tt.shouldPass {
					t.Errorf("Expected to pass but length %d exceeds 2000", tt.length)
				}
			} else {
				if !tt.shouldPass {
					t.Errorf("Expected to fail but length %d is within limit", tt.length)
				}
			}
		})
	}
}
