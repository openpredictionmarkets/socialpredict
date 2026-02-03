// Package testing provides lightweight test infrastructure for the SocialPredict backend.
package testing

import (
	"context"
	"strconv"
	"testing"
	"time"

	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"socialpredict/setup"

	"gorm.io/gorm"
)

// TestSuite provides a minimal test environment with a database and factories.
type TestSuite struct {
	t      *testing.T
	DB     *gorm.DB
	Config *setup.EconomicConfig

	Users   *UserFactory
	Markets *MarketFactory
	Bets    *BetFactory
}

// NewTestSuite creates a new test suite with a fresh in-memory database.
func NewTestSuite(t *testing.T) *TestSuite {
	t.Helper()

	db := modelstesting.NewFakeDB(t)
	config, _ := modelstesting.UseStandardTestEconomics(t)
	modelstesting.SeedWPAMFromConfig(config)

	ts := &TestSuite{
		t:      t,
		DB:     db,
		Config: config,
	}

	ts.Users = NewUserFactory(db)
	ts.Markets = NewMarketFactory(db)
	ts.Bets = NewBetFactory(db)

	t.Cleanup(func() {
		if sqlDB, _ := db.DB(); sqlDB != nil {
			_ = sqlDB.Close()
		}
	})

	return ts
}

// Context returns a context with a reasonable timeout for test operations.
func (ts *TestSuite) Context() context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	ts.t.Cleanup(cancel)
	return ctx
}

// ClearDatabase removes all data from common tables.
func (ts *TestSuite) ClearDatabase() error {
	tables := []string{"ledger_entries", "bets", "markets", "users"}
	for _, table := range tables {
		if err := ts.DB.Exec("DELETE FROM " + table).Error; err != nil {
			return err
		}
	}
	return nil
}

// UserFactory creates test users with configurable attributes.
type UserFactory struct {
	db      *gorm.DB
	counter int64
	Created []*models.User
}

// NewUserFactory creates a new user factory.
func NewUserFactory(db *gorm.DB) *UserFactory {
	return &UserFactory{db: db}
}

// UserOption is a function that modifies a user before creation.
type UserOption func(*models.User)

// WithUsername sets the username for a user.
func WithUsername(username string) UserOption {
	return func(u *models.User) {
		u.Username = username
	}
}

// WithBalance sets the account balance for a user.
func WithBalance(balance int64) UserOption {
	return func(u *models.User) {
		u.AccountBalance = balance
		u.InitialAccountBalance = balance
	}
}

// WithEmail sets the email for a user.
func WithEmail(email string) UserOption {
	return func(u *models.User) {
		u.Email = email
	}
}

// Create creates a single test user with optional overrides.
func (uf *UserFactory) Create(opts ...UserOption) (*models.User, error) {
	uf.counter++
	now := time.Now().UnixNano()

	user := &models.User{
		PublicUser: models.PublicUser{
			Username:              generateUsername(uf.counter),
			DisplayName:           generateDisplayName(uf.counter),
			UserType:              "regular",
			InitialAccountBalance: 1000,
			AccountBalance:        1000,
		},
		PrivateUser: models.PrivateUser{
			Email:    generateEmail(uf.counter, now),
			APIKey:   generateAPIKey(uf.counter, now),
			Password: "testpassword",
		},
	}

	for _, opt := range opts {
		opt(user)
	}

	if err := uf.db.Create(user).Error; err != nil {
		return nil, err
	}

	uf.Created = append(uf.Created, user)
	return user, nil
}

// CreateMany creates multiple test users.
func (uf *UserFactory) CreateMany(count int, opts ...UserOption) ([]*models.User, error) {
	users := make([]*models.User, count)
	for i := 0; i < count; i++ {
		user, err := uf.Create(opts...)
		if err != nil {
			return nil, err
		}
		users[i] = user
	}
	return users, nil
}

// MarketFactory creates test markets with configurable attributes.
type MarketFactory struct {
	db      *gorm.DB
	counter int64
	Created []*models.Market
}

// NewMarketFactory creates a new market factory.
func NewMarketFactory(db *gorm.DB) *MarketFactory {
	return &MarketFactory{db: db}
}

// MarketOption is a function that modifies a market before creation.
type MarketOption func(*models.Market)

// WithTitle sets the market title.
func WithTitle(title string) MarketOption {
	return func(m *models.Market) {
		m.QuestionTitle = title
	}
}

// WithDescription sets the market description.
func WithDescription(description string) MarketOption {
	return func(m *models.Market) {
		m.Description = description
	}
}

// WithCreator sets the market creator.
func WithCreator(username string) MarketOption {
	return func(m *models.Market) {
		m.CreatorUsername = username
	}
}

// WithResolutionDate sets the market resolution date.
func WithResolutionDate(date time.Time) MarketOption {
	return func(m *models.Market) {
		m.ResolutionDateTime = date
	}
}

// Create creates a single test market with optional overrides.
func (mf *MarketFactory) Create(creatorUsername string, opts ...MarketOption) (*models.Market, error) {
	mf.counter++

	market := &models.Market{
		QuestionTitle:      generateMarketTitle(mf.counter),
		Description:        generateMarketDescription(mf.counter),
		OutcomeType:        "BINARY",
		ResolutionDateTime: time.Now().Add(24 * time.Hour),
		InitialProbability: 0.5,
		CreatorUsername:    creatorUsername,
	}

	for _, opt := range opts {
		opt(market)
	}

	if err := mf.db.Create(market).Error; err != nil {
		return nil, err
	}

	mf.Created = append(mf.Created, market)
	return market, nil
}

// CreateManyForUser creates multiple test markets for a specific user.
func (mf *MarketFactory) CreateManyForUser(count int, username string, opts ...MarketOption) ([]*models.Market, error) {
	markets := make([]*models.Market, count)
	for i := 0; i < count; i++ {
		market, err := mf.Create(username, opts...)
		if err != nil {
			return nil, err
		}
		markets[i] = market
	}
	return markets, nil
}

// BetFactory creates test bets with configurable attributes.
type BetFactory struct {
	db      *gorm.DB
	counter int64
	Created []*models.Bet
}

// NewBetFactory creates a new bet factory.
func NewBetFactory(db *gorm.DB) *BetFactory {
	return &BetFactory{db: db}
}

// BetOption is a function that modifies a bet before creation.
type BetOption func(*models.Bet)

// WithAmount sets the bet amount.
func WithAmount(amount int64) BetOption {
	return func(b *models.Bet) {
		b.Amount = amount
	}
}

// WithOutcome sets the bet outcome (YES/NO).
func WithOutcome(outcome string) BetOption {
	return func(b *models.Bet) {
		b.Outcome = outcome
	}
}

// Create creates a single test bet with optional overrides.
func (bf *BetFactory) Create(marketID uint, username string, opts ...BetOption) (*models.Bet, error) {
	bf.counter++

	bet := &models.Bet{
		MarketID: marketID,
		Username: username,
		Amount:   100,
		Outcome:  "YES",
		PlacedAt: time.Now(),
	}

	for _, opt := range opts {
		opt(bet)
	}

	if err := bf.db.Create(bet).Error; err != nil {
		return nil, err
	}

	bf.Created = append(bf.Created, bet)
	return bet, nil
}

// CreateMany creates multiple test bets.
func (bf *BetFactory) CreateMany(count int, marketID uint, username string, opts ...BetOption) ([]*models.Bet, error) {
	bets := make([]*models.Bet, count)
	for i := 0; i < count; i++ {
		bet, err := bf.Create(marketID, username, opts...)
		if err != nil {
			return nil, err
		}
		bets[i] = bet
	}
	return bets, nil
}

// Helper functions for generating test data.
func generateUsername(counter int64) string {
	return "testuser" + itoa(counter)
}

func generateDisplayName(counter int64) string {
	return "Test User " + itoa(counter)
}

func generateEmail(counter, timestamp int64) string {
	return "testuser" + itoa(counter) + "_" + itoa(timestamp) + "@test.local"
}

func generateAPIKey(counter, timestamp int64) string {
	return "apikey_" + itoa(counter) + "_" + itoa(timestamp)
}

func generateMarketTitle(counter int64) string {
	return "Test Market " + itoa(counter)
}

func generateMarketDescription(counter int64) string {
	return "This is test market number " + itoa(counter)
}

func itoa(n int64) string {
	return strconv.FormatInt(n, 10)
}
