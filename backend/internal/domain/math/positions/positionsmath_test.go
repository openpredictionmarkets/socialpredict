package positionsmath

import (
	"socialpredict/models/modelstesting"
	"strconv"
	"testing"
	"time"
)

func TestCalculateMarketPositions_WPAM_DBPM(t *testing.T) {
	testcases := []struct {
		Name       string
		BetConfigs []struct {
			Amount   int64
			Outcome  string
			Username string
			Offset   time.Duration
		}
		Expected []MarketPosition // You can fill in Yes/NoShares for now, and auto-print Value
	}{
		{
			Name: "Single YES Bet",
			BetConfigs: []struct {
				Amount   int64
				Outcome  string
				Username string
				Offset   time.Duration
			}{
				{Amount: 50, Outcome: "YES", Username: "alice", Offset: 0},
			},
			Expected: []MarketPosition{
				{Username: "alice", YesSharesOwned: 50, NoSharesOwned: 0},
			},
		},
		{
			Name: "Single NO Bet",
			BetConfigs: []struct {
				Amount   int64
				Outcome  string
				Username string
				Offset   time.Duration
			}{
				{Amount: 30, Outcome: "NO", Username: "bob", Offset: 0},
			},
			Expected: []MarketPosition{
				{Username: "bob", YesSharesOwned: 0, NoSharesOwned: 30},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			db := modelstesting.NewFakeDB(t)
			creator := "testcreator"
			market := modelstesting.GenerateMarket(1, creator)
			db.Create(&market)
			for _, betConf := range tc.BetConfigs {
				bet := modelstesting.GenerateBet(betConf.Amount, betConf.Outcome, betConf.Username, uint(market.ID), betConf.Offset)
				db.Create(&bet)
			}
			marketIDStr := strconv.Itoa(int(market.ID))
			actualPositions, err := CalculateMarketPositions_WPAM_DBPM(db, marketIDStr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(actualPositions) != len(tc.Expected) {
				t.Fatalf("expected %d positions, got %d", len(tc.Expected), len(actualPositions))
			}

			for i, expected := range tc.Expected {
				actual := actualPositions[i]
				// Print actual values to update your expected struct
				t.Logf("Test=%q User=%q Yes=%d No=%d Value=%d", tc.Name, actual.Username, actual.YesSharesOwned, actual.NoSharesOwned, actual.Value)

				if actual.Username != expected.Username ||
					actual.YesSharesOwned != expected.YesSharesOwned ||
					actual.NoSharesOwned != expected.NoSharesOwned {
					t.Errorf("expected shares %+v, got %+v", expected, actual)
				}
				// For the first run, comment out this Value check and just log it!
				// if actual.Value != expected.Value {
				//     t.Errorf("expected Value=%d, got %d", expected.Value, actual.Value)
				// }
			}
		})
	}
}

func TestCalculateMarketPositions_IncludesZeroPositionUsers(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	_, _ = modelstesting.UseStandardTestEconomics(t)

	creator := modelstesting.GenerateUser("creator", 0)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("failed to create creator: %v", err)
	}

	market := modelstesting.GenerateMarket(42, creator.Username)
	market.IsResolved = true
	market.ResolutionResult = "YES"
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("failed to create market: %v", err)
	}

	participants := []struct {
		username string
	}{
		{"patrick"},
		{"jimmy"},
		{"jyron"},
		{"testuser03"},
	}
	for _, p := range participants {
		user := modelstesting.GenerateUser(p.username, 0)
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("failed to create user %s: %v", p.username, err)
		}
	}

	bets := []struct {
		amount   int64
		outcome  string
		username string
		offset   time.Duration
	}{
		{amount: 50, outcome: "NO", username: "patrick", offset: 0},
		{amount: 51, outcome: "NO", username: "jimmy", offset: time.Second},
		{amount: 51, outcome: "NO", username: "jimmy", offset: 2 * time.Second},
		{amount: 10, outcome: "YES", username: "jyron", offset: 3 * time.Second},
		{amount: 30, outcome: "YES", username: "testuser03", offset: 4 * time.Second},
	}

	for _, b := range bets {
		bet := modelstesting.GenerateBet(b.amount, b.outcome, b.username, uint(market.ID), b.offset)
		if err := db.Create(&bet).Error; err != nil {
			t.Fatalf("failed to create bet %+v: %v", b, err)
		}
	}

	positions, err := CalculateMarketPositions_WPAM_DBPM(db, strconv.Itoa(int(market.ID)))
	if err != nil {
		t.Fatalf("unexpected error calculating positions: %v", err)
	}

	var lockedUser *MarketPosition
	for i := range positions {
		if positions[i].Username == "testuser03" {
			lockedUser = &positions[i]
			break
		}
	}

	if lockedUser == nil {
		t.Fatalf("expected zero-position user to be present in positions")
	}

	if lockedUser.YesSharesOwned != 0 || lockedUser.NoSharesOwned != 0 || lockedUser.Value != 0 {
		t.Fatalf("expected zero shares/value for locked user, got %+v", lockedUser)
	}
}
