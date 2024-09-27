package repository_test

import (
	"reflect"
	"socialpredict/models"
	"socialpredict/repository"
	"testing"
)

func TestGetAllMarkets(t *testing.T) {

	mockMarkets := []models.Market{
		{ID: 1, QuestionTitle: "Market 1"},
		{ID: 2, QuestionTitle: "Market 2"},
	}

	mockDB := &MockDatabase{
		markets: mockMarkets,
		err:     nil,
	}

	// Use the NewMarketRepository constructor to create the repository
	marketRepo := repository.NewMarketRepository(mockDB)

	markets, err := marketRepo.GetAllMarkets()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !reflect.DeepEqual(markets, mockMarkets) {
		t.Errorf("Expected markets %+v, got %+v", mockMarkets, markets)
	}
}

func TestGetMarketByID(t *testing.T) {
	mockMarkets := []models.Market{
		{ID: 1, QuestionTitle: "Market 1"},
		{ID: 2, QuestionTitle: "Market 2"},
	}

	mockDB := &MockDatabase{
		markets: mockMarkets,
		err:     nil,
	}

	// Use the NewMarketRepository constructor to create the repository
	marketRepo := repository.NewMarketRepository(mockDB)

	market, err := marketRepo.GetMarketByID(1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedMarket := &models.Market{ID: 1, QuestionTitle: "Market 1"}
	if !reflect.DeepEqual(market, expectedMarket) {
		t.Errorf("Expected market %+v, got %+v", expectedMarket, market)
	}
}

func TestCountMarkets(t *testing.T) {
	mockMarkets := []models.Market{
		{ID: 1, QuestionTitle: "Market 1"},
		{ID: 2, QuestionTitle: "Market 2"},
	}

	mockDB := &MockDatabase{
		markets: mockMarkets,
		err:     nil,
	}

	marketRepo := repository.NewMarketRepository(mockDB)
	count, err := marketRepo.CountMarkets()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedCount := int64(len(mockMarkets))
	if count != expectedCount {
		t.Errorf("Expected count %d, got %d", expectedCount, count)
	}
}
