package betutils

import (
	"errors"
	"socialpredict/models"

	"gorm.io/gorm"
)

func init() {

}

// CheckMarketStatus checks if the market is resolved or closed.
// It returns an error if the market is not suitable for placing a bet.
func CheckMarketStatus(db *gorm.DB, marketID uint) error {
	var market models.Market
	if result := db.First(&market, marketID); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("market not found")
		}
		return errors.New("error fetching market")
	}

	if market.IsClosed() {
		return errors.New("cannot place an order on a closed market")
	}

	return nil
}
