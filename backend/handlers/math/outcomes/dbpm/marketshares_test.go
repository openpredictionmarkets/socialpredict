package dbpm

import (
	"reflect"
	"socialpredict/handlers/math/probabilities/wpam"
	"socialpredict/models"
	"socialpredict/setup"
	"testing"
	"time"
)

var now = time.Now()

// helper function to generate bets succiently
func generateBet(amount int64, outcome, username string, marketID uint, offset time.Duration) models.Bet {
	return models.Bet{
		Amount:   amount,
		Outcome:  outcome,
		Username: username,
		PlacedAt: time.Now().Add(offset),
		MarketID: marketID,
	}
}

// helper function to create wpam.ProbabilityChange points succiently
func generateProbability(probabilities ...float64) []wpam.ProbabilityChange {
	var changes []wpam.ProbabilityChange
	for _, p := range probabilities {
		changes = append(changes, wpam.ProbabilityChange{Probability: p})
	}
	return changes
}

// helper function to create course payouts succiently
func generateCoursePayouts(payouts []float64, outcomes []string) []CourseBetPayout {
	if len(payouts) != len(outcomes) {
		panic("payouts and outcomes slices must have the same length")
	}

	var coursePayouts []CourseBetPayout
	for i, payout := range payouts {
		coursePayouts = append(coursePayouts, CourseBetPayout{
			Payout:  payout,
			Outcome: outcomes[i],
		})
	}
	return coursePayouts
}

func TestDivideUpMarketPoolSharesDBPM(t *testing.T) {
	testcases := []struct {
		Name               string
		Bets               []models.Bet
		ProbabilityChanges []wpam.ProbabilityChange
		S_YES              int64
		S_NO               int64
	}{
		{
			Name:               "InitialMarketState",
			Bets:               []models.Bet{},
			ProbabilityChanges: generateProbability(0.500),
			S_YES:              0,
			S_NO:               0,
		},
		{
			Name: "FirstBetNoDirection",
			Bets: []models.Bet{
				generateBet(20, "NO", "one", 1, 0),
			},
			ProbabilityChanges: generateProbability(0.500, 0.167),
			S_YES:              3,
			S_NO:               17,
		},
		{
			Name: "SecondBetYesDirection",
			Bets: []models.Bet{
				generateBet(20, "NO", "one", 1, 0),
				generateBet(10, "YES", "two", 1, time.Minute),
			},
			ProbabilityChanges: generateProbability(0.500, 0.167, 0.375),
			S_YES:              11,
			S_NO:               19,
		},
		{
			Name: "ThirdBetYesDirection",
			Bets: []models.Bet{
				generateBet(20, "NO", "one", 1, 0),
				generateBet(10, "YES", "two", 1, time.Minute),
				generateBet(10, "YES", "three", 1, 2*time.Minute),
			},
			ProbabilityChanges: generateProbability(0.500, 0.167, 0.375, 0.500),
			S_YES:              20,
			S_NO:               20,
		},
		{
			Name: "FourthBetNegativeNoDirection",
			Bets: []models.Bet{
				generateBet(20, "NO", "one", 1, 0),
				generateBet(10, "YES", "two", 1, time.Minute),
				generateBet(10, "YES", "three", 1, 2*time.Minute),
				generateBet(-10, "NO", "one", 1, 3*time.Minute),
			},
			ProbabilityChanges: generateProbability(0.500, 0.167, 0.375, 0.500, 0.625),
			S_YES:              19,
			S_NO:               11,
		},
		{
			Name: "NOResolution",
			Bets: []models.Bet{
				generateBet(20, "NO", "one", 1, 0),
				generateBet(10, "YES", "two", 1, time.Minute),
			},
			ProbabilityChanges: generateProbability(0.500, 0.167, 0.0), // Final resolution R = 0
			S_YES:              0,
			S_NO:               30, // All shares go to NO
		},
		{
			Name: "YESResolution",
			Bets: []models.Bet{
				generateBet(20, "NO", "one", 1, 0),
				generateBet(10, "YES", "two", 1, time.Minute),
			},
			ProbabilityChanges: generateProbability(0.500, 0.167, 1.0), // Final resolution R = 1
			S_YES:              30,                                     // All shares go to YES
			S_NO:               0,
		},
	}

	// Example economics config setup
	ec := setup.EconomicsConfig()
	ec.Economics.MarketCreation.InitialMarketSubsidization = 0
	ec.Economics.MarketIncentives.CreateMarketCost = 1

	// Iterate through test cases
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			yes, no := DivideUpMarketPoolSharesDBPM(tc.Bets, tc.ProbabilityChanges)
			if yes != tc.S_YES || no != tc.S_NO {
				t.Errorf("%s: expected (%d, %d), got (%d, %d)", tc.Name, tc.S_YES, tc.S_NO, yes, no)
			}
		})
	}
}

func TestCalculateCoursePayoutsDBPM(t *testing.T) {
	testcases := []struct {
		Name               string
		Bets               []models.Bet
		ProbabilityChanges []wpam.ProbabilityChange
		ExpectedPayouts    []CourseBetPayout
	}{
		{
			Name:               "InitialMarketState",
			Bets:               []models.Bet{},
			ProbabilityChanges: generateProbability(0.500),
			ExpectedPayouts:    nil, // No bets -> No payouts
		},
		{
			Name: "FirstBetNoDirection",
			Bets: []models.Bet{
				generateBet(20, "NO", "one", 1, 0),
			},
			ProbabilityChanges: generateProbability(0.500, 0.167),
			ExpectedPayouts: generateCoursePayouts(
				[]float64{0}, // Payout = |0.167 - 0.167| * 20
				[]string{"NO"},
			),
		},
		{
			Name: "SecondBetYesDirection",
			Bets: []models.Bet{
				generateBet(20, "NO", "one", 1, 0),
				generateBet(10, "YES", "two", 1, time.Minute),
			},
			ProbabilityChanges: generateProbability(0.500, 0.167, 0.375),
			ExpectedPayouts: generateCoursePayouts(
				[]float64{4.1600000000000001, 0},
				[]string{"NO", "YES"},
			),
		},
		{
			Name: "ThirdBetYesDirection",
			Bets: []models.Bet{
				generateBet(20, "NO", "one", 1, 0),
				generateBet(10, "YES", "two", 1, time.Minute),
				generateBet(10, "YES", "three", 1, 2*time.Minute),
			},
			ProbabilityChanges: generateProbability(0.500, 0.167, 0.375, 0.500),
			ExpectedPayouts: generateCoursePayouts(
				[]float64{6.6599999999999993, 1.25, 0},
				[]string{"NO", "YES", "YES"},
			),
		},
		{
			Name: "FourthBetNegativeNoDirection",
			Bets: []models.Bet{
				generateBet(20, "NO", "one", 1, 0),
				generateBet(10, "YES", "two", 1, time.Minute),
				generateBet(10, "YES", "three", 1, 2*time.Minute),
				generateBet(-10, "NO", "one", 1, 3*time.Minute),
			},
			ProbabilityChanges: generateProbability(0.500, 0.167, 0.375, 0.500, 0.625),
			ExpectedPayouts: generateCoursePayouts(
				[]float64{9.1600000000000001, 2.5, 1.25, 0},
				[]string{"NO", "YES", "YES", "NO"},
			),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			actualPayouts := CalculateCoursePayoutsDBPM(tc.Bets, tc.ProbabilityChanges)

			// Log all actual payouts for debugging
			for i, payout := range actualPayouts {
				t.Logf(
					"%s: Actual payout %d -> {Payout: %.17g, Outcome: %s}",
					tc.Name, i, payout.Payout, payout.Outcome,
				)
			}

			if len(actualPayouts) != len(tc.ExpectedPayouts) {
				t.Fatalf("%s: expected %d payouts, got %d", tc.Name, len(tc.ExpectedPayouts), len(actualPayouts))
			}

			for i, payout := range actualPayouts {
				expected := tc.ExpectedPayouts[i]

				// Log both expected and actual for each mismatch
				if payout.Payout != expected.Payout || payout.Outcome != expected.Outcome {
					t.Errorf(
						"%s: payout %d mismatch. Expected {Payout: %.17g, Outcome: %s}, got {Payout: %.17g, Outcome: %s}",
						tc.Name, i, expected.Payout, expected.Outcome, payout.Payout, payout.Outcome,
					)
				}
			}
		})
	}

}

func TestCalculateNormalizationFactorsDBPM(t *testing.T) {
	testcases := []struct {
		Name          string
		CoursePayouts []CourseBetPayout
		S_YES         int64
		S_NO          int64
		ExpectedF_YES float64
		ExpectedF_NO  float64
	}{
		{
			Name:          "InitialMarketState",
			CoursePayouts: nil,
			S_YES:         0,
			S_NO:          0,
			ExpectedF_YES: 0,
			ExpectedF_NO:  0,
		},
		{
			Name: "FirstBetNoDirection",
			CoursePayouts: generateCoursePayouts(
				[]float64{0}, // Payout = |0.167 - 0.167| * 20
				[]string{"NO"},
			),
			S_YES:         3,
			S_NO:          17,
			ExpectedF_YES: 0,
			ExpectedF_NO:  0,
		},
		{
			Name: "SecondBetYesDirection",
			CoursePayouts: generateCoursePayouts(
				[]float64{4.1600000000000001, 0},
				[]string{"NO", "YES"},
			),
			S_YES:         11,
			S_NO:          19,
			ExpectedF_YES: 0,
			ExpectedF_NO:  4.5673076923076925,
		},
		{
			Name: "ThirdBetYesDirection",
			CoursePayouts: generateCoursePayouts(
				[]float64{6.6599999999999993, 1.25, 0},
				[]string{"NO", "YES", "YES"},
			),
			S_YES:         20,
			S_NO:          20,
			ExpectedF_YES: 16,
			ExpectedF_NO:  3.0030030030030033,
		},
		{
			Name: "FourthBetNegativeNoDirection",
			CoursePayouts: generateCoursePayouts(
				[]float64{9.1600000000000001, 2.5, 1.25, 0},
				[]string{"NO", "YES", "YES", "NO"},
			),
			S_YES:         19,
			S_NO:          11,
			ExpectedF_YES: 5.066666666666666,
			ExpectedF_NO:  1.2008733624454149,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			actualF_YES, actualF_NO := CalculateNormalizationFactorsDBPM(tc.S_YES, tc.S_NO, tc.CoursePayouts)

			// Log the results for manual verification
			t.Logf("%s: F_YES=%f, F_NO=%f", tc.Name, actualF_YES, actualF_NO)

			if actualF_YES != tc.ExpectedF_YES || actualF_NO != tc.ExpectedF_NO {
				t.Errorf(
					"%s: expected F_YES=%f, F_NO=%f; got F_YES=%f, F_NO=%f",
					tc.Name, tc.ExpectedF_YES, tc.ExpectedF_NO, actualF_YES, actualF_NO,
				)
			}
		})
	}

}

func TestCalculateScaledPayoutsDBPM(t *testing.T) {
	testcases := []struct {
		Name          string
		Bets          []models.Bet
		CoursePayouts []CourseBetPayout
		F_YES         float64
		F_NO          float64
		ScaledPayouts []int64
	}{
		{
			Name: "PreventSimultaneousSharesHeld",
			Bets: []models.Bet{
				{
					Amount:   3,
					Outcome:  "YES",
					Username: "user1",
					PlacedAt: time.Date(2024, 5, 18, 5, 7, 31, 428975000, time.UTC),
					MarketID: 3,
				},
				{
					Amount:   1,
					Outcome:  "NO",
					Username: "user1",
					PlacedAt: time.Date(2024, 5, 18, 5, 8, 13, 922665000, time.UTC),
					MarketID: 3,
				},
			},
			CoursePayouts: []CourseBetPayout{
				{Payout: 0.5999999999999999, Outcome: "YES"},
				{Payout: 0.17500000000000004, Outcome: "NO"},
			},
			F_YES:         5.000000000000001, // Actual output from function
			F_NO:          5.714285714285713, // Actual output from function
			ScaledPayouts: []int64{3, 1},
		},
		{
			Name: "InfinityAvoidance",
			Bets: []models.Bet{
				{
					Amount:   1,
					Outcome:  "YES",
					Username: "user2",
					PlacedAt: now,
					MarketID: 1,
				},
				{
					Amount:   -1,
					Outcome:  "YES",
					Username: "user2",
					PlacedAt: now.Add(time.Minute),
					MarketID: 1,
				},
				{
					Amount:   1,
					Outcome:  "NO",
					Username: "user1",
					PlacedAt: now.Add(2 * time.Minute),
					MarketID: 1,
				},
				{
					Amount:   -1,
					Outcome:  "NO",
					Username: "user1",
					PlacedAt: now.Add(3 * time.Minute),
					MarketID: 1,
				},
				{
					Amount:   1,
					Outcome:  "NO",
					Username: "user1",
					PlacedAt: now.Add(4 * time.Minute),
					MarketID: 1,
				},
			},
			CoursePayouts: []CourseBetPayout{
				{Payout: 0.25, Outcome: "YES"},
				{Payout: -0.5, Outcome: "YES"},
				{Payout: 0.25, Outcome: "NO"},
				{Payout: -0, Outcome: "NO"}, // golang math.Round() rounds to -0 and +0
				{Payout: 0.25, Outcome: "NO"},
			},
			F_YES:         0,
			F_NO:          2,
			ScaledPayouts: []int64{0, 0, 1, 0, 1},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			actualPayouts := CalculateScaledPayoutsDBPM(tc.Bets, tc.CoursePayouts, tc.F_YES, tc.F_NO)
			if len(actualPayouts) != len(tc.ScaledPayouts) {
				t.Fatalf("Test %s failed: expected %d payouts, got %d", tc.Name, len(tc.ScaledPayouts), len(actualPayouts))
			}
			for i, payout := range actualPayouts {
				if payout != tc.ScaledPayouts[i] {
					t.Errorf("Test %s failed at index %d: expected payout %d, got %d", tc.Name, i, tc.ScaledPayouts[i], payout)
				}
			}
		})
	}
}

func TestCalculateExcess(t *testing.T) {
	bets := []models.Bet{
		{Amount: 10},
		{Amount: 20},
	}
	scaledPayouts := []int64{5, 15, 10}

	expectedExcess := int64(0)
	actualExcess := calculateExcess(bets, scaledPayouts)

	if actualExcess != expectedExcess {
		t.Errorf("Expected excess: %d, got: %d", expectedExcess, actualExcess)
	}
}

func TestAdjustForPositiveExcess(t *testing.T) {
	scaledPayouts := []int64{10, 20, 30}
	excess := int64(5)

	expectedPayouts := []int64{9, 18, 28}
	actualPayouts := adjustForPositiveExcess(scaledPayouts, excess)

	for i, payout := range actualPayouts {
		if payout != expectedPayouts[i] {
			t.Errorf("At index %d, expected payout: %d, got: %d", i, expectedPayouts[i], payout)
		}
	}
}

func TestAdjustForNegativeExcess(t *testing.T) {
	scaledPayouts := []int64{10, 20, 30}
	excess := int64(-5)

	expectedPayouts := []int64{12, 22, 31}
	actualPayouts := adjustForNegativeExcess(scaledPayouts, excess)

	for i, payout := range actualPayouts {
		if payout != expectedPayouts[i] {
			t.Errorf("At index %d, expected payout: %d, got: %d", i, expectedPayouts[i], payout)
		}
	}
}

func TestAdjustPayouts(t *testing.T) {
	bets := []models.Bet{
		{Amount: 10},
		{Amount: 20},
	}
	scaledPayouts := []int64{10, 20, 15}

	expectedPayouts := []int64{5, 15, 10}
	actualPayouts := AdjustPayouts(bets, scaledPayouts)

	for i, payout := range actualPayouts {
		if payout != expectedPayouts[i] {
			t.Errorf("At index %d, expected payout: %d, got: %d", i, expectedPayouts[i], payout)
		}
	}
}

func TestAggregateUserPayoutsDBPM(t *testing.T) {
	testcases := []struct {
		Name                  string
		Bets                  []models.Bet
		AdjustedScaledPayouts []int64
		AggregatedPositions   []MarketPosition
	}{
		{
			Name: "PreventSimultaneousSharesHeld",
			Bets: []models.Bet{
				{
					Amount:   3,
					Outcome:  "YES",
					Username: "user1",
					PlacedAt: time.Date(2024, 5, 18, 5, 7, 31, 428975000, time.UTC),
					MarketID: 3,
				},
				{
					Amount:   1,
					Outcome:  "NO",
					Username: "user1",
					PlacedAt: time.Date(2024, 5, 18, 5, 8, 13, 922665000, time.UTC),
					MarketID: 3,
				},
			},
			AdjustedScaledPayouts: []int64{3, 1},
			AggregatedPositions: []MarketPosition{
				{Username: "user1", YesSharesOwned: 3, NoSharesOwned: 1},
			},
		},
		{
			Name: "InfinityAvoidance",
			Bets: []models.Bet{
				{
					Amount:   1,
					Outcome:  "YES",
					Username: "user2",
					PlacedAt: now,
					MarketID: 1,
				},
				{
					Amount:   -1,
					Outcome:  "YES",
					Username: "user2",
					PlacedAt: now.Add(time.Minute),
					MarketID: 1,
				},
				{
					Amount:   1,
					Outcome:  "NO",
					Username: "user1",
					PlacedAt: now.Add(2 * time.Minute),
					MarketID: 1,
				},
				{
					Amount:   -1,
					Outcome:  "NO",
					Username: "user1",
					PlacedAt: now.Add(3 * time.Minute),
					MarketID: 1,
				},
				{
					Amount:   1,
					Outcome:  "NO",
					Username: "user1",
					PlacedAt: now.Add(4 * time.Minute),
					MarketID: 1,
				},
			},
			AdjustedScaledPayouts: []int64{0, 0, 1, 0, 0},
			AggregatedPositions: []MarketPosition{
				{Username: "user1", YesSharesOwned: 0, NoSharesOwned: 1},
				{Username: "user2", YesSharesOwned: 0, NoSharesOwned: 0},
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			result := AggregateUserPayoutsDBPM(tc.Bets, tc.AdjustedScaledPayouts)

			// Create a map for expected results for easy lookup
			expectedResults := make(map[string]MarketPosition)
			for _, pos := range tc.AggregatedPositions {
				expectedResults[pos.Username] = pos
			}

			// Create a map from the results for comparison
			resultsMap := make(map[string]MarketPosition)
			for _, pos := range result {
				resultsMap[pos.Username] = pos
			}

			// Check if the results match expected results for each user
			for username, expectedPos := range expectedResults {
				resultPos, ok := resultsMap[username]
				if !ok {
					t.Errorf("Test %s failed: missing position for username %s", tc.Name, username)
					continue
				}
				if resultPos.YesSharesOwned != expectedPos.YesSharesOwned || resultPos.NoSharesOwned != expectedPos.NoSharesOwned {
					t.Errorf("Test %s failed for %s: expected %+v, got %+v", tc.Name, username, expectedPos, resultPos)
				}
			}

			// Check for any unexpected extra users in the results
			for username := range resultsMap {
				if _, ok := expectedResults[username]; !ok {
					t.Errorf("Test %s failed: unexpected position for username %s", tc.Name, username)
				}
			}
		})
	}
}

func TestNetAggregateMarketPositions(t *testing.T) {
	testcases := []struct {
		Name                string
		AggregatedPositions []MarketPosition
		NetPositions        []MarketPosition
	}{
		{
			Name: "PreventSimultaneousSharesHeldPreventSimultaneousSharesHeldPreventSimultaneousSharesHeld",
			AggregatedPositions: []MarketPosition{
				{Username: "user1", YesSharesOwned: 3, NoSharesOwned: 1},
			},
			NetPositions: []MarketPosition{
				{Username: "user1", YesSharesOwned: 2, NoSharesOwned: 0},
			},
		},
		{
			Name: "InfinityAvoidance",
			AggregatedPositions: []MarketPosition{
				{Username: "user1", YesSharesOwned: 0, NoSharesOwned: 1},
				{Username: "user2", YesSharesOwned: 0, NoSharesOwned: 0},
			},
			NetPositions: []MarketPosition{
				{Username: "user1", YesSharesOwned: 0, NoSharesOwned: 1},
				{Username: "user2", YesSharesOwned: 0, NoSharesOwned: 0},
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			actual := NetAggregateMarketPositions(tc.AggregatedPositions)
			expected := tc.NetPositions

			if !reflect.DeepEqual(actual, expected) {
				t.Errorf("Failed %s: expected %+v, got %+v", tc.Name, expected, actual)
			}
		})
	}
}

func TestSingleShareYesNoAllocator(t *testing.T) {
	tests := []struct {
		name     string
		bets     []models.Bet
		expected string
	}{
		{
			name: "Positive YES outcome",
			bets: []models.Bet{
				{Amount: 3, Outcome: "YES"},
				{Amount: 1, Outcome: "NO"},
			},
			expected: "YES",
		},
		{
			name: "Negative NO outcome",
			bets: []models.Bet{
				{Amount: 1, Outcome: "YES"},
				{Amount: 3, Outcome: "NO"},
			},
			expected: "NO",
		},
		{
			name: "No outcome",
			bets: []models.Bet{
				{Amount: 2, Outcome: "YES"},
				{Amount: 2, Outcome: "NO"},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SingleShareYesNoAllocator(tt.bets)
			if result != tt.expected {
				t.Errorf("got %v, expected %v", result, tt.expected)
			}
		})
	}
}
