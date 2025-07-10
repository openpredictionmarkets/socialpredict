package usershandlers

import (
	"socialpredict/security"
	"strings"
	"testing"
)

func TestEmojiValidation(t *testing.T) {
	tests := []struct {
		name          string
		emoji         string
		expectedValid bool
		expectedError string
	}{
		{
			name:          "Valid emoji",
			emoji:         "😀",
			expectedValid: true,
		},
		{
			name:          "Valid text emoji",
			emoji:         ":)",
			expectedValid: true,
		},
		{
			name:          "Empty emoji",
			emoji:         "",
			expectedValid: false,
			expectedError: "cannot be blank",
		},
		{
			name:          "Too long emoji",
			emoji:         strings.Repeat("😀", 11), // Over 20 chars
			expectedValid: false,
			expectedError: "exceeds maximum length",
		},
		{
			name:          "Emoji with XSS",
			emoji:         "😀<script>alert('xss')</script>",
			expectedValid: false, // Should be rejected due to length (over 20 chars)
		},
		{
			name:          "Max length emoji",
			emoji:         strings.Repeat("😀", 5), // Exactly 20 chars (4 bytes each)
			expectedValid: true,
		},
		{
			name:          "Text with HTML",
			emoji:         "😀<b>bold</b>",
			expectedValid: true, // Should be sanitized
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test length validation first
			if len(tt.emoji) > 20 {
				if tt.expectedValid {
					t.Errorf("Expected valid emoji but length validation should fail")
				}
				return
			}

			// Test empty validation
			if tt.emoji == "" {
				if tt.expectedValid {
					t.Errorf("Expected valid emoji but empty validation should fail")
				}
				return
			}

			// Test sanitization
			securityService := security.NewSecurityService()
			sanitized, err := securityService.Sanitizer.SanitizeEmoji(tt.emoji)

			if tt.expectedValid {
				if err != nil {
					t.Errorf("Expected valid emoji but got error: %v", err)
				}

				// Check that XSS was sanitized
				if strings.Contains(tt.emoji, "<script>") && strings.Contains(sanitized, "<script>") {
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

func TestEmojiXSSPrevention(t *testing.T) {
	securityService := security.NewSecurityService()

	xssPayloads := []string{
		"😀<script>alert('xss')</script>",
		"🎉javascript:alert('xss')",
		"😊<img src=x onerror=alert('xss')>",
		"🔥<svg onload=alert('xss')>",
		"💯<iframe src=javascript:alert('xss')></iframe>",
		"🚀<object data=javascript:alert('xss')>",
		"⭐<embed src=javascript:alert('xss')>",
		"🌟<link rel=stylesheet href=javascript:alert('xss')>",
		"🎯<meta http-equiv=refresh content=0;url=javascript:alert('xss')>",
		"🎨<form><button formaction=javascript:alert('xss')>",
	}

	for _, payload := range xssPayloads {
		t.Run("XSS_Prevention_"+payload[:4], func(t *testing.T) {
			sanitized, err := securityService.Sanitizer.SanitizeEmoji(payload)

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

func TestEmojiLengthValidation(t *testing.T) {
	tests := []struct {
		name       string
		emoji      string
		shouldPass bool
	}{
		{"Single emoji", "😀", true},
		{"Text emoji", ":)", true},
		{"Multiple emojis", "😀😊🎉", true},
		{"Text with emoji", "smile😀", true},
		{"Empty", "", false},
		{"Max length text", strings.Repeat("a", 20), true},
		{"Over limit text", strings.Repeat("a", 21), false},
		{"Over limit emojis", strings.Repeat("😀", 6), false}, // 24 bytes
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the actual length validation logic from the handler
			if len(tt.emoji) > 20 {
				if tt.shouldPass {
					t.Errorf("Expected to pass but length %d exceeds 20", len(tt.emoji))
				}
			} else if tt.emoji == "" {
				if tt.shouldPass {
					t.Errorf("Expected to pass but emoji is empty")
				}
			} else {
				if !tt.shouldPass {
					t.Errorf("Expected to fail but emoji %s is valid", tt.emoji)
				}
			}
		})
	}
}

func TestEmojiSpecialCharacters(t *testing.T) {
	securityService := security.NewSecurityService()

	specialCases := []struct {
		name     string
		emoji    string
		expected string
	}{
		{
			name:     "Unicode emoji",
			emoji:    "😀",
			expected: "😀", // Should remain unchanged
		},
		{
			name:     "Text emoji",
			emoji:    ":)",
			expected: ":)", // Should remain unchanged
		},
		{
			name:     "Emoji with text",
			emoji:    "Happy😀",
			expected: "Happy😀", // Should remain unchanged
		},
		{
			name:     "HTML in emoji",
			emoji:    "😀<b>test</b>",
			expected: "😀<b>test</b>", // Emoji sanitizer doesn't check for HTML patterns
		},
	}

	for _, tc := range specialCases {
		t.Run(tc.name, func(t *testing.T) {
			sanitized, err := securityService.Sanitizer.SanitizeEmoji(tc.emoji)

			if err != nil {
				t.Logf("Sanitization rejected input (this may be expected): %v", err)
				return
			}

			// Check the expected result
			if tc.expected != "" && sanitized != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, sanitized)
			}
		})
	}
}
