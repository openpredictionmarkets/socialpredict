package markets

import (
	"testing"
	"time"

	"socialpredict/models"
)

func TestMapModelBetsToBoundaryPreservesBetFields(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	dbBets := []models.Bet{{
		ID:       9,
		Username: "alice",
		MarketID: 7,
		Amount:   25,
		Outcome:  "YES",
		PlacedAt: now,
	}}

	got := mapModelBetsToBoundary(dbBets)
	if len(got) != 1 {
		t.Fatalf("expected 1 bet, got %d", len(got))
	}
	if got[0].ID != 9 || got[0].Username != "alice" || got[0].MarketID != 7 || got[0].Amount != 25 || got[0].Outcome != "YES" || !got[0].PlacedAt.Equal(now) {
		t.Fatalf("unexpected mapped bet: %#v", got[0])
	}
}
