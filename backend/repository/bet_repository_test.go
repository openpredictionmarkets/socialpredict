package repository_test

import (
	"socialpredict/models"
	"socialpredict/repository"
	"testing"
)

func TestFirstTimeBets(t *testing.T) {

	mockBets := []models.Bet{
		{Username: "user1", MarketID: 1, Amount: 1},
		{Username: "user1", MarketID: 1, Amount: 1}, // Duplicate, should not be counted
		{Username: "user2", MarketID: 1, Amount: 1},
		{Username: "user2", MarketID: 2, Amount: 1},
		{Username: "user1", MarketID: 2, Amount: 1},
	}

	// Initialize the mock database
	mockDB := &MockDatabase{
		bets: mockBets,
		err:  nil,
	}

	// Initialize the bets repository with the mock database
	betsRepo := repository.NewBetsRepository(mockDB)

	// Call the FirstTimeBets function, which will internally call the Raw() method
	totalBets, err := betsRepo.FirstTimeBets()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check if the result matches the expected mock result
	expectedBets := int64(4) // 4 unique user/market combinations
	if totalBets != expectedBets {
		t.Errorf("Expected total first-time bets %d, got %d", expectedBets, totalBets)
	}
}
