package positions

import (
	"socialpredict/handlers/math/outcomes/dbpm"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"sort"
	"strconv"
	"testing"
	"time"
)

func TestCalculateMarketPositions_WPAM_DBPM(t *testing.T) {
	// Define test cases
	testcases := []struct {
		Name              string
		Bets              []models.Bet
		ExpectedPositions []dbpm.MarketPosition
	}{
		{
			Name:              "InitialMarketState",
			Bets:              []models.Bet{},
			ExpectedPositions: []dbpm.MarketPosition{},
		},
		{
			Name: "FirstBetNoDirection",
			Bets: []models.Bet{
				{
					Username: "one",
					MarketID: 1,
					Amount:   20,
					Outcome:  "NO",
					PlacedAt: time.Now(),
				},
			},
			ExpectedPositions: []dbpm.MarketPosition{
				{Username: "one", NoSharesOwned: 20, YesSharesOwned: 0},
			},
		},
		{
			Name: "SecondBetYesDirection",
			Bets: []models.Bet{
				{
					Username: "one",
					MarketID: 1,
					Amount:   20,
					Outcome:  "NO",
					PlacedAt: time.Now(),
				},
				{
					Username: "two",
					MarketID: 1,
					Amount:   10,
					Outcome:  "YES",
					PlacedAt: time.Now().Add(time.Minute),
				},
			},
			ExpectedPositions: []dbpm.MarketPosition{
				{Username: "one", NoSharesOwned: 25, YesSharesOwned: 0},
				{Username: "two", NoSharesOwned: 0, YesSharesOwned: 5},
			},
		},
		{
			Name: "ThirdBetYesDirection",
			Bets: []models.Bet{
				{
					Username: "one",
					MarketID: 1,
					Amount:   20,
					Outcome:  "NO",
					PlacedAt: time.Now(),
				},
				{
					Username: "two",
					MarketID: 1,
					Amount:   10,
					Outcome:  "YES",
					PlacedAt: time.Now().Add(time.Minute),
				},
				{
					Username: "three",
					MarketID: 1,
					Amount:   10,
					Outcome:  "YES",
					PlacedAt: time.Now().Add(2 * time.Minute),
				},
			},
			ExpectedPositions: []dbpm.MarketPosition{
				{Username: "one", NoSharesOwned: 20, YesSharesOwned: 0},
				{Username: "two", NoSharesOwned: 0, YesSharesOwned: 20},
				{Username: "three", NoSharesOwned: 0, YesSharesOwned: 0},
			},
		},
		{
			Name: "FourthBetNegativeNoDirection",
			Bets: []models.Bet{
				{
					Username: "one",
					MarketID: 1,
					Amount:   20,
					Outcome:  "NO",
					PlacedAt: time.Now(),
				},
				{
					Username: "two",
					MarketID: 1,
					Amount:   10,
					Outcome:  "YES",
					PlacedAt: time.Now().Add(time.Minute),
				},
				{
					Username: "three",
					MarketID: 1,
					Amount:   10,
					Outcome:  "YES",
					PlacedAt: time.Now().Add(2 * time.Minute),
				},
				{
					Username: "one",
					MarketID: 1,
					Amount:   -10,
					Outcome:  "NO",
					PlacedAt: time.Now().Add(3 * time.Minute),
				},
			},
			ExpectedPositions: []dbpm.MarketPosition{
				{Username: "one", NoSharesOwned: 11, YesSharesOwned: 0},
				{Username: "two", NoSharesOwned: 0, YesSharesOwned: 13},
				{Username: "three", NoSharesOwned: 0, YesSharesOwned: 6},
			},
		},
	}

	// Iterate over test cases
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			// Setup the fake database
			db := modelstesting.NewFakeDB(t)

			// Create a fake market
			market := models.Market{
				ID:         1,
				IsResolved: false,
			}
			db.Create(&market)

			// Insert bets into the database
			for _, bet := range tc.Bets {
				db.Create(&bet)
			}

			// Call the function under test
			marketIDStr := strconv.Itoa(int(market.ID))
			actualPositions, err := CalculateMarketPositions_WPAM_DBPM(db, marketIDStr)

			// Verify no error
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify lengths match
			if len(actualPositions) != len(tc.ExpectedPositions) {
				t.Fatalf("expected %d positions, got %d", len(tc.ExpectedPositions), len(actualPositions))
			}

			// Verify the positions match exactly
			sort.Slice(actualPositions, func(i, j int) bool {
				return actualPositions[i].Username < actualPositions[j].Username
			})
			sort.Slice(tc.ExpectedPositions, func(i, j int) bool {
				return tc.ExpectedPositions[i].Username < tc.ExpectedPositions[j].Username
			})

			for i, pos := range actualPositions {
				expected := tc.ExpectedPositions[i]
				if pos.Username != expected.Username ||
					pos.YesSharesOwned != expected.YesSharesOwned ||
					pos.NoSharesOwned != expected.NoSharesOwned {
					t.Errorf("expected %+v, got %+v", expected, pos)
				}
			}
		})
	}
}

func TestCalculateMarketPositionForUser_WPAM_DBPM(t *testing.T) {
	// Define test cases
	testcases := []struct {
		Name             string
		Bets             []models.Bet
		Username         string
		ExpectedPosition UserMarketPosition
	}{
		{
			Name:             "InitialMarketState",
			Bets:             []models.Bet{},
			Username:         "user1",
			ExpectedPosition: UserMarketPosition{YesSharesOwned: 0, NoSharesOwned: 0},
		},
		{
			Name: "FirstBetNoDirection",
			Bets: []models.Bet{
				{
					Username: "user1",
					MarketID: 1,
					Amount:   20,
					Outcome:  "NO",
					PlacedAt: time.Now(),
				},
			},
			Username:         "user1",
			ExpectedPosition: UserMarketPosition{YesSharesOwned: 0, NoSharesOwned: 20},
		},
		{
			Name: "SecondBetYesDirection",
			Bets: []models.Bet{
				{
					Username: "user1",
					MarketID: 1,
					Amount:   20,
					Outcome:  "NO",
					PlacedAt: time.Now(),
				},
				{
					Username: "user2",
					MarketID: 1,
					Amount:   10,
					Outcome:  "YES",
					PlacedAt: time.Now().Add(time.Minute),
				},
			},
			Username:         "user1",
			ExpectedPosition: UserMarketPosition{YesSharesOwned: 0, NoSharesOwned: 25},
		},
		{
			Name: "ThirdBetYesDirection",
			Bets: []models.Bet{
				{
					Username: "user1",
					MarketID: 1,
					Amount:   20,
					Outcome:  "NO",
					PlacedAt: time.Now(),
				},
				{
					Username: "user2",
					MarketID: 1,
					Amount:   10,
					Outcome:  "YES",
					PlacedAt: time.Now().Add(time.Minute),
				},
				{
					Username: "user3",
					MarketID: 1,
					Amount:   10,
					Outcome:  "YES",
					PlacedAt: time.Now().Add(2 * time.Minute),
				},
			},
			Username:         "user1",
			ExpectedPosition: UserMarketPosition{YesSharesOwned: 0, NoSharesOwned: 20},
		},
		{
			Name: "FourthBetNegativeNoDirection",
			Bets: []models.Bet{
				{
					Username: "user1",
					MarketID: 1,
					Amount:   20,
					Outcome:  "NO",
					PlacedAt: time.Now(),
				},
				{
					Username: "user2",
					MarketID: 1,
					Amount:   10,
					Outcome:  "YES",
					PlacedAt: time.Now().Add(time.Minute),
				},
				{
					Username: "user3",
					MarketID: 1,
					Amount:   10,
					Outcome:  "YES",
					PlacedAt: time.Now().Add(2 * time.Minute),
				},
				{
					Username: "user1",
					MarketID: 1,
					Amount:   -10,
					Outcome:  "NO",
					PlacedAt: time.Now().Add(3 * time.Minute),
				},
			},
			Username:         "user1",
			ExpectedPosition: UserMarketPosition{YesSharesOwned: 0, NoSharesOwned: 11},
		},
	}

	// Iterate over test cases
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			// Setup the fake database
			db := modelstesting.NewFakeDB(t)

			// Create a fake market
			market := models.Market{
				ID:         1,
				IsResolved: false,
			}
			db.Create(&market)

			// Insert bets into the database
			for _, bet := range tc.Bets {
				db.Create(&bet)
			}

			// Call the function under test
			marketIDStr := strconv.Itoa(int(market.ID))
			actualPosition, err := CalculateMarketPositionForUser_WPAM_DBPM(db, marketIDStr, tc.Username)

			// Verify no error
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify the calculated position matches the expected position
			if actualPosition.YesSharesOwned != tc.ExpectedPosition.YesSharesOwned ||
				actualPosition.NoSharesOwned != tc.ExpectedPosition.NoSharesOwned {
				t.Errorf("expected position %+v, got %+v", tc.ExpectedPosition, actualPosition)
			}
		})
	}
}
