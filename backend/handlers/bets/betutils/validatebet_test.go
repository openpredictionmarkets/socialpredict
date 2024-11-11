package betutils

import (
	"testing"

	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestValidateBuy(t *testing.T) {

	db := modelstesting.NewFakeDB(t)

	user := &models.User{
		PublicUser: models.PublicUser{
			Username:       "testuser",
			AccountBalance: 0,
		},
	}

	market := models.Market{
		ID:         1,
		IsResolved: false,
	}

	db.Create(&user)
	db.Create(&market)

	tests := []struct {
		name       string
		bet        models.Bet
		expectsErr bool
		errMsg     string
	}{
		{
			name: "Valid bet amount",
			bet: models.Bet{
				Username: "testuser",
				MarketID: 1,
				Amount:   50,
				Outcome:  "YES",
			},
			expectsErr: false,
		},
		{
			name: "Invalid bet amount (less than 1)",
			bet: models.Bet{
				Username: "testuser",
				MarketID: 1,
				Amount:   0,
				Outcome:  "YES",
			},
			expectsErr: true,
			errMsg:     "Buy amount must be greater than or equal to 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBuy(db, &tt.bet)
			if (err != nil) != tt.expectsErr {
				t.Errorf("got error = %v, expected error = %v", err, tt.expectsErr)
			}
			if err != nil && tt.expectsErr && err.Error() != tt.errMsg {
				t.Errorf("expected error message %v, got %v", tt.errMsg, err.Error())
			}
		})
	}
}

func TestValidateSale(t *testing.T) {

	db := modelstesting.NewFakeDB(t)

	user := &models.User{
		PublicUser: models.PublicUser{
			Username:       "testuser",
			AccountBalance: 0,
		},
	}

	market := models.Market{
		ID:         1,
		IsResolved: false,
	}

	db.Create(&user)
	db.Create(&market)

	tests := []struct {
		name       string
		bet        models.Bet
		expectsErr bool
		errMsg     string
	}{
		{
			name: "Valid sale amount",
			bet: models.Bet{
				Username: "testuser",
				MarketID: 1,
				Amount:   -50,
				Outcome:  "YES",
			},
			expectsErr: false,
		},
		{
			name: "Invalid sale amount (greater than -1)",
			bet: models.Bet{
				Username: "testuser",
				MarketID: 1,
				Amount:   0,
				Outcome:  "YES",
			},
			expectsErr: true,
			errMsg:     "Sale amount must be greater than or equal to 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSale(db, &tt.bet)
			if (err != nil) != tt.expectsErr {
				t.Errorf("got error = %v, expected error = %v", err, tt.expectsErr)
			}
			if err != nil && tt.expectsErr && err.Error() != tt.errMsg {
				t.Errorf("expected error message %v, got %v", tt.errMsg, err.Error())
			}
		})
	}
}
