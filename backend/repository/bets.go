package repository

type BetsRepository struct {
	db Database
}

func NewBetsRepository(db Database) *BetsRepository {
	return &BetsRepository{db: db}
}

func (repo *BetsRepository) FirstTimeBets() (int64, error) {
	// Define a struct to hold the query result
	var bets []struct {
		MarketID string
		Username string
	}

	// Execute raw SQL query to group by both market_id and username
	// gorm GROUP does not allow two inputs, so this has to be a raw function
	result := repo.db.Raw(`
		SELECT market_id, username 
		FROM bets 
		GROUP BY market_id, username
	`).Scan(&bets)

	if result.Error() != nil {
		return 0, result.Error()
	}

	usersByMarket := make(map[string][]string)

	// create map of market_id's and users for all bets
	for _, bet := range bets {
		if usersByMarket[bet.MarketID] == nil {
			// initialize the slice
			usersByMarket[bet.MarketID] = []string{}
		}
		usersByMarket[bet.MarketID] = append(usersByMarket[bet.MarketID], bet.Username)
	}

	// Count total first-time bets (users across markets)
	var totalFirstTimeBets int64
	for _, users := range usersByMarket {
		totalFirstTimeBets += int64(len(users))
	}

	return totalFirstTimeBets, nil
}
