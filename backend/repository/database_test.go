package repository_test

import (
	"fmt"
	"socialpredict/models"
	"socialpredict/repository"

	"gorm.io/gorm"
)

type Condition struct {
	Query string
	Args  []interface{}
}

type MockDatabase struct {
	model        interface{}
	users        []models.User
	markets      []models.Market
	bets         []models.Bet
	conditions   []Condition
	table        string
	selectFields string
	group        string
	joins        string
	err          error
}

// Clone creates a copy of the MockDatabase, preserving the state.
// This simulates GORM's method chaining behavior.
func (m *MockDatabase) clone() *MockDatabase {
	newConditions := make([]Condition, len(m.conditions))
	copy(newConditions, m.conditions)
	return &MockDatabase{
		users:      m.users,
		markets:    m.markets,
		conditions: newConditions,
		err:        m.err,
	}
}

func (m *MockDatabase) Model(value interface{}) repository.Database {
	clone := m.clone()
	clone.model = value
	return clone
}

func (m *MockDatabase) First(dest interface{}, conds ...interface{}) repository.Database {
	if m.err != nil {
		return m
	}
	// Clone the database to simulate method chaining
	clone := m.clone()
	defer func() {
		clone.conditions = nil
	}()
	switch dest := dest.(type) {
	case *models.User:
		users := clone.users
		// Handle conds (e.g., primary key or username)
		if len(conds) > 0 {
			switch v := conds[0].(type) {
			case string:
				for _, user := range users {
					if user.Username == v {
						*dest = user
						return clone
					}
				}
			default:
				clone.err = fmt.Errorf("invalid argument type")
				return clone
			}
			clone.err = gorm.ErrRecordNotFound
			return clone
		}
		// Apply conditions from Where
		for _, condition := range clone.conditions {
			var filteredUsers []models.User
			switch condition.Query {
			case "username = ?":
				for _, user := range users {
					if user.Username == condition.Args[0] {
						filteredUsers = append(filteredUsers, user)
					}
				}
			default:
				clone.err = fmt.Errorf("unsupported query: %s", condition.Query)
				return clone
			}
			users = filteredUsers
		}
		if len(users) > 0 {
			*dest = users[0]
			return clone
		}
		clone.err = gorm.ErrRecordNotFound
		return clone
	case *models.Market:
		markets := clone.markets
		// Handle conds (e.g., ID)
		if len(conds) > 0 {
			id, ok := conds[0].(int64)
			if ok {
				for _, market := range markets {
					if market.ID == id {
						*dest = market
						return clone
					}
				}
				clone.err = gorm.ErrRecordNotFound
				return clone
			}
			clone.err = fmt.Errorf("invalid argument type")
			return clone
		}
		// Apply conditions from Where (if needed)
		// For simplicity, we'll assume no conditions for markets
		if len(markets) > 0 {
			*dest = markets[0]
			return clone
		}
		clone.err = gorm.ErrRecordNotFound
		return clone
	default:
		clone.err = fmt.Errorf("unsupported type in First")
		return clone
	}
}

func (m *MockDatabase) Preload(query string, args ...interface{}) repository.Database {
	// Clone the database to simulate method chaining
	clone := m.clone()
	// For the mock, we can ignore the actual preload logic
	return clone
}

func (m *MockDatabase) Find(dest interface{}, conds ...interface{}) repository.Database {
	if m.err != nil {
		return m
	}
	// Clone the database to simulate method chaining
	clone := m.clone()
	defer func() {
		clone.conditions = nil
	}()
	switch dest := dest.(type) {
	case *[]models.User:
		users := clone.users
		// Apply conditions from Where
		for _, condition := range clone.conditions {
			var filteredUsers []models.User
			switch condition.Query {
			case "username = ?":
				for _, user := range users {
					if user.Username == condition.Args[0] {
						filteredUsers = append(filteredUsers, user)
					}
				}
			default:
				clone.err = fmt.Errorf("unsupported query: %s", condition.Query)
				return clone
			}
			users = filteredUsers
		}
		*dest = users
		return clone
	case *[]models.Market:
		markets := clone.markets
		// Apply conditions from Where (if needed)
		// For simplicity, we'll assume no conditions for markets
		*dest = markets
		return clone
	default:
		clone.err = fmt.Errorf("unsupported type in Find")
		return clone
	}
}

func (m *MockDatabase) Where(query interface{}, args ...interface{}) repository.Database {
	if m.err != nil {
		return m
	}
	// Clone the database to simulate method chaining
	clone := m.clone()
	if clone.conditions == nil {
		clone.conditions = []Condition{}
	}
	switch q := query.(type) {
	case string:
		clone.conditions = append(clone.conditions, Condition{Query: q, Args: args})
	default:
		clone.err = fmt.Errorf("unsupported query type in Where")
	}
	return clone
}

func (m *MockDatabase) Count(count *int64) repository.Database {
	if m.err != nil {
		return m
	}
	// Determine which dataset to count based on the last operation
	// For simplicity, let's assume we're counting users
	var dataToCount int64
	switch {
	case len(m.users) > 0:
		dataToCount = int64(len(m.users))
	case len(m.markets) > 0:
		dataToCount = int64(len(m.markets))
	default:
		dataToCount = 0
	}
	*count = dataToCount
	return m
}

func (m *MockDatabase) Table(name string) repository.Database {
	clone := m.clone()
	clone.table = name
	return clone
}

func (m *MockDatabase) Select(query string, args ...interface{}) repository.Database {
	clone := m.clone()
	clone.selectFields = query
	return clone
}

func (m *MockDatabase) Group(query string) repository.Database {
	clone := m.clone()
	clone.group = query
	return clone
}

func (m *MockDatabase) Joins(query string, args ...interface{}) repository.Database {
	clone := m.clone()
	clone.joins = query
	return clone
}

func (m *MockDatabase) Scan(dest interface{}) repository.Database {
	clone := m.clone()

	if clone.err != nil {
		return clone
	}

	// Simply assign a mock result, simulating a database query result.
	switch d := dest.(type) {
	case *int64:
		*d = int64(4)                               // Example mock value, based on expected result
		fmt.Println("Mock Scan result set to:", *d) // Debugging output
	default:
		clone.err = fmt.Errorf("unsupported type in Scan")
	}

	return clone
}

func (m *MockDatabase) SubQuery() *gorm.DB {
	// Since this is a mock, we'll simulate the SubQuery behavior by returning a gorm.DB placeholder.
	// You could use a real GORM DB instance connected to an in-memory database or simulate it.
	return &gorm.DB{}
}

func (m *MockDatabase) Raw(sql string, values ...interface{}) repository.Database {
	clone := m.clone()
	// Since this is a mock, we can simulate the result based on the query
	// You can add conditionals to simulate different behaviors based on the SQL string
	if sql == "SELECT SUM(total_unique_users) AS total_first_time_bets FROM (?) AS subquery" {
		// Assuming a mock response for the total first-time bets
		clone.err = nil
		// You can set specific mock data here or store it in m.bets
	}
	return clone
}

func (m *MockDatabase) Error() error {
	return m.err
}
