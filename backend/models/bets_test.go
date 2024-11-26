package models

import (
	"testing"
	"time"
)

func TestGetNumMarketUsers(t *testing.T) {
	tests := []struct {
		name string
		bets Bets
		want int
	}{
		{
			name: "0 users",
			bets: Bets{},
			want: 0,
		},
		{
			name: "1 user",
			bets: Bets{buildBet(t, 1, "u1")},
			want: 1,
		},
		{
			name: "1 user",
			bets: Bets{buildBet(t, 1, "u1"), buildBet(t, 2, "u1")},
			want: 1,
		},
		{
			name: "2 users",
			bets: Bets{buildBet(t, 1, "u1"), buildBet(t, 3, "u2"), buildBet(t, 2, "u1")},
			want: 2,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := GetNumMarketUsers(test.bets)
			if got != test.want {
				t.Errorf("%d market users, want %d", got, test.want)
			}
		})
	}
}

func buildBet(t *testing.T, id uint, username string) Bet {
	t.Helper()
	return Bet{
		//Action   string    `json:"action"`
		ID:       id,       //       uint      `json:"id" gorm:"primary_key"`
		Username: username, // string    `json:"username"`
		//User     User      `gorm:"foreignKey:Username;references:Username"`
		//MarketID uint      `json:"marketId"`
		//Market   Market    `gorm:"foreignKey:ID;references:MarketID"`
		//Amount   int64     `json:"amount"`
		//PlacedAt time.Time `json:"placedAt"`
		//Outcome  string    `json:"outcome,omitempty"`
	}
}

func TestCreateBet(t *testing.T) {
	username := "testuser"
	marketID := uint(123)
	amount := int64(100)
	outcome := "YES"

	bet := CreateBet(username, marketID, amount, outcome)

	// Check if the fields are set correctly
	if bet.Username != username {
		t.Errorf("expected Username %v, got %v", username, bet.Username)
	}
	if bet.MarketID != marketID {
		t.Errorf("expected MarketID %v, got %v", marketID, bet.MarketID)
	}
	if bet.Amount != amount {
		t.Errorf("expected Amount %v, got %v", amount, bet.Amount)
	}
	if bet.Outcome != outcome {
		t.Errorf("expected Outcome %v, got %v", outcome, bet.Outcome)
	}

	// Compare timestamps
	now := time.Now()
	if bet.PlacedAt.After(now) {
		t.Errorf("PlacedAt time should be in the past or now, got %v", bet.PlacedAt)
	}
}
