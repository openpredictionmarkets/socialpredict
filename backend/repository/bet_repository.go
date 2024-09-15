package repository

type BetsRepository struct {
	db Database
}

func NewBetsRepository(db Database) *BetsRepository {
	return &BetsRepository{db: db}
}

func (repo *BetsRepository) FirstTimeBets() (int64, error) {
	var totalFirstTimeBets int64

	// Subquery for counting distinct users per market
	subquery := repo.db.Table("bets").
		Select("COUNT(DISTINCT username) AS total_unique_users").
		Group("market_id").SubQuery()

	// Use Raw() to execute the outer query with the subquery
	result := repo.db.Raw("SELECT SUM(total_unique_users) AS total_first_time_bets FROM (?) AS subquery", subquery).
		Scan(&totalFirstTimeBets)

	if result.Error() != nil {
		return 0, result.Error()
	}

	return totalFirstTimeBets, nil
}
