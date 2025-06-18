package payout

import (
	"socialpredict/handlers/math/outcomes/dbpm"
	"socialpredict/handlers/positions"
	"socialpredict/models"
	modelstesting "socialpredict/models/modelstesting"
	"testing"
)

func TestSelectWinningPositions_YesResolution(t *testing.T) {
	rawPositions := []dbpm.DBPMMarketPosition{
		{Username: "followbot", YesSharesOwned: 347, NoSharesOwned: 0},
		{Username: "opposebot", YesSharesOwned: 295, NoSharesOwned: 0},
		{Username: "fabulousbot", YesSharesOwned: 83, NoSharesOwned: 0},
		{Username: "vancebot", YesSharesOwned: 36, NoSharesOwned: 0},
	}

	expectedTotalShares := int64(761)
	expectedUsers := map[string]int64{
		"followbot":   347,
		"opposebot":   295,
		"fabulousbot": 83,
		"vancebot":    36,
	}

	winningPositions, totalShares := SelectWinningPositions("YES", rawPositions)

	if totalShares != expectedTotalShares {
		t.Errorf("totalShares = %d, want %d", totalShares, expectedTotalShares)
	}

	if len(winningPositions) != len(expectedUsers) {
		t.Fatalf("got %d winning users, want %d", len(winningPositions), len(expectedUsers))
	}

	for _, pos := range winningPositions {
		expectedShares, ok := expectedUsers[pos.Username]
		if !ok {
			t.Errorf("unexpected winner: %s", pos.Username)
		}
		if pos.YesSharesOwned != expectedShares {
			t.Errorf("user %s shares = %d, want %d", pos.Username, pos.YesSharesOwned, expectedShares)
		}
		if pos.NoSharesOwned != 0 {
			t.Errorf("user %s should have 0 NO shares, got %d", pos.Username, pos.NoSharesOwned)
		}
	}
}

func TestAllocateWinningSharePool_RoundedDistribution_With_Fee(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	market := modelstesting.GenerateMarket(1, "creator")
	market.ResolutionResult = "YES"
	db.Create(&market)

	totalVolume := int64(3798)

	var initialBettingFee int64 = 1

	// Users with precomputed winning share amounts and spend
	// Users with precomputed winning share amounts and their actual spend
	users := []struct {
		username string
		shares   int64
		spend    int64
	}{
		{"followbot", 347, 488 + initialBettingFee},
		{"opposebot", 295, 466 + initialBettingFee},
		{"fabulousbot", 83, 412 + initialBettingFee},
		{"vancebot", 36, 91 + initialBettingFee},
	}

	var totalWinningShares int64
	var thepositions []positions.MarketPosition

	for _, u := range users {
		user := modelstesting.GenerateUser(u.username, 0)
		db.Create(&user)

		// Simulate betting spend (debt)
		user.AccountBalance = -u.spend
		db.Save(&user)

		totalWinningShares += u.shares
		thepositions = append(thepositions, positions.MarketPosition{
			Username:       u.username,
			YesSharesOwned: u.shares,
			NoSharesOwned:  0,
		})
	}

	expected := map[string]int64{
		"followbot":   1243,
		"opposebot":   1005,
		"fabulousbot": 1,
		"vancebot":    88,
	}

	err := AllocateWinningSharePool(db, &market, thepositions, totalWinningShares, totalVolume)
	if err != nil {
		t.Fatalf("AllocateWinningSharePool failed: %v", err)
	}

	for user, expectedBalance := range expected {
		var u models.User
		if err := db.First(&u, "username = ?", user).Error; err != nil {
			t.Errorf("user %s not found", user)
			continue
		}

		if u.AccountBalance != expectedBalance {
			t.Errorf("user %s balance = %d, want %d", user, u.AccountBalance, expectedBalance)
		}
	}
}
