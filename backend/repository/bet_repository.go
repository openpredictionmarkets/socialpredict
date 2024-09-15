package repository

type BetsRepository struct {
	db Database
}

func NewBetsRepository(db Database) *BetsRepository {
	return &BetsRepository{db: db}
}

func (repo *BetsRepository) FirstTimeBets() (int64, error) {
	var totalFirstTimeBetFees int64

	// Subquery to select the first bet for each user, ordered by bet_time ASC
	subquery := repo.db.Table("bets").
		Select("MIN(bet_time) AS first_bet_time, user_id").
		Group("user_id")

	// Perform the final query to sum the fees of the first bets
	result := repo.db.Table("bets").
		Select("SUM(fee) as total_first_time_bet_fees").
		Joins("JOIN (?) AS first_bets ON bets.user_id = first_bets.user_id AND bets.bet_time = first_bets.first_bet_time", subquery).
		Scan(&totalFirstTimeBetFees)

	if result.Error() != nil {
		return 0, result.Error()
	}

	return totalFirstTimeBetFees, nil
}
