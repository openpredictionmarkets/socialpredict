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

	totalVolume := int64(3798) // actual total pool

	// Users with precomputed winning share amounts and spend
	// Users with precomputed winning share amounts and their actual spend
	users := []struct {
		username string
		shares   int64
		spend    int64
	}{
		{"followbot", 347, 977},
		{"opposebot", 295, 933},
		{"fabulousbot", 83, 412},
		{"vancebot", 36, 91},
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
		"followbot":   1244,
		"opposebot":   1005,
		"fabulousbot": 1,
		"vancebot":    88,
	}

	for _, u := range users {
		user := modelstesting.GenerateUser(u.username, u.spend)
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
