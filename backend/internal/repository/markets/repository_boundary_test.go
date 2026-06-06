package markets

import (
	"testing"
	"time"
)

func TestMapModelBetsToBoundaryPreservesBetFields(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	dbBets := []betDustRow{{
		DustRecorded: true,
	}}
	dbBets[0].ID = 9
	dbBets[0].Username = "alice"
	dbBets[0].MarketID = 7
	dbBets[0].Amount = 25
	dbBets[0].Dust = 2
	dbBets[0].Outcome = "YES"
	dbBets[0].PlacedAt = now

	got := mapModelBetsToBoundary(dbBets)
	if len(got) != 1 {
		t.Fatalf("expected 1 bet, got %d", len(got))
	}
	if got[0].ID != 9 || got[0].Username != "alice" || got[0].MarketID != 7 || got[0].Amount != 25 || got[0].Dust != 2 || !got[0].DustRecorded || got[0].Outcome != "YES" || !got[0].PlacedAt.Equal(now) {
		t.Fatalf("unexpected mapped bet: %#v", got[0])
	}
}
