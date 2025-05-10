package marketshandlers

import (
	"socialpredict/handlers/positions"
	"socialpredict/models"
	modelstesting "socialpredict/models/modelstesting"
	"testing"
)

func TestAllocateWinningSharePool_RoundedDistribution(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	market := modelstesting.GenerateMarket(1, "creator")
	market.ResolutionResult = "YES"

	db.Create(&market)

	totalVolume := int64(1000) // or calculate from your raw data

	// Users with precomputed winning share amounts
	users := []struct {
		username string
		shares   int64
	}{
		{"followbot", 347},
		{"opposebot", 295},
		{"fabulousbot", 83},
		{"vancebot", 36},
	}

	var totalWinningShares int64
	var thepositions []positions.MarketPosition

	for _, u := range users {
		user := modelstesting.GenerateUser(u.username, 0)
		db.Create(&user)
		totalWinningShares += u.shares
		thepositions = append(thepositions, positions.MarketPosition{
			Username:       u.username,
			YesSharesOwned: u.shares,
			NoSharesOwned:  0,
		})
	}

	err := AllocateWinningSharePool(db, &market, thepositions, totalWinningShares, totalVolume)
	if err != nil {
		t.Fatalf("AllocateWinningSharePool failed: %v", err)
	}

	// Check user balances
	expected := map[string]int64{
		"followbot":   457, // one extra credit rounded to followbot from liquidity pool
		"opposebot":   387,
		"fabulousbot": 109,
		"vancebot":    47,
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
