package models

import (
	"testing"
	"time"
)

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
