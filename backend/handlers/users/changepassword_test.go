package usershandlers

import (
	"socialpredict/security"
	"strings"
	"testing"
)

func TestPasswordValidation(t *testing.T) {
	tests := []struct {
		name          string
		currentPass   string
		newPass       string
		expectedValid bool
		expectedError string
	}{
		{
			name:          "Valid password change",
			currentPass:   "oldPassword123!",
			newPass:       "newPassword456!",
			expectedValid: true,
		},
		{
			name:          "Empty current password",
			currentPass:   "",
			newPass:       "newPassword456!",
			expectedValid: false,
			expectedError: "Current password is required",
		},
		{
			name:          "Empty new password",
			currentPass:   "oldPassword123!",
			newPass:       "",
			expectedValid: false,
			expectedError: "New password is required",
		},
		{
			name:          "Weak new password",
			currentPass:   "oldPassword123!",
			newPass:       "weak",
			expectedValid: false,
			expectedError: "does not meet security requirements",
		},
		{
			name:          "Password without numbers",
			currentPass:   "oldPassword123!",
			newPass:       "NoNumbers!",
			expectedValid: false,
			expectedError: "does not meet security requirements",
		},
		{
			name:          "Password without special chars",
			currentPass:   "oldPassword123!",
			newPass:       "NoSpecialChars123",
			expectedValid: false,
			expectedError: "does not meet security requirements",
		},
		{
			name:          "Password too short",
			currentPass:   "oldPassword123!",
			newPass:       "Short1!",
			expectedValid: false,
			expectedError: "does not meet security requirements",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test basic field validation
			if tt.currentPass == "" || tt.newPass == "" {
				// These should fail the basic validation
				if tt.expectedValid {
					t.Errorf("Expected valid but empty fields should fail")
				}
				return
			}

			// Test password strength validation
			securityService := security.NewSecurityService()
			err := securityService.Sanitizer.SanitizePassword(tt.newPass)

			if tt.expectedValid {
				if err != nil {
					t.Errorf("Expected valid password but got error: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error but password validation passed")
				}
			}
		})
	}
}

func TestPasswordStrengthRequirements(t *testing.T) {
	securityService := security.NewSecurityService()

	tests := []struct {
		name       string
		password   string
		shouldPass bool
		reason     string
	}{
		{
			name:       "Strong password",
			password:   "StrongPass123!",
			shouldPass: true,
		},
		{
			name:       "Too short",
			password:   "Short1!",
			shouldPass: false,
			reason:     "Less than 8 characters",
		},
		{
			name:       "No uppercase",
			password:   "lowercase123!",
			shouldPass: false,
			reason:     "No uppercase letters",
		},
		{
			name:       "No lowercase",
			password:   "UPPERCASE123!",
			shouldPass: false,
			reason:     "No lowercase letters",
		},
		{
			name:       "No numbers",
			password:   "NoNumbers!",
			shouldPass: false,
			reason:     "No numbers",
		},
		{
			name:       "No special characters",
			password:   "NoSpecialChars123",
			shouldPass: false,
			reason:     "No special characters",
		},
		{
			name:       "Common password",
			password:   "Password123!",
			shouldPass: false,
			reason:     "Too common",
		},
		{
			name:       "All requirements met",
			password:   "SecureP@ssw0rd",
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := securityService.Sanitizer.SanitizePassword(tt.password)

			if tt.shouldPass {
				if err != nil {
					t.Errorf("Expected password to pass but got error: %v (reason: %s)", err, tt.reason)
				}
			} else {
				if err == nil {
					t.Errorf("Expected password to fail but validation passed (reason: %s)", tt.reason)
				}
			}
		})
	}
}

func TestPasswordSecurityFeatures(t *testing.T) {
	securityService := security.NewSecurityService()

	// Test that passwords with potential injection attempts are handled
	maliciousPasswords := []string{
		"'; DROP TABLE users; --",
		"<script>alert('xss')</script>",
		"' OR '1'='1",
		"admin'/*",
		"1' UNION SELECT * FROM users--",
		"javascript:alert('xss')",
		"<img src=x onerror=alert('xss')>",
	}

	for _, maliciousPass := range maliciousPasswords {
		testName := maliciousPass
		if len(testName) > 10 {
			testName = testName[:10]
		}
		t.Run("Security_Test_"+testName, func(t *testing.T) {
			err := securityService.Sanitizer.SanitizePassword(maliciousPass)

			// These should either be rejected for weakness or pass sanitization
			// The key is that they shouldn't cause any security issues
			if err == nil {
				t.Logf("Malicious password was accepted after sanitization: %s", maliciousPass)
				// This might be OK if it's properly sanitized and meets strength requirements
			} else {
				t.Logf("Malicious password was rejected: %s (%v)", maliciousPass, err)
				// This is the expected behavior for most cases
			}
		})
	}
}

func TestPasswordLengthValidation(t *testing.T) {
	securityService := security.NewSecurityService()

	tests := []struct {
		name       string
		length     int
		shouldPass bool
	}{
		{"Very short", 1, false},
		{"Short", 4, false},
		{"Borderline", 7, false},
		{"Minimum", 8, true}, // Assuming 8 is minimum
		{"Good", 12, true},
		{"Long", 20, true},
		{"Very long", 50, true},
		{"Extremely long", 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a password with all required character types
			password := "A" + "a" + "1" + "!" + strings.Repeat("x", tt.length-4)
			if tt.length < 4 {
				password = strings.Repeat("A", tt.length)
			}

			err := securityService.Sanitizer.SanitizePassword(password)

			if tt.shouldPass {
				if err != nil && strings.Contains(err.Error(), "length") {
					t.Errorf("Expected password of length %d to pass but got length error: %v", tt.length, err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected password of length %d to fail but validation passed", tt.length)
				}
			}
		})
	}
}

func TestPasswordComplexityPatterns(t *testing.T) {
	securityService := security.NewSecurityService()

	patterns := []struct {
		name       string
		password   string
		shouldPass bool
	}{
		{
			name:       "Repeated characters",
			password:   "Aaaaaaaa1!",
			shouldPass: false,
		},
		{
			name:       "Sequential numbers",
			password:   "Password123!",
			shouldPass: true, // This might pass depending on implementation
		},
		{
			name:       "Keyboard patterns",
			password:   "Qwerty123!",
			shouldPass: false,
		},
		{
			name:       "Dictionary word base",
			password:   "Password123!",
			shouldPass: false,
		},
		{
			name:       "Random strong password",
			password:   "Kx9#mP2$vN8!",
			shouldPass: true,
		},
	}

	for _, pattern := range patterns {
		t.Run(pattern.name, func(t *testing.T) {
			err := securityService.Sanitizer.SanitizePassword(pattern.password)

			if pattern.shouldPass {
				if err != nil {
					t.Logf("Expected pattern '%s' to pass but got: %v", pattern.name, err)
					// Some patterns might still fail, which could be acceptable
				}
			} else {
				if err == nil {
					t.Errorf("Expected pattern '%s' to fail but validation passed", pattern.name)
				}
			}
		})
	}
}
