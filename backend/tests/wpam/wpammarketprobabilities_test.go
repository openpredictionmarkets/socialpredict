package test

import (
	"socialpredict/handlers/math/probabilities/wpam"
	"socialpredict/models"
	"testing"
	"time"
)

func TestCalculateMarketProbabilitiesWPAM(t *testing.T) {
	// Define the test case
	bets := []models.Bet{
		{Amount: 1, Outcome: "YES"},
		{Amount: -1, Outcome: "YES"},
		{Amount: 1, Outcome: "NO"},
		{Amount: -1, Outcome: "NO"},
		{Amount: 1, Outcome: "NO"},
		{Amount: 1, Outcome: "NO"},
		{Amount: -1, Outcome: "NO"},
		{Amount: -1, Outcome: "NO"},
		{Amount: 500, Outcome: "YES"},
		{Amount: 500, Outcome: "NO"},
		{Amount: 500, Outcome: "NO"},
		{Amount: 500, Outcome: "YES"},
	}
	marketCreatedAt := time.Now()

	// Call the function under test
	probChanges := wpam.CalculateMarketProbabilitiesWPAM(marketCreatedAt, bets)

	// Print results (or you can add assertions here to automatically check expected results)
	for _, pc := range probChanges {
		t.Logf("At %v, Probability: %f", pc.Timestamp, pc.Probability)
	}
}
