package modelstesting

import (
	"socialpredict/models"
	"testing"
)

func TestNewFakeDB(t *testing.T) {
	db := NewFakeDB(t)
	if db == nil {
		t.Error("Failed to create fake db")
	}
	tests := []struct {
		Name     string
		Table    interface{}
		WantRows int
	}{
		{
			Name:     "QueryUsers0",
			Table:    &models.User{},
			WantRows: 0,
		},
		{
			Name:     "QueryMarkets0",
			Table:    &models.Market{},
			WantRows: 0,
		},
		{
			Name:     "QueryBets0",
			Table:    &models.Bet{},
			WantRows: 0,
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			result := db.Find(test.Table)
			if int(result.RowsAffected) != test.WantRows {
				t.Fail()
			}
		})
	}
}
