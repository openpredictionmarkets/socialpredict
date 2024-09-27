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

	// Execute the raw SQL query
	result := repo.db.Raw(`
		SELECT market_id, username 
		FROM bets 
		GROUP BY market_id, username
	`).Scan(&bets)

	if result.Error() != nil {
		return 0, result.Error()
	}

	// Initialize the usersByMarket map
	usersByMarket := make(map[string][]string)

	// Build the usersByMarket map
	for _, bet := range bets {
		if usersByMarket[bet.MarketID] == nil {
			usersByMarket[bet.MarketID] = []string{} // Initialize the slice for each market_id
		}
		usersByMarket[bet.MarketID] = append(usersByMarket[bet.MarketID], bet.Username)
	}

	// Count total first-time bets (users across markets)
	var totalFirstTimeBets int64
	for _, users := range usersByMarket {
		totalFirstTimeBets += int64(len(users)) // Count all unique user-market pairs
	}

	return totalFirstTimeBets, nil
}
