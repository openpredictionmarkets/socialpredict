package positions

import (
	"socialpredict/models/modelstesting"
	"socialpredict/setup"
	"socialpredict/setup/setuptesting"
	"testing"

	"gorm.io/gorm"
)

func TestCalculateMarketPositions_WPAM_DBPM(t *testing.T) {
	tests := []struct {
		Name string
		MCL  setup.MarketCreationLoader
		DB   *gorm.DB
	}{
		{
			Name: "",
			MCL:  setuptesting.BuildInitialMarketAppConfig(t, .5, 0, 0, 0),
			DB:   modelstesting.NewFakeDB(t),
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			t.Skip()
		})
	}
}
