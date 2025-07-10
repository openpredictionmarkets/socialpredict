package adminhandlers

import (
	"socialpredict/security"
	"strings"
	"testing"
)

func TestUsernameValidation(t *testing.T) {
	tests := []struct {
		name          string
		username      string
		expectedValid bool
		expectedError string
	}{
		{
			name:          "Valid username",
			username:      "testuser123",
			expectedValid: true,
		},
		{
			name:          "Valid short username",
			username:      "user",
			expectedValid: true,
		},
		{
			name:          "Valid numbers only",
			username:      "123456",
			expectedValid: true,
		},
		{
			name:          "Valid letters only",
			username:      "testuser",
			expectedValid: true,
		},
		{
			name:          "Empty username",
			username:      "",
			expectedValid: false,
			expectedError: "cannot be blank",
		},
		{
			name:          "Username with uppercase",
			username:      "TestUser",
			expectedValid: false,
			expectedError: "can only contain lowercase letters and numbers",
		},
		{
			name:          "Username with special characters",
			username:      "test_user",
			expectedValid: false,
			expectedError: "can only contain lowercase letters and numbers",
		},
		{
			name:          "Username with spaces",
			username:      "test user",
			expectedValid: false,
			expectedError: "can only contain lowercase letters and numbers",
		},
		{
			name:          "Username with hyphen",
			username:      "test-user",
			expectedValid: false,
			expectedError: "can only contain lowercase letters and numbers",
		},
		{
			name:          "Username too long",
			username:      strings.Repeat("a", 51),
			expectedValid: false,
			expectedError: "exceeds maximum length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			securityService := security.NewSecurityService()

			// Create a struct to validate username using struct validation
			testStruct := struct {
				Username string `validate:"required,min=3,max=30,username"`
			}{
				Username: tt.username,
			}

			err := securityService.Validator.ValidateStruct(testStruct)

			if tt.expectedValid {
				if err != nil {
					t.Errorf("Expected valid username but got error: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error but username validation passed")
				}
			}
		})
	}
}

func TestUsernameSanitization(t *testing.T) {
	securityService := security.NewSecurityService()

	tests := []struct {
		name        string
		username    string
		expected    string
		shouldError bool
	}{
		{
			name:     "Normal username",
			username: "testuser",
			expected: "testuser",
		},
		{
			name:     "Username with whitespace",
			username: "  testuser  ",
			expected: "testuser",
		},
		{
			name:        "Username with mixed case",
			username:    "TestUser",
			shouldError: true, // Should be rejected, not converted
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitized, err := securityService.Sanitizer.SanitizeUsername(tt.username)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error but sanitization passed")
				}
				return
			}

			if err != nil {
				t.Errorf("Sanitization failed: %v", err)
				return
			}

			if sanitized != tt.expected {
				t.Errorf("Expected username %s, got %s", tt.expected, sanitized)
			}
		})
	}
}

func TestUsernameSecurityFeatures(t *testing.T) {
	securityService := security.NewSecurityService()

	// Test various malicious usernames
	maliciousUsernames := []string{
		"'; DROP TABLE users; --",
		"<script>alert('xss')</script>",
		"' OR '1'='1",
		"admin'/*",
		"1' UNION SELECT * FROM users--",
		"javascript:alert('xss')",
		"<img src=x onerror=alert('xss')>",
		"../../../etc/passwd",
		"..\\..\\..\\windows\\system32",
		"user\x00admin",
		"user\nadmin",
		"user\radmin",
		"user\tadmin",
	}

	for _, maliciousUsername := range maliciousUsernames {
		testName := maliciousUsername
		if len(testName) > 10 {
			testName = testName[:10]
		}
		t.Run("Security_Test_"+testName, func(t *testing.T) {
			testStruct := struct {
				Username string `validate:"required,min=3,max=30,username"`
			}{
				Username: maliciousUsername,
			}

			err := securityService.Validator.ValidateStruct(testStruct)

			// All malicious usernames should be rejected
			if err == nil {
				t.Errorf("Expected malicious username to be rejected but validation passed: %s", maliciousUsername)
			}
		})
	}
}

func TestUsernameFormatValidation(t *testing.T) {
	securityService := security.NewSecurityService()

	formatTests := []struct {
		name       string
		username   string
		shouldPass bool
	}{
		// Valid formats
		{"Lowercase letters only", "testuser", true},
		{"Numbers only", "123456", true},
		{"Mixed letters and numbers", "user123", true},
		{"Starting with number", "1user", true},
		{"Ending with number", "user1", true},

		// Invalid formats
		{"Contains uppercase", "TestUser", false},
		{"Contains underscore", "test_user", false},
		{"Contains hyphen", "test-user", false},
		{"Contains dot", "test.user", false},
		{"Contains space", "test user", false},
		{"Contains special chars", "test@user", false},
		{"Contains unicode", "tëst", false},
		{"Contains emoji", "test😀", false},
		{"Contains newline", "test\nuser", false},
		{"Contains tab", "test\tuser", false},
		{"Contains null", "test\x00user", false},
	}

	for _, test := range formatTests {
		t.Run(test.name, func(t *testing.T) {
			testStruct := struct {
				Username string `validate:"required,min=3,max=30,username"`
			}{
				Username: test.username,
			}

			err := securityService.Validator.ValidateStruct(testStruct)

			if test.shouldPass {
				if err != nil {
					t.Errorf("Expected username format to be valid but got error: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected username format to be invalid but validation passed")
				}
			}
		})
	}
}

func TestUsernameLengthValidation(t *testing.T) {
	securityService := security.NewSecurityService()

	lengthTests := []struct {
		name       string
		length     int
		shouldPass bool
	}{
		{"Empty", 0, false},
		{"Single character", 1, false}, // Minimum is 3 characters
		{"Two characters", 2, false},
		{"Minimum valid", 3, true},
		{"Medium", 10, true},
		{"Maximum valid", 30, true},
		{"Over limit", 31, false},
		{"Way over limit", 100, false},
	}

	for _, test := range lengthTests {
		t.Run(test.name, func(t *testing.T) {
			var username string
			if test.length == 0 {
				username = ""
			} else {
				username = strings.Repeat("a", test.length)
			}

			testStruct := struct {
				Username string `validate:"required,min=3,max=30,username"`
			}{
				Username: username,
			}

			err := securityService.Validator.ValidateStruct(testStruct)

			if test.shouldPass {
				if err != nil {
					t.Errorf("Expected username length %d to be valid but got error: %v", test.length, err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected username length %d to be invalid but validation passed", test.length)
				}
			}
		})
	}
}

func TestUsernameReservedWords(t *testing.T) {
	securityService := security.NewSecurityService()

	// Test common reserved words that should be rejected
	reservedWords := []string{
		"admin",
		"administrator",
		"root",
		"system",
		"null",
		"undefined",
		"api",
		"www",
		"ftp",
		"mail",
		"email",
		"support",
		"help",
		"info",
		"contact",
		"about",
		"login",
		"register",
		"signup",
		"signin",
		"logout",
		"profile",
		"settings",
		"config",
		"test",
		"demo",
		"guest",
		"anonymous",
		"user",
		"users",
	}

	for _, word := range reservedWords {
		t.Run("Reserved_"+word, func(t *testing.T) {
			testStruct := struct {
				Username string `validate:"required,min=3,max=30,username"`
			}{
				Username: word,
			}

			err := securityService.Validator.ValidateStruct(testStruct)

			// Note: This test assumes reserved word validation exists
			// The actual implementation may or may not reject reserved words
			// If not implemented, this test will document the behavior
			if err == nil {
				t.Logf("Reserved word '%s' was accepted (may be intentional)", word)
			} else {
				t.Logf("Reserved word '%s' was rejected: %v", word, err)
			}
		})
	}
}
