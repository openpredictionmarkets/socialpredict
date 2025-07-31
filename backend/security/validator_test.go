package security

import (
	"testing"
)

func TestValidator_ValidateStruct(t *testing.T) {
	v := NewValidator()

	type TestStruct struct {
		Username    string `validate:"required,min=3,max=30,username"`
		Password    string `validate:"required,strong_password"`
		DisplayName string `validate:"required,min=1,max=50,safe_string"`
		Description string `validate:"max=2000,safe_string"`
		Amount      int64  `validate:"positive_amount"`
		Outcome     string `validate:"market_outcome"`
	}

	tests := []struct {
		name    string
		data    TestStruct
		wantErr bool
	}{
		{
			name: "valid struct",
			data: TestStruct{
				Username:    "testuser123",
				Password:    "StrongPass123",
				DisplayName: "Test User",
				Description: "Valid description",
				Amount:      100,
				Outcome:     "YES",
			},
			wantErr: false,
		},
		{
			name: "invalid username",
			data: TestStruct{
				Username:    "Test_User",
				Password:    "StrongPass123",
				DisplayName: "Test User",
				Description: "Valid description",
				Amount:      100,
				Outcome:     "YES",
			},
			wantErr: true,
		},
		{
			name: "weak password",
			data: TestStruct{
				Username:    "testuser123",
				Password:    "weak",
				DisplayName: "Test User",
				Description: "Valid description",
				Amount:      100,
				Outcome:     "YES",
			},
			wantErr: true,
		},
		{
			name: "invalid outcome",
			data: TestStruct{
				Username:    "testuser123",
				Password:    "StrongPass123",
				DisplayName: "Test User",
				Description: "Valid description",
				Amount:      100,
				Outcome:     "MAYBE",
			},
			wantErr: true,
		},
		{
			name: "negative amount",
			data: TestStruct{
				Username:    "testuser123",
				Password:    "StrongPass123",
				DisplayName: "Test User",
				Description: "Valid description",
				Amount:      -100,
				Outcome:     "YES",
			},
			wantErr: true,
		},
		{
			name: "dangerous display name",
			data: TestStruct{
				Username:    "testuser123",
				Password:    "StrongPass123",
				DisplayName: "Test <script>alert('xss')</script> User",
				Description: "Valid description",
				Amount:      100,
				Outcome:     "YES",
			},
			wantErr: true,
		},
		{
			name: "dangerous description",
			data: TestStruct{
				Username:    "testuser123",
				Password:    "StrongPass123",
				DisplayName: "Test User",
				Description: "javascript:alert('xss')",
				Amount:      100,
				Outcome:     "YES",
			},
			wantErr: true,
		},
		{
			name: "empty required fields",
			data: TestStruct{
				Username:    "",
				Password:    "",
				DisplayName: "",
				Description: "",
				Amount:      100,
				Outcome:     "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateStruct(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test validation rules constants
func TestValidationRules(t *testing.T) {
	rules := ValidationRules

	if rules.Username == "" {
		t.Error("Username validation rule should not be empty")
	}

	if rules.Password == "" {
		t.Error("Password validation rule should not be empty")
	}

	if rules.DisplayName == "" {
		t.Error("DisplayName validation rule should not be empty")
	}

	if rules.MarketTitle == "" {
		t.Error("MarketTitle validation rule should not be empty")
	}

	if rules.MarketOutcome == "" {
		t.Error("MarketOutcome validation rule should not be empty")
	}

	if rules.BetAmount == "" {
		t.Error("BetAmount validation rule should not be empty")
	}

	if rules.MarketID == "" {
		t.Error("MarketID validation rule should not be empty")
	}
}
