package models

import (
	"testing"
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
