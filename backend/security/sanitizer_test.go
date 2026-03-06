package security

import (
	"testing"
)

func TestSanitizer_SanitizeUsername(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "valid username",
			input:    "testuser123",
			expected: "testuser123",
			wantErr:  false,
		},
		{
			name:     "username with uppercase",
			input:    "TestUser",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "username with special characters",
			input:    "test_user",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "username too short",
			input:    "ab",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "username too long",
			input:    "abcdefghijklmnopqrstuvwxyz12345",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "username with whitespace",
			input:    "  testuser  ",
			expected: "testuser",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := s.SanitizeUsername(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SanitizeUsername() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("SanitizeUsername() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSanitizer_SanitizeDisplayName(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "valid display name",
			input:    "John Doe",
			expected: "John Doe",
			wantErr:  false,
		},
		{
			name:     "display name with HTML",
			input:    "John <script>alert('xss')</script> Doe",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "empty display name",
			input:    "",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "display name too long",
			input:    "This is a very long display name that exceeds fifty characters limit",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "display name with javascript",
			input:    "javascript:alert('xss')",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := s.SanitizeDisplayName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SanitizeDisplayName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("SanitizeDisplayName() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSanitizer_SanitizeDescription(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "valid description",
			input:    "This is a normal description.",
			expected: "This is a normal description.",
			wantErr:  false,
		},
		{
			name:     "description with basic HTML",
			input:    "This is <b>bold</b> text.",
			expected: "This is <b>bold</b> text.",
			wantErr:  false,
		},
		{
			name:     "description with script",
			input:    "This is <script>alert('xss')</script> dangerous.",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "empty description",
			input:    "",
			expected: "",
			wantErr:  false,
		},
		{
			name:     "description too long",
			input:    string(make([]byte, 2001)),
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := s.SanitizeDescription(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SanitizeDescription() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("SanitizeDescription() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSanitizer_SanitizeMarketTitle(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "valid market title",
			input:    "Will it rain tomorrow?",
			expected: "Will it rain tomorrow?",
			wantErr:  false,
		},
		{
			name:     "market title with script",
			input:    "Will <script>alert('xss')</script> it rain?",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "empty market title",
			input:    "",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "market title too long",
			input:    string(make([]byte, 161)),
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := s.SanitizeMarketTitle(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SanitizeMarketTitle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("SanitizeMarketTitle() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSanitizer_SanitizePersonalLink(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "valid HTTPS URL",
			input:    "https://example.com",
			expected: "https://example.com",
			wantErr:  false,
		},
		{
			name:     "valid HTTP URL",
			input:    "http://example.com",
			expected: "http://example.com",
			wantErr:  false,
		},
		{
			name:     "URL without scheme",
			input:    "example.com",
			expected: "https://example.com",
			wantErr:  false,
		},
		{
			name:     "empty URL",
			input:    "",
			expected: "",
			wantErr:  false,
		},
		{
			name:     "malicious URL with localhost",
			input:    "http://localhost:8080",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "invalid scheme",
			input:    "ftp://example.com",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "URL too long",
			input:    "https://" + string(make([]byte, 200)) + ".com",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := s.SanitizePersonalLink(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SanitizePersonalLink() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("SanitizePersonalLink() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSanitizer_SanitizeEmoji(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "valid emoji",
			input:    "😀",
			expected: "😀",
			wantErr:  false,
		},
		{
			name:     "simple ASCII emoji",
			input:    ":)",
			expected: ":)",
			wantErr:  false,
		},
		{
			name:     "empty emoji",
			input:    "",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "emoji too long",
			input:    string(make([]byte, 21)),
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := s.SanitizeEmoji(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SanitizeEmoji() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("SanitizeEmoji() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSanitizer_SanitizePassword(t *testing.T) {
	s := NewSanitizer()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid strong password",
			input:   "StrongPass123",
			wantErr: false,
		},
		{
			name:    "password too short",
			input:   "Short1",
			wantErr: true,
		},
		{
			name:    "password without uppercase",
			input:   "lowercase123",
			wantErr: true,
		},
		{
			name:    "password without lowercase",
			input:   "UPPERCASE123",
			wantErr: true,
		},
		{
			name:    "password without digit",
			input:   "NoDigitPass",
			wantErr: true,
		},
		{
			name:    "password too long",
			input:   string(make([]byte, 129)),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := s.SanitizePassword(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SanitizePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestContainsSuspiciousPatterns(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "normal text",
			input:    "This is normal text",
			expected: false,
		},
		{
			name:     "javascript protocol",
			input:    "javascript:alert('xss')",
			expected: true,
		},
		{
			name:     "script tag",
			input:    "<script>alert('xss')</script>",
			expected: true,
		},
		{
			name:     "onclick event",
			input:    "onclick=alert('xss')",
			expected: true,
		},
		{
			name:     "data protocol",
			input:    "data:text/html,<script>alert('xss')</script>",
			expected: true,
		},
		{
			name:     "HTML comment",
			input:    "<!-- malicious comment -->",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsSuspiciousPatterns(tt.input)
			if result != tt.expected {
				t.Errorf("containsSuspiciousPatterns() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestContainsMaliciousDomain(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "safe domain",
			input:    "example.com",
			expected: false,
		},
		{
			name:     "localhost",
			input:    "localhost",
			expected: true,
		},
		{
			name:     "private IP",
			input:    "192.168.1.1",
			expected: true,
		},
		{
			name:     "URL shortener",
			input:    "bit.ly",
			expected: true,
		},
		{
			name:     "local IP",
			input:    "127.0.0.1",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsMaliciousDomain(tt.input)
			if result != tt.expected {
				t.Errorf("containsMaliciousDomain() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsValidEmoji(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid unicode emoji",
			input:    "😀",
			expected: true,
		},
		{
			name:     "multiple unicode emojis",
			input:    "😀👍",
			expected: true,
		},
		{
			name:     "ASCII emoji",
			input:    ":)",
			expected: true,
		},
		{
			name:     "regional indicator pair",
			input:    "🇺🇸",
			expected: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "contains control character",
			input:    "😀\n",
			expected: false,
		},
		{
			name:     "non-emoji non-ascii characters",
			input:    "中文",
			expected: false,
		},
		{
			name:     "regular text",
			input:    "abc",
			expected: true, // ASCII characters are allowed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidEmoji(tt.input)
			if result != tt.expected {
				t.Errorf("isValidEmoji() = %v, want %v", result, tt.expected)
			}
		})
	}
}
