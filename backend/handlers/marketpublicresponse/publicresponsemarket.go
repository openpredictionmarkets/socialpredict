package marketpublicresponse

import (
	"errors"
	"socialpredict/models"
	"time"

	"gorm.io/gorm"
)

type PublicResponseMarket struct {
	ID                      int64     `json:"id"`
	QuestionTitle           string    `json:"questionTitle"`
	Description             string    `json:"description"`
	OutcomeType             string    `json:"outcomeType"`
	ResolutionDateTime      time.Time `json:"resolutionDateTime"`
	FinalResolutionDateTime time.Time `json:"finalResolutionDateTime"`
	UTCOffset               int       `json:"utcOffset"`
	IsResolved              bool      `json:"isResolved"`
	ResolutionResult        string    `json:"resolutionResult"`
	InitialProbability      float64   `json:"initialProbability"`
	CreatorUsername         string    `json:"creatorUsername"`
	CreatedAt               time.Time `json:"createdAt"`
}

// GetPublicResponseMarketByID retrieves a market by its ID using an existing database connection,
// and constructs a PublicResponseMarket.
func GetPublicResponseMarketByID(db *gorm.DB, marketId string) (PublicResponseMarket, error) {
	if db == nil {
		return PublicResponseMarket{}, errors.New("database connection is nil")
	}

	var market models.Market
	result := db.Where("ID = ?", marketId).First(&market)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return PublicResponseMarket{}, result.Error // Market not found
		}
		return PublicResponseMarket{}, result.Error // Error fetching market
	}

	responseMarket := PublicResponseMarket{
		ID:                      market.ID,
		QuestionTitle:           market.QuestionTitle,
		Description:             market.Description,
		OutcomeType:             market.OutcomeType,
		ResolutionDateTime:      market.ResolutionDateTime,
		FinalResolutionDateTime: market.FinalResolutionDateTime,
		UTCOffset:               market.UTCOffset,
		IsResolved:              market.IsResolved,
		ResolutionResult:        market.ResolutionResult,
		InitialProbability:      market.InitialProbability,
		CreatorUsername:         market.CreatorUsername,
		CreatedAt:               market.CreatedAt,
	}

	return responseMarket, nil
}
