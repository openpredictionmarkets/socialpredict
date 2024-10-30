package betutils

import (
	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"testing"
	"time"
)

func TestCheckMarketStatus(t *testing.T) {

	db := modelstesting.NewFakeDB(t)

	resolvedMarket := models.Market{
		ID:                 1,
		IsResolved:         true,
		ResolutionDateTime: time.Now().Add(-time.Hour),
	}
	closedMarket := models.Market{
		ID:                 2,
		IsResolved:         false,
		ResolutionDateTime: time.Now().Add(-time.Hour),
	}
	openMarket := models.Market{
		ID:                 3,
		IsResolved:         false,
		ResolutionDateTime: time.Now().Add(time.Hour),
	}

	db.Create(&resolvedMarket)
	db.Create(&closedMarket)
	db.Create(&openMarket)

	tests := []struct {
		name       string
		marketID   uint
		expectsErr bool
		errMsg     string
	}{
		{
			name:       "Market not found",
			marketID:   999,
			expectsErr: true,
			errMsg:     "market not found",
		},
		{
			name:       "Resolved market",
			marketID:   1,
			expectsErr: true,
			errMsg:     "cannot place a bet on a resolved market",
		},
		{
			name:       "Closed market",
			marketID:   2,
			expectsErr: true,
			errMsg:     "cannot place a bet on a closed market",
		},
		{
			name:       "Open market",
			marketID:   3,
			expectsErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckMarketStatus(db, tt.marketID)
			if (err != nil) != tt.expectsErr {
				t.Errorf("got error = %v, expected error = %v", err, tt.expectsErr)
			}
			if err != nil && tt.expectsErr && err.Error() != tt.errMsg {
				t.Errorf("expected error message %v, got %v", tt.errMsg, err.Error())
			}
		})
	}
}
