package repository

import "socialpredict/models"

type BetsRepository struct {
	db Database
}

func NewBetsRepository(db Database) *BetsRepository {
	return &BetsRepository{db: db}
}

func (repo *BetsRepository) FirstTimeBets() (int64, error) {

	var bets []models.Bet

	result := repo.db.Model(&models.Bet{}).Select("market_id", "username").Group("market_id").Find(&bets)
	if err := result.Error(); err != nil {
		return 0, err
	}

	var usersByMarket map[uint][]string

	for _, bet := range bets {
		usersByMarket[bet.MarketID] = append(usersByMarket[bet.MarketID], bet.Username)
	}

	var totalFirstTimeBets int64
	for _, users := range usersByMarket {
		totalFirstTimeBets += int64(len(users))
	}

	return totalFirstTimeBets, nil

}
